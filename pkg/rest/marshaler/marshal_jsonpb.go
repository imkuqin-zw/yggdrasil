package marshaler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/rest/convert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func init() {
	RegisterMarshallerBuilder("jsonpb", NewJsonPbMarshaler)
}

func NewJsonPbMarshaler() Marshaler {
	cfg := &JSONPbConfig{}
	err := config.Get(fmt.Sprintf(config.KeyRestMarshalerCfg, "jsonpb")).Scan(&cfg)
	if err != nil {
		logger.Fatalf("fault to load jsonpb marshaler config")
	}
	return &JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			Multiline:       cfg.MarshalOptions.Multiline,
			Indent:          cfg.MarshalOptions.Indent,
			AllowPartial:    cfg.MarshalOptions.AllowPartial,
			UseProtoNames:   cfg.MarshalOptions.UseProtoNames,
			UseEnumNumbers:  cfg.MarshalOptions.UseEnumNumbers,
			EmitUnpopulated: cfg.MarshalOptions.EmitUnpopulated,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			AllowPartial:   cfg.UnmarshalOptions.AllowPartial,
			DiscardUnknown: cfg.UnmarshalOptions.DiscardUnknown,
			RecursionLimit: cfg.UnmarshalOptions.RecursionLimit,
		},
	}
}

type JSONPbConfig struct {
	MarshalOptions struct {
		// Multiline specifies whether the marshaler should format the output in
		// indented-form with every textual element on a new line.
		// If Indent is an empty string, then an arbitrary indent is chosen.
		Multiline bool

		// Indent specifies the set of indentation characters to use in a multiline
		// formatted output such that every entry is preceded by Indent and
		// terminated by a newline. If non-empty, then Multiline is treated as true.
		// Indent can only be composed of space or tab characters.
		Indent string

		// AllowPartial allows messages that have missing required fields to marshal
		// without returning an error. If AllowPartial is false (the default),
		// Marshal will return error if there are any missing required fields.
		AllowPartial bool

		// UseProtoNames uses proto field name instead of lowerCamelCase name in JSON
		// field names.
		UseProtoNames bool

		// UseEnumNumbers emits enum values as numbers.
		UseEnumNumbers bool

		// EmitUnpopulated specifies whether to emit unpopulated fields. It does not
		// emit unpopulated oneof fields or unpopulated extension fields.
		// The JSON value emitted for unpopulated fields are as follows:
		//  ╔═══════╤════════════════════════════╗
		//  ║ JSON  │ Protobuf field             ║
		//  ╠═══════╪════════════════════════════╣
		//  ║ false │ proto3 boolean fields      ║
		//  ║ 0     │ proto3 numeric fields      ║
		//  ║ ""    │ proto3 string/bytes fields ║
		//  ║ null  │ proto2 scalar fields       ║
		//  ║ null  │ message fields             ║
		//  ║ []    │ list fields                ║
		//  ║ {}    │ map fields                 ║
		//  ╚═══════╧════════════════════════════╝
		EmitUnpopulated bool

		// EmitDefaultValues specifies whether to emit default-valued primitive fields,
		// empty lists, and empty maps. The fields affected are as follows:
		//  ╔═══════╤════════════════════════════════════════╗
		//  ║ JSON  │ Protobuf field                         ║
		//  ╠═══════╪════════════════════════════════════════╣
		//  ║ false │ non-optional scalar boolean fields     ║
		//  ║ 0     │ non-optional scalar numeric fields     ║
		//  ║ ""    │ non-optional scalar string/byte fields ║
		//  ║ []    │ empty repeated fields                  ║
		//  ║ {}    │ empty map fields                       ║
		//  ╚═══════╧════════════════════════════════════════╝
		//
		// Behaves similarly to EmitUnpopulated, but does not emit "null"-value fields,
		// i.e. presence-sensing fields that are omitted will remain omitted to preserve
		// presence-sensing.
		// EmitUnpopulated takes precedence over EmitDefaultValues since the former generates
		// a strict superset of the latter.
		EmitDefaultValues bool
	}
	UnmarshalOptions struct {
		// If AllowPartial is set, input for messages that will result in missing
		// required fields will not return an error.
		AllowPartial bool

		// If DiscardUnknown is set, unknown fields and enum name values are ignored.
		DiscardUnknown bool

		// RecursionLimit limits how deeply messages may be nested.
		// If zero, a default limit is applied.
		RecursionLimit int
	}
}

// JSONPb is a Marshaler which marshals/unmarshals into/from JSON
// with the "google.golang.org/protobuf/encoding/protojson" marshaler.
// It supports the full functionality of protobuf unlike JSONBuiltin.
//
// The NewDecoder method returns a DecoderWrapper, so the underlying
// *json.Decoder methods can be used.
type JSONPb struct {
	protojson.MarshalOptions
	protojson.UnmarshalOptions
}

// ContentType always returns "application/json".
func (*JSONPb) ContentType(_ interface{}) string {
	return "application/json"
}

// Marshal marshals "v" into JSON.
func (j *JSONPb) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := j.marshalTo(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (j *JSONPb) marshalTo(w io.Writer, v interface{}) error {
	p, ok := v.(proto.Message)
	if !ok {
		buf, err := j.marshalNonProtoField(v)
		if err != nil {
			return err
		}
		if j.Indent != "" {
			b := &bytes.Buffer{}
			if err := json.Indent(b, buf, "", j.Indent); err != nil {
				return err
			}
			buf = b.Bytes()
		}
		_, err = w.Write(buf)
		return err
	}

	b, err := j.MarshalOptions.Marshal(p)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}

var (
	// protoMessageType is stored to prevent constant lookup of the same type at runtime.
	protoMessageType = reflect.TypeOf((*proto.Message)(nil)).Elem()
)

// marshalNonProto marshals a non-message field of a protobuf message.
// This function does not correctly marshal arbitrary data structures into JSON,
// it is only capable of marshaling non-message field values of protobuf,
// i.e. primitive types, enums; pointers to primitives or enums; maps from
// integer/string types to primitives/enums/pointers to messages.
func (j *JSONPb) marshalNonProtoField(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return []byte("null"), nil
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice {
		if rv.IsNil() {
			if j.EmitUnpopulated {
				return []byte("[]"), nil
			}
			return []byte("null"), nil
		}

		if rv.Type().Elem().Implements(protoMessageType) {
			var buf bytes.Buffer
			if err := buf.WriteByte('['); err != nil {
				return nil, err
			}
			for i := 0; i < rv.Len(); i++ {
				if i != 0 {
					if err := buf.WriteByte(','); err != nil {
						return nil, err
					}
				}
				if err := j.marshalTo(&buf, rv.Index(i).Interface().(proto.Message)); err != nil {
					return nil, err
				}
			}
			if err := buf.WriteByte(']'); err != nil {
				return nil, err
			}

			return buf.Bytes(), nil
		}

		if rv.Type().Elem().Implements(typeProtoEnum) {
			var buf bytes.Buffer
			if err := buf.WriteByte('['); err != nil {
				return nil, err
			}
			for i := 0; i < rv.Len(); i++ {
				if i != 0 {
					if err := buf.WriteByte(','); err != nil {
						return nil, err
					}
				}
				var err error
				if j.UseEnumNumbers {
					_, err = buf.WriteString(strconv.FormatInt(rv.Index(i).Int(), 10))
				} else {
					_, err = buf.WriteString("\"" + rv.Index(i).Interface().(protoEnum).String() + "\"")
				}
				if err != nil {
					return nil, err
				}
			}
			if err := buf.WriteByte(']'); err != nil {
				return nil, err
			}

			return buf.Bytes(), nil
		}
	}

	if rv.Kind() == reflect.Map {
		m := make(map[string]*json.RawMessage)
		for _, k := range rv.MapKeys() {
			buf, err := j.Marshal(rv.MapIndex(k).Interface())
			if err != nil {
				return nil, err
			}
			m[fmt.Sprintf("%v", k.Interface())] = (*json.RawMessage)(&buf)
		}
		return json.Marshal(m)
	}
	if enum, ok := rv.Interface().(protoEnum); ok && !j.UseEnumNumbers {
		return json.Marshal(enum.String())
	}
	return json.Marshal(rv.Interface())
}

// Unmarshal unmarshals JSON "data" into "v"
func (j *JSONPb) Unmarshal(data []byte, v interface{}) error {
	return unmarshalJSONPb(data, j.UnmarshalOptions, v)
}

// NewDecoder returns a Decoder which reads JSON stream from "r".
func (j *JSONPb) NewDecoder(r io.Reader) Decoder {
	d := json.NewDecoder(r)
	return DecoderWrapper{
		Decoder:          d,
		UnmarshalOptions: j.UnmarshalOptions,
	}
}

// DecoderWrapper is a wrapper around a *json.Decoder that adds
// support for protos to the Decode method.
type DecoderWrapper struct {
	*json.Decoder
	protojson.UnmarshalOptions
}

// Decode wraps the embedded decoder's Decode method to support
// protos using a jsonpb.Unmarshaler.
func (d DecoderWrapper) Decode(v interface{}) error {
	return decodeJSONPb(d.Decoder, d.UnmarshalOptions, v)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *JSONPb) NewEncoder(w io.Writer) Encoder {
	return EncoderFunc(func(v interface{}) error {
		if err := j.marshalTo(w, v); err != nil {
			return err
		}
		// mimic json.Encoder by adding a newline (makes output
		// easier to read when it contains multiple encoded items)
		_, err := w.Write(j.Delimiter())
		return err
	})
}

func unmarshalJSONPb(data []byte, unmarshaler protojson.UnmarshalOptions, v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	return decodeJSONPb(d, unmarshaler, v)
}

func decodeJSONPb(d *json.Decoder, unmarshaler protojson.UnmarshalOptions, v interface{}) error {
	p, ok := v.(proto.Message)
	if !ok {
		return decodeNonProtoField(d, unmarshaler, v)
	}

	// Decode into bytes for marshalling
	var b json.RawMessage
	if err := d.Decode(&b); err != nil {
		return err
	}

	return unmarshaler.Unmarshal([]byte(b), p)
}

func decodeNonProtoField(d *json.Decoder, unmarshaler protojson.UnmarshalOptions, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not a pointer", v)
	}
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		if rv.Type().ConvertibleTo(typeProtoMessage) {
			// Decode into bytes for marshalling
			var b json.RawMessage
			if err := d.Decode(&b); err != nil {
				return err
			}

			return unmarshaler.Unmarshal([]byte(b), rv.Interface().(proto.Message))
		}
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Map {
		if rv.IsNil() {
			rv.Set(reflect.MakeMap(rv.Type()))
		}
		conv, ok := convFromType[rv.Type().Key().Kind()]
		if !ok {
			return fmt.Errorf("unsupported type of map field key: %v", rv.Type().Key())
		}

		m := make(map[string]*json.RawMessage)
		if err := d.Decode(&m); err != nil {
			return err
		}
		for k, v := range m {
			result := conv.Call([]reflect.Value{reflect.ValueOf(k)})
			if err := result[1].Interface(); err != nil {
				return err.(error)
			}
			bk := result[0]
			bv := reflect.New(rv.Type().Elem())
			if v == nil {
				null := json.RawMessage("null")
				v = &null
			}
			if err := unmarshalJSONPb([]byte(*v), unmarshaler, bv.Interface()); err != nil {
				return err
			}
			rv.SetMapIndex(bk, bv.Elem())
		}
		return nil
	}
	if rv.Kind() == reflect.Slice {
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			var sl []byte
			if err := d.Decode(&sl); err != nil {
				return err
			}
			if sl != nil {
				rv.SetBytes(sl)
			}
			return nil
		}

		var sl []json.RawMessage
		if err := d.Decode(&sl); err != nil {
			return err
		}
		if sl != nil {
			rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))
		}
		for _, item := range sl {
			bv := reflect.New(rv.Type().Elem())
			if err := unmarshalJSONPb([]byte(item), unmarshaler, bv.Interface()); err != nil {
				return err
			}
			rv.Set(reflect.Append(rv, bv.Elem()))
		}
		return nil
	}
	if _, ok := rv.Interface().(protoEnum); ok {
		var repr interface{}
		if err := d.Decode(&repr); err != nil {
			return err
		}
		switch v := repr.(type) {
		case string:
			return fmt.Errorf("unmarshaling of symbolic enum %q not supported: %T", repr, rv.Interface())
		case float64:
			rv.Set(reflect.ValueOf(int32(v)).Convert(rv.Type()))
			return nil
		default:
			return fmt.Errorf("cannot assign %#v into Go type %T", repr, rv.Interface())
		}
	}
	return d.Decode(v)
}

type protoEnum interface {
	fmt.Stringer
	EnumDescriptor() ([]byte, []int)
}

var typeProtoEnum = reflect.TypeOf((*protoEnum)(nil)).Elem()

var typeProtoMessage = reflect.TypeOf((*proto.Message)(nil)).Elem()

// Delimiter for newline encoded JSON streams.
func (j *JSONPb) Delimiter() []byte {
	return []byte("\n")
}

var (
	convFromType = map[reflect.Kind]reflect.Value{
		reflect.String:  reflect.ValueOf(convert.String),
		reflect.Bool:    reflect.ValueOf(convert.Bool),
		reflect.Float64: reflect.ValueOf(convert.Float64),
		reflect.Float32: reflect.ValueOf(convert.Float32),
		reflect.Int64:   reflect.ValueOf(convert.Int64),
		reflect.Int32:   reflect.ValueOf(convert.Int32),
		reflect.Uint64:  reflect.ValueOf(convert.Uint64),
		reflect.Uint32:  reflect.ValueOf(convert.Uint32),
		reflect.Slice:   reflect.ValueOf(convert.Bytes),
	}
)
