package marshaler

import (
	"errors"
	"mime"
	"net/http"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"google.golang.org/protobuf/encoding/protojson"
)

func BuildMarshalerRegistry(names ...string) Registry {
	mr := &marshalerRegistry{mimeMap: make(map[string]Marshaler)}
	for _, item := range names {
		_ = mr.add(item, getMarshaller(item))
	}
	return mr
}

var defaultMarshaler = &JSONPb{
	MarshalOptions: protojson.MarshalOptions{
		EmitUnpopulated: true,
	},
	UnmarshalOptions: protojson.UnmarshalOptions{
		DiscardUnknown: true,
	},
}

var (
	acceptHeader      = http.CanonicalHeaderKey("Accept")
	contentTypeHeader = http.CanonicalHeaderKey("Content-Type")
)

type marshalerRegistry struct {
	mimeMap map[string]Marshaler
}

func (mr *marshalerRegistry) GetMarshaler(r *http.Request) (inbound Marshaler, outbound Marshaler) {
	for _, acceptVal := range r.Header[acceptHeader] {
		if m, ok := mr.mimeMap[acceptVal]; ok {
			outbound = m
			break
		}
	}
	for _, contentTypeVal := range r.Header[contentTypeHeader] {
		contentType, _, err := mime.ParseMediaType(contentTypeVal)
		if err != nil {
			logger.Errorf("Failed to parse Content-Type %s: %v", contentTypeVal, err)
			continue
		}
		if m, ok := mr.mimeMap[contentType]; ok {
			inbound = m
			break
		}
	}

	if inbound == nil {
		inbound = defaultMarshaler
	}

	if outbound == nil {
		outbound = inbound
	}

	return inbound, outbound
}

// add adds a marshaler for a case-sensitive MIME type string ("*" to match any
// MIME type).
func (mr *marshalerRegistry) add(mime string, marshaler Marshaler) error {
	if len(mime) == 0 {
		return errors.New("empty MIME type")
	}

	mr.mimeMap[mime] = marshaler

	return nil
}
