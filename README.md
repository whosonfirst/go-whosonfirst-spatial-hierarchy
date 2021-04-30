# go-whosonfirst-spatial-hierarchy

Opionated Who's On First (WOF) hierarchy for `go-whosonfirst-spatial` packages.

## IMPORTANT

This is work in progress. Documentation to follow.

## Example

```
import (
	"github.com/whosonfirst/go-whosonfirst-spatial-hierarchy"	
	_ "github.com/whosonfirst/go-whosonfirst-spatial-sqlite"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
)

body := []byte(`{"type":"Feature" ...}`)

spatial_db, _ := database.NewSpatialDatabase(ctx, "sqlite://?dsn=/usr/local/data/whosonfirst.db")

resolver, _ := hierarchy.NewPointInPolygonHierarchyResolver(ctx, spatial_db, nil)

inputs := &filter.SPRInputs{}

results_cb := hierarchy.FirstButForgivingSPRResultsFunc
update_cb := hierarchy.DefaultPointInPolygonHierarchyResolverUpdateCallback()
		
new_body, _ := resolver.PointInPolygonAndUpdate(ctx, inputs, results_cb, update_cb, body)
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-spatial
* https://github.com/whosonfirst/go-whosonfirst-spatial-pip
* https://github.com/whosonfirst/go-whosonfirst-exporter
* https://github.com/sfomuseum/go-sfomuseum-mapshaper