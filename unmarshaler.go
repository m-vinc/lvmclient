package lvmclient

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"golang.org/x/text/encoding/ianaindex"
)

// some xml output from dbus is not unmarshable for a shady reason
// found a snippet https://dzhg.dev/posts/2020/08/how-to-parse-xml-with-non-utf8-encoding-in-go/
// to parse dbus output
func customUnmarshal[T any](b []byte, out *T) error {
	decoder := xml.NewDecoder(bytes.NewBuffer(b))
	decoder.CharsetReader = func(charset string, reader io.Reader) (io.Reader, error) {
		enc, err := ianaindex.IANA.Encoding("utf-8")
		if err != nil {
			return nil, fmt.Errorf("charset %s: %s", charset, err.Error())
		}
		if enc == nil {
			return reader, nil
		}

		return enc.NewDecoder().Reader(reader), nil
	}

	if err := decoder.Decode(out); err != nil {
		return err
	}
	return nil
}
