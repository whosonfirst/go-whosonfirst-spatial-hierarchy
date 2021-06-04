package filter

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

type SingleSPRResultsFilter struct {
	SPRResultsFilter
}

func init() {
	ctx := context.Background()
	RegisterSPRResultsFilter(ctx, "single", NewSingleSPRResultsFilter)
}

func NewSingleSPRResultsFilter(ctx context.Context, uri string) (SPRResultsFilter, error) {

	f := &SingleSPRResultsFilter{}
	return f, nil
}

func (f *SingleSPRResultsFilter) FilterResults(ctx context.Context, r reader.Reader, body []byte, possible []spr.StandardPlacesResult) (spr.StandardPlacesResult, error) {

	count := len(possible)

	if count != 1 {
		return nil, fmt.Errorf("Invalid result count (%d)", count)
	}

	parent_spr := possible[0]
	return parent_spr, nil
}
