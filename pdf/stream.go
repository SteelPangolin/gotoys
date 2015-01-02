package main

import (
	"bytes"
	"compress/zlib"
	"encoding/ascii85"
	"fmt"
	"io"
	"io/ioutil"
)

// PDF stream object. Bytes, but may be compressed or otherwise encoded.
type Stream struct {
	buf   []byte
	attrs PDFMap
}

func (o Stream) String() string {
	return fmt.Sprintf("Stream (%d bytes) %v", len(o.buf), o.attrs)
}

func (o Stream) Val() interface{} {
	// check length
	expected := o.attrs["Length"].Val().(int64)
	actual := int64(len(o.buf))
	if expected != actual {
		panic(fmt.Errorf("Stream: expected length %d bytes, actual length %d bytes", expected, actual))
	}

	filterAttr, filtered := o.attrs["Filter"]
	// fast path for unfiltered streams
	if !filtered {
		return o.buf
	}

	// get a list of filter names
	filterNames := []string{}
	switch filterVal := filterAttr.Val().(type) {
	case string:
		filterNames = append(filterNames, filterVal)
	case PDFList:
		for _, filterValElem := range filterVal {
			filterNames = append(filterNames, filterValElem.Val().(string))
		}
	default:
		panic(fmt.Errorf("Stream: illegal filter: %T %s", filterVal, filterVal))
	}

	// build a filter chain
	var r io.Reader = bytes.NewReader(o.buf)
	for _, filterName := range filterNames {
		switch filterName {
		case "FlateDecode":
			filter, err := zlib.NewReader(r)
			if err != nil {
				panic(err)
			}
			defer func() {
				if err := filter.Close(); err != nil {
					panic(err)
				}
			}()
			r = filter

		case "ASCII85Decode":
			r = ascii85.NewDecoder(r)

		default:
			panic(fmt.Errorf("Unknown filter: %s", filterName))
		}
	}

	// read all data through the filter chain
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return contents
}
