// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package logger

import (
	"encoding/json"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/logger/buffer"
)

// DefaultLineEnding defines the default line ending when writing logs.
// Alternate line endings specified in EncoderConfig can override this
// behavior.
const DefaultLineEnding = "\n"

// OmitKey defines the key to use when callers want to remove a key from global output.
const OmitKey = ""

// A TimeEncoder serializes a time.Time to a primitive type.
type TimeEncoder func(time.Time, PrimitiveArrayEncoder)

// EpochTimeEncoder serializes a time.Time to a floating-point number of seconds
// since the Unix epoch.
func EpochTimeEncoder(t time.Time, enc PrimitiveArrayEncoder) {
	nanos := t.UnixNano()
	sec := float64(nanos) / float64(time.Second)
	enc.AppendFloat64(sec)
}

// EpochMillisTimeEncoder serializes a time.Time to a floating-point number of
// milliseconds since the Unix epoch.
func EpochMillisTimeEncoder(t time.Time, enc PrimitiveArrayEncoder) {
	nanos := t.UnixNano()
	millis := float64(nanos) / float64(time.Millisecond)
	enc.AppendFloat64(millis)
}

// EpochNanosTimeEncoder serializes a time.Time to an integer number of
// nanoseconds since the Unix epoch.
func EpochNanosTimeEncoder(t time.Time, enc PrimitiveArrayEncoder) {
	enc.AppendInt64(t.UnixNano())
}

func encodeTimeLayout(t time.Time, layout string, enc PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

// ISO8601TimeEncoder serializes a time.Time to an ISO8601-formatted string
// with millisecond precision.
//
// If enc supports AppendTimeLayout(t time.Time,layout string), it's used
// instead of appending a pre-formatted string value.
func ISO8601TimeEncoder(t time.Time, enc PrimitiveArrayEncoder) {
	encodeTimeLayout(t, "2006-01-02T15:04:05.000Z0700", enc)
}

// RFC3339TimeEncoder serializes a time.Time to an RFC3339-formatted string.
//
// If enc supports AppendTimeLayout(t time.Time,layout string), it's used
// instead of appending a pre-formatted string value.
func RFC3339TimeEncoder(t time.Time, enc PrimitiveArrayEncoder) {
	encodeTimeLayout(t, time.RFC3339, enc)
}

// RFC3339NanoTimeEncoder serializes a time.Time to an RFC3339-formatted string
// with nanosecond precision.
//
// If enc supports AppendTimeLayout(t time.Time,layout string), it's used
// instead of appending a pre-formatted string value.
func RFC3339NanoTimeEncoder(t time.Time, enc PrimitiveArrayEncoder) {
	encodeTimeLayout(t, time.RFC3339Nano, enc)
}

// TimeEncoderOfLayout returns TimeEncoder which serializes a time.Time using
// given layout.
func TimeEncoderOfLayout(layout string) TimeEncoder {
	return func(t time.Time, enc PrimitiveArrayEncoder) {
		encodeTimeLayout(t, layout, enc)
	}
}

// UnmarshalText unmarshals text to a TimeEncoder.
// "rfc3339nano" and "RFC3339Nano" are unmarshaled to RFC3339NanoTimeEncoder.
// "rfc3339" and "RFC3339" are unmarshaled to RFC3339TimeEncoder.
// "iso8601" and "ISO8601" are unmarshaled to ISO8601TimeEncoder.
// "millis" is unmarshaled to EpochMillisTimeEncoder.
// "nanos" is unmarshaled to EpochNanosEncoder.
// Anything else is unmarshaled to EpochTimeEncoder.
func (e *TimeEncoder) UnmarshalText(text []byte) error {
	switch string(text) {
	case "rfc3339nano", "RFC3339Nano":
		*e = RFC3339NanoTimeEncoder
	case "rfc3339", "RFC3339":
		*e = RFC3339TimeEncoder
	case "iso8601", "ISO8601":
		*e = ISO8601TimeEncoder
	case "millis":
		*e = EpochMillisTimeEncoder
	case "nanos":
		*e = EpochNanosTimeEncoder
	default:
		*e = EpochTimeEncoder
	}
	return nil
}

// UnmarshalYAML unmarshals YAML to a TimeEncoder.
// If value is an object with a "layout" field, it will be unmarshaled to  TimeEncoder with given layout.
//
//	timeEncoder:
//	  layout: 06/01/02 03:04pm
//
// If value is string, it uses UnmarshalText.
//
//	timeEncoder: iso8601
func (e *TimeEncoder) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var o struct {
		Layout string `json:"layout" yaml:"layout"`
	}
	if err := unmarshal(&o); err == nil {
		*e = TimeEncoderOfLayout(o.Layout)
		return nil
	}

	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	return e.UnmarshalText([]byte(s))
}

// UnmarshalJSON unmarshals JSON to a TimeEncoder as same way UnmarshalYAML does.
func (e *TimeEncoder) UnmarshalJSON(data []byte) error {
	return e.UnmarshalYAML(func(v interface{}) error {
		return json.Unmarshal(data, v)
	})
}

// A DurationEncoder serializes a time.Duration to a primitive type.
type DurationEncoder func(time.Duration, PrimitiveArrayEncoder)

// SecondsDurationEncoder serializes a time.Duration to a floating-point number of seconds elapsed.
func SecondsDurationEncoder(d time.Duration, enc PrimitiveArrayEncoder) {
	enc.AppendFloat64(float64(d) / float64(time.Second))
}

// NanosDurationEncoder serializes a time.Duration to an integer number of
// nanoseconds elapsed.
func NanosDurationEncoder(d time.Duration, enc PrimitiveArrayEncoder) {
	enc.AppendInt64(int64(d))
}

// MillisDurationEncoder serializes a time.Duration to an integer number of
// milliseconds elapsed.
func MillisDurationEncoder(d time.Duration, enc PrimitiveArrayEncoder) {
	enc.AppendInt64(d.Nanoseconds() / 1e6)
}

// StringDurationEncoder serializes a time.Duration using its built-in String
// method.
func StringDurationEncoder(d time.Duration, enc PrimitiveArrayEncoder) {
	enc.AppendString(d.String())
}

// UnmarshalText unmarshals text to a DurationEncoder. "string" is unmarshaled
// to StringDurationEncoder, and anything else is unmarshaled to
// NanosDurationEncoder.
func (e *DurationEncoder) UnmarshalText(text []byte) error {
	switch string(text) {
	case "string":
		*e = StringDurationEncoder
	case "nanos":
		*e = NanosDurationEncoder
	case "ms":
		*e = MillisDurationEncoder
	default:
		*e = SecondsDurationEncoder
	}
	return nil
}

// A NameEncoder serializes a period-separated global name to a primitive
// type.
type NameEncoder func(string, PrimitiveArrayEncoder)

// FullNameEncoder serializes the global name as-is.
func FullNameEncoder(loggerName string, enc PrimitiveArrayEncoder) {
	enc.AppendString(loggerName)
}

// UnmarshalText unmarshals text to a NameEncoder. Currently, everything is
// unmarshaled to FullNameEncoder.
func (e *NameEncoder) UnmarshalText(text []byte) error {
	switch string(text) {
	case "full":
		*e = FullNameEncoder
	default:
		*e = FullNameEncoder
	}
	return nil
}

// ObjectEncoder is a strongly-typed, encoding-agnostic interface for adding a
// map- or struct-like object to the logging context. Like maps, ObjectEncoders
// aren't safe for concurrent use (though typical use shouldn't require locks).
type ObjectEncoder interface {
	// Logging-specific marshalers.
	AddArray(key string, marshaler ArrayMarshaler) error
	AddObject(key string, marshaler ObjectMarshaler) error

	// Built-in
	AddBinary(key string, value []byte)     // for arbitrary bytes
	AddByteString(key string, value []byte) // for UTF-8 encoded bytes
	AddBool(key string, value bool)
	AddComplex128(key string, value complex128)
	AddComplex64(key string, value complex64)
	AddDuration(key string, value time.Duration)
	AddFloat64(key string, value float64)
	AddFloat32(key string, value float32)
	AddInt(key string, value int)
	AddInt64(key string, value int64)
	AddInt32(key string, value int32)
	AddInt16(key string, value int16)
	AddInt8(key string, value int8)
	AddString(key, value string)
	AddTime(key string, value time.Time)
	AddUint(key string, value uint)
	AddUint64(key string, value uint64)
	AddUint32(key string, value uint32)
	AddUint16(key string, value uint16)
	AddUint8(key string, value uint8)
	AddUintptr(key string, value uintptr)

	// AddReflected uses reflection to serialize arbitrary objects, so it can be
	// slow and allocation-heavy.
	AddReflected(key string, value interface{}) error
	// OpenNamespace opens an isolated namespace where all subsequent fields will
	// be added. Applications can use namespaces to prevent key collisions when
	// injecting loggers into sub-components or third-party libraries.
	OpenNamespace(key string)

	SetDurationEncoder(de DurationEncoder)
	SetTimeEncoder(te TimeEncoder)
}

// ArrayEncoder is a strongly-typed, encoding-agnostic interface for adding
// array-like objects to the logging context. Of note, it supports mixed-type
// arrays even though they aren't typical in Go. Like slices, ArrayEncoders
// aren't safe for concurrent use (though typical use shouldn't require locks).
type ArrayEncoder interface {
	// Built-in
	PrimitiveArrayEncoder

	// Time-related
	AppendDuration(time.Duration)
	AppendTime(time.Time)

	// Logging-specific marshalers.
	AppendArray(ArrayMarshaler) error
	AppendObject(ObjectMarshaler) error

	// AppendReflected uses reflection to serialize arbitrary objects, so it's
	// slow and allocation-heavy.
	AppendReflected(value interface{}) error
}

// PrimitiveArrayEncoder is the subset of the ArrayEncoder interface that deals
// only in Go's built-in  It's included only so that Duration- and
// TimeEncoders cannot trigger infinite recursion.
type PrimitiveArrayEncoder interface {
	// Built-in
	AppendBool(bool)
	AppendByteString([]byte) // for UTF-8 encoded bytes
	AppendComplex128(complex128)
	AppendComplex64(complex64)
	AppendFloat64(float64)
	AppendFloat32(float32)
	AppendInt(int)
	AppendInt64(int64)
	AppendInt32(int32)
	AppendInt16(int16)
	AppendInt8(int8)
	AppendString(string)
	AppendUint(uint)
	AppendUint64(uint64)
	AppendUint32(uint32)
	AppendUint16(uint16)
	AppendUint8(uint8)
	AppendUintptr(uintptr)
}

// Encoder is a format-agnostic interface for all global entry marshalers. Since
// global encoders don't need to support the same wide range of use cases as
// general-purpose marshalers, it's possible to make them faster and
// lower-allocation.
//
// Implementations of the ObjectEncoder interface's methods can, of course,
// freely modify the receiver. However, the Clone and EncodeEntry methods will
// be called concurrently and shouldn't modify the receiver.
type Encoder interface {
	ObjectEncoder

	// Clone copies the encoder, ensuring that adding fields to the copy doesn't
	// affect the original.
	Clone() Encoder

	// EncodeEntry encodes an entry and fields, along with any accumulated
	// context, into a byte buffer and returns it. Any fields that are empty,
	// including fields on the `Entry` type, should be omitted.
	Encode([]Field) (*buffer.Buffer, error)
}
