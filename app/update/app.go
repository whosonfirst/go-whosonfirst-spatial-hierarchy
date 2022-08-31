package update

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-sfomuseum-mapshaper"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-spatial-hierarchy"
	hierarchy_filter "github.com/whosonfirst/go-whosonfirst-spatial-hierarchy/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	wof_writer "github.com/whosonfirst/go-whosonfirst-writer/v2"
	"github.com/whosonfirst/go-writer/v2"
	"io"
	"log"
)

type UpdateApplicationOptions struct {
	Writer             writer.Writer
	WriterURI          string
	Exporter           export.Exporter
	ExporterURI        string
	MapshaperServerURI string
	SpatialDatabase    database.SpatialDatabase
	SpatialDatabaseURI string
	ToIterator         string
	FromIterator       string
	SPRFilterInputs    *filter.SPRInputs
	SPRResultsFunc     hierarchy_filter.FilterSPRResultsFunc                   // This one chooses one result among many (or nil)
	PIPUpdateFunc      hierarchy.PointInPolygonHierarchyResolverUpdateCallback // This one constructs a map[string]interface{} to update the target record (or not)
}

type UpdateApplicationPaths struct {
	To   []string
	From []string
}

// UpdateApplication is a
type UpdateApplication struct {
	to                  string
	from                string
	tool                *hierarchy.PointInPolygonHierarchyResolver
	writer              writer.Writer
	exporter            export.Exporter
	spatial_db          database.SpatialDatabase
	sprResultsFunc      hierarchy_filter.FilterSPRResultsFunc
	sprFilterInputs     *filter.SPRInputs
	hierarchyUpdateFunc hierarchy.PointInPolygonHierarchyResolverUpdateCallback
	logger              *log.Logger
}

func Run(ctx context.Context, logger *log.Logger) error {

	fs, err := DefaultFlagSet(ctx)

	if err != nil {
		fmt.Errorf("Failed to create application flag set, %w", err)
	}

	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *log.Logger) error {

	flagset.Parse(fs)

	inputs := &filter.SPRInputs{}

	inputs.IsCurrent = is_current
	inputs.IsCeased = is_ceased
	inputs.IsDeprecated = is_deprecated
	inputs.IsSuperseded = is_superseded
	inputs.IsSuperseding = is_superseding

	opts := &UpdateApplicationOptions{
		WriterURI:          writer_uri,
		ExporterURI:        exporter_uri,
		SpatialDatabaseURI: spatial_database_uri,
		MapshaperServerURI: mapshaper_server,
		SPRResultsFunc:     hierarchy_filter.FirstButForgivingSPRResultsFunc, // sudo make me configurable
		SPRFilterInputs:    inputs,
		ToIterator:         iterator_uri,
		FromIterator:       spatial_iterator_uri,
	}

	hierarchy_paths := fs.Args()

	paths := &UpdateApplicationPaths{
		To:   hierarchy_paths,
		From: spatial_paths,
	}

	var ex export.Exporter
	var wr writer.Writer
	var spatial_db database.SpatialDatabase

	if opts.Exporter != nil {
		ex = opts.Exporter
	} else {

		_ex, err := export.NewExporter(ctx, opts.ExporterURI)

		if err != nil {
			return fmt.Errorf("Failed to create exporter for '%s', %v", opts.ExporterURI, err)
		}

		ex = _ex
	}

	if opts.Writer != nil {
		wr = opts.Writer
	} else {
		_wr, err := writer.NewWriter(ctx, opts.WriterURI)

		if err != nil {
			return fmt.Errorf("Failed to create writer for '%s', %v", opts.WriterURI, err)
		}

		wr = _wr
	}

	if opts.SpatialDatabase != nil {
		spatial_db = opts.SpatialDatabase
	} else {

		_db, err := database.NewSpatialDatabase(ctx, opts.SpatialDatabaseURI)

		if err != nil {
			return fmt.Errorf("Failed to create spatial database for '%s', %v", opts.SpatialDatabaseURI, err)
		}

		spatial_db = _db
	}

	// All of this mapshaper stuff can't be retired/replaced fast enough...
	// (20210222/thisisaaronland)

	var ms_client *mapshaper.Client

	if opts.MapshaperServerURI != "" {

		// Set up mapshaper endpoint (for deriving centroids during PIP operations)
		// Make sure it's working

		client, err := mapshaper.NewClient(ctx, opts.MapshaperServerURI)

		if err != nil {
			return fmt.Errorf("Failed to create mapshaper client for '%s', %v", opts.MapshaperServerURI, err)
		}

		ok, err := client.Ping()

		if err != nil {
			return fmt.Errorf("Failed to ping '%s', %v", opts.MapshaperServerURI, err)
		}

		if !ok {
			return fmt.Errorf("'%s' returned false", opts.MapshaperServerURI)
		}

		ms_client = client
	}

	update_cb := opts.PIPUpdateFunc

	if update_cb == nil {
		update_cb = hierarchy.DefaultPointInPolygonHierarchyResolverUpdateCallback()
	}

	tool, err := hierarchy.NewPointInPolygonHierarchyResolver(ctx, spatial_db, ms_client)

	if err != nil {
		return fmt.Errorf("Failed to create PIP tool, %v", err)
	}

	app := &UpdateApplication{
		to:                  opts.ToIterator,
		from:                opts.FromIterator,
		spatial_db:          spatial_db,
		tool:                tool,
		exporter:            ex,
		writer:              wr,
		sprFilterInputs:     opts.SPRFilterInputs,
		sprResultsFunc:      opts.SPRResultsFunc,
		hierarchyUpdateFunc: update_cb,
		logger:              logger,
	}

	return app.Run(ctx, paths)
}

func (app *UpdateApplication) Run(ctx context.Context, paths *UpdateApplicationPaths) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// These are the data we are indexing to HIERARCHY from

	err := app.IndexSpatialDatabase(ctx, paths.From...)

	if err != nil {
		return err
	}

	// These are the data we are HIERARCHY-ing

	to_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read '%s', %v", path, err)
		}

		_, err = app.UpdateAndPublishFeature(ctx, body)

		if err != nil {
			return fmt.Errorf("Failed to update feature for '%s', %v", path, err)
		}

		return nil
	}

	to_iter, err := iterator.NewIterator(ctx, app.to, to_cb)

	if err != nil {
		return fmt.Errorf("Failed to create new HIERARCHY (to) iterator for input, %v", err)
	}

	err = to_iter.IterateURIs(ctx, paths.To...)

	if err != nil {
		return err
	}

	// This is important for something things like
	// whosonfirst/go-writer-featurecollection
	// (20210219/thisisaaronland)

	return app.writer.Close(ctx)
}

func (app *UpdateApplication) IndexSpatialDatabase(ctx context.Context, uris ...string) error {

	from_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		geom_type, err := geometry.Type(body)

		if err != nil {
			return fmt.Errorf("Failed to derive geometry type for %s, %w", path, err)
		}

		switch geom_type {
		case "Polygon", "MultiPolygon":
			return app.spatial_db.IndexFeature(ctx, body)
		default:
			return nil
		}
	}

	from_iter, err := iterator.NewIterator(ctx, app.from, from_cb)

	if err != nil {
		return fmt.Errorf("Failed to create spatial (from) iterator, %v", err)
	}

	err = from_iter.IterateURIs(ctx, uris...)

	if err != nil {
		return fmt.Errorf("Failed to iteratre URIs, %w", err)
	}

	return nil
}

// UpdateAndPublishFeature will invoke the `PointInPolygonAndUpdate` method using the `hierarchy.PointInPolygonHierarchyResolver` instance
// associated with 'app' using 'body' as its input. If successful and there are changes the result will be published using the `PublishFeature`
// method.
func (app *UpdateApplication) UpdateAndPublishFeature(ctx context.Context, body []byte) ([]byte, error) {

	has_changed, new_body, err := app.UpdateFeature(ctx, body)

	if err != nil {
		return nil, fmt.Errorf("Failed to update feature, %w", err)
	}

	if has_changed {

		new_body, err = app.PublishFeature(ctx, new_body)

		if err != nil {
			return nil, fmt.Errorf("Failed to publish feature, %w", err)
		}
	}

	return new_body, nil
}

// UpdateFeature will invoke the `PointInPolygonAndUpdate` method using the `hierarchy.PointInPolygonHierarchyResolver` instance
// associated with 'app' using 'body' as its input.
func (app *UpdateApplication) UpdateFeature(ctx context.Context, body []byte) (bool, []byte, error) {

	return app.tool.PointInPolygonAndUpdate(ctx, app.sprFilterInputs, app.sprResultsFunc, app.hierarchyUpdateFunc, body)
}

// PublishFeature exports 'body' using the `whosonfirst/go-writer/v2` instance associated with 'app'.
func (app *UpdateApplication) PublishFeature(ctx context.Context, body []byte) ([]byte, error) {

	new_body, err := app.exporter.Export(ctx, body)

	if err != nil {
		return nil, err
	}

	_, err = wof_writer.WriteBytes(ctx, app.writer, new_body)

	if err != nil {
		return nil, err
	}

	return new_body, nil
}
