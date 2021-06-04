package filter

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"net/url"
	"strconv"
)

type FirstSPRResultsFilter struct {
	SPRResultsFilter
	forgiving bool
}

func init() {
	ctx := context.Background()
	RegisterSPRResultsFilter(ctx, "first", NewFirstSPRResultsFilter)
}

func NewFirstSPRResultsFilter(ctx context.Context, uri string) (SPRResultsFilter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	forgiving := false

	q := u.Query()

	str_forgiving := q.Get("forgiving")

	if str_forgiving != "" {

		f, err := strconv.ParseBool(str_forgiving)

		if err != nil {
			return nil, err
		}

		forgiving = f
	}

	f := &FirstSPRResultsFilter{
		forgiving: forgiving,
	}

	return f, nil
}

func (f *FirstSPRResultsFilter) FilterResults(ctx context.Context, r reader.Reader, body []byte, possible []spr.StandardPlacesResult) (spr.StandardPlacesResult, error) {

	if len(possible) == 0 {

		if f.forgiving {
			return nil, nil
		}

		return nil, fmt.Errorf("No results")
	}

	parent_spr := possible[0]
	return parent_spr, nil
}
