package update

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"net/url"
	"sort"
	"strings"
)

type Updater interface {
	UpdateResults(context.Context, reader.Reader, spr.StandardPlacesResult) (map[string]interface{}, error)
}

type UpdaterInitializeFunc func(ctx context.Context, uri string) (Updater, error)

var spr_updaters roster.Roster

func ensureSpatialRoster() error {

	if spr_updaters == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		spr_updaters = r
	}

	return nil
}

func RegisterUpdater(ctx context.Context, scheme string, f UpdaterInitializeFunc) error {

	err := ensureSpatialRoster()

	if err != nil {
		return err
	}

	return spr_updaters.Register(ctx, scheme, f)
}

func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureSpatialRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range spr_updaters.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

func NewUpdater(ctx context.Context, uri string) (Updater, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := spr_updaters.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	f := i.(UpdaterInitializeFunc)
	return f(ctx, uri)
}
