package export

import (
	"bytes"
	"encoding/json"
	_ "fmt"
	"io"

	"github.com/whosonfirst/go-whosonfirst-export/v2/properties"
	format "github.com/whosonfirst/go-whosonfirst-format"
)

func Export(feature []byte, opts *Options, wr io.Writer) error {

	var err error

	feature, err = prepareWithoutTimestamps(feature, opts)

	if err != nil {
		return err
	}

	feature, err = prepareTimestamps(feature, opts)

	if err != nil {
		return err
	}

	feature, err = Format(feature, opts)

	if err != nil {
		return err
	}

	r := bytes.NewReader(feature)
	_, err = io.Copy(wr, r)

	return err
}

// ExportChanged returns a boolean which indicates whether the file was changed
// by comparing it to the `existingFeature` byte slice, before the lastmodified
// timestamp is incremented. If the `feature` is identical to `existingFeature`
// it doesn't write to the `io.Writer`.

func ExportChanged(feature []byte, existingFeature []byte, opts *Options, wr io.Writer) (changed bool, err error) {

	changed = false

	feature, err = prepareWithoutTimestamps(feature, opts)

	if err != nil {
		return
	}

	feature, err = Format(feature, opts)

	if err != nil {
		return
	}

	changed = !bytes.Equal(feature, existingFeature)

	if !changed {
		return
	}

	feature, err = prepareTimestamps(feature, opts)

	if err != nil {
		return
	}

	feature, err = Format(feature, opts)

	if err != nil {
		return
	}

	r := bytes.NewReader(feature)
	_, err = io.Copy(wr, r)

	return
}

func Prepare(feature []byte, opts *Options) ([]byte, error) {
	var err error

	feature, err = prepareWithoutTimestamps(feature, opts)
	if err != nil {
		return nil, err
	}

	feature, err = prepareTimestamps(feature, opts)
	if err != nil {
		return nil, err
	}

	return feature, nil
}

func Format(feature []byte, opts *Options) ([]byte, error) {
	var f format.Feature
	json.Unmarshal(feature, &f)
	return format.FormatFeature(&f)
}

func prepareWithoutTimestamps(feature []byte, opts *Options) ([]byte, error) {

	var err error

	feature, err = properties.EnsureWOFId(feature, opts.IDProvider)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureRequired(feature)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureEDTF(feature)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureParentId(feature)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureHierarchy(feature)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureBelongsTo(feature)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureSupersedes(feature)

	if err != nil {
		return nil, err
	}

	feature, err = properties.EnsureSupersededBy(feature)

	if err != nil {
		return nil, err
	}

	return feature, nil
}

func prepareTimestamps(feature []byte, opts *Options) ([]byte, error) {
	return properties.EnsureTimestamps(feature)
}
