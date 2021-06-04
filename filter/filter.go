// Package filter defines interfaces for filtering point-in-polygon results.
package filter

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

type SPRResultsFilter interface {
	FilterResults(context.Context, reader.Reader, []byte, []spr.StandardPlacesResult) (spr.StandardPlacesResult, error)
}

type SPRResultsFilterInitializeFunc func(ctx context.Context, uri string) (SPRResultsFilter, error)

var spr_filters roster.Roster

func ensureSpatialRoster() error {

	if spr_filters == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		spr_filters = r
	}

	return nil
}

func RegisterSPRResultsFilter(ctx context.Context, scheme string, f SPRResultsFilterInitializeFunc) error {

	err := ensureSpatialRoster()

	if err != nil {
		return err
	}

	return spr_filters.Register(ctx, scheme, f)
}

func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureSpatialRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range spr_filters.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

func NewSPRResultsFilter(ctx context.Context, uri string) (SPRResultsFilter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := spr_filters.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	f := i.(SPRResultsFilterInitializeFunc)
	return f(ctx, uri)
}
