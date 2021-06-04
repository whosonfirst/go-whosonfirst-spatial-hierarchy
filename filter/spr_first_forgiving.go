package filter

import (
	"context"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

type FirstButForgivingSPRResultsFilter struct {
	SPRResultsFilter
}

func NewFirstButForgivingSPRResultsFilter(ctx context.Context, uri string) (SPRResultsFilter, error) {
	f := &FirstButForgivingSPRResultsFilter{}
	return f, nil
}

func (f *FirstButForgivingSPRResultsFilter) FilterResults(ctx context.Context, r reader.Reader, body []byte, possible []spr.StandardPlacesResult) (spr.StandardPlacesResult, error) {

	if len(possible) == 0 {
		return nil, nil
	}

	parent_spr := possible[0]
	return parent_spr, nil
}
