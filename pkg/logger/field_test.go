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
	"errors"
	"fmt"
	"math"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type users int

func (u users) String() string {
	return fmt.Sprintf("%d users", int(u))
}

func (u users) MarshalLogObject(enc ObjectEncoder) error {
	if int(u) < 0 {
		return errors.New("too few users")
	}
	enc.AddInt("users", int(u))
	return nil
}

func (u users) MarshalLogArray(enc ArrayEncoder) error {
	if int(u) < 0 {
		return errors.New("too few users")
	}
	for i := 0; i < int(u); i++ {
		enc.AppendString("user")
	}
	return nil
}

type obj struct {
	kind int
}

func (o *obj) String() string {
	if o == nil {
		return "nil obj"
	}

	if o.kind == 1 {
		panic("panic with string")
	} else if o.kind == 2 {
		panic(errors.New("panic with reason"))
	} else if o.kind == 3 {
		// panic with an arbitrary object that causes a panic itself
		// when being converted to a string
		panic((*url.URL)(nil))
	}

	return "obj"
}

type errObj struct {
	kind   int
	errMsg string
}

func (eobj *errObj) Error() string {
	if eobj.kind == 1 {
		panic("panic in Error() method")
	} else {
		return eobj.errMsg
	}
}

func TestUnknownFieldType(t *testing.T) {
	unknown := Field{Key: "k", String: "foo"}
	assert.Equal(t, UnknownType, unknown.Type, "Expected zero value of FieldType to be UnknownType.")
	assert.Panics(t, func() {
		unknown.AddTo(NewJsonEncoder())
	}, "Expected using a field with unknown type to panic.")
}

func TestFieldAddingError(t *testing.T) {
	var empty interface{}
	tests := []struct {
		t     FieldType
		iface interface{}
		want  interface{}
		err   string
	}{
		{t: ArrayMarshalerType, iface: users(-1), want: []interface{}{}, err: "too few users"},
		{t: ObjectMarshalerType, iface: users(-1), want: map[string]interface{}{}, err: "too few users"},
		{t: InlineMarshalerType, iface: users(-1), want: nil, err: "too few users"},
		{t: StringerType, iface: obj{}, want: empty, err: "PANIC=interface conversion: logger.obj is not fmt.Stringer: missing method String"},
		{t: StringerType, iface: &obj{1}, want: empty, err: "PANIC=panic with string"},
		{t: StringerType, iface: &obj{2}, want: empty, err: "PANIC=panic with reason"},
		{t: StringerType, iface: &obj{3}, want: empty, err: "PANIC=<nil>"},
		{t: ErrorType, iface: &errObj{kind: 1}, want: empty, err: "PANIC=panic in Error() method"},
	}
	for _, tt := range tests {
		f := Field{Key: "k", Interface: tt.iface, Type: tt.t}
		enc := NewMapObjectEncoder()
		assert.NotPanics(t, func() { f.AddTo(enc) }, "Unexpected panic when adding fields returns an reason.")
		assert.Equal(t, tt.want, enc.Fields["k"], "On reason, expected zero value in field.Key.")
		assert.Equal(t, tt.err, enc.Fields["kError"], "Expected reason message in global context.")
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		t     FieldType
		i     int64
		s     string
		iface interface{}
		want  interface{}
	}{
		{t: ArrayMarshalerType, iface: users(2), want: []interface{}{"user", "user"}},
		{t: ObjectMarshalerType, iface: users(2), want: map[string]interface{}{"users": 2}},
		{t: BoolType, i: 0, want: false},
		{t: ByteStringType, iface: []byte("foo"), want: "foo"},
		{t: Complex128Type, iface: 1 + 2i, want: 1 + 2i},
		{t: Complex64Type, iface: complex64(1 + 2i), want: complex64(1 + 2i)},
		{t: DurationType, i: 1000, want: time.Microsecond},
		{t: Float64Type, i: int64(math.Float64bits(3.14)), want: 3.14},
		{t: Float32Type, i: int64(math.Float32bits(3.14)), want: float32(3.14)},
		{t: Int64Type, i: 42, want: int64(42)},
		{t: Int32Type, i: 42, want: int32(42)},
		{t: Int16Type, i: 42, want: int16(42)},
		{t: Int8Type, i: 42, want: int8(42)},
		{t: StringType, s: "foo", want: "foo"},
		{t: TimeType, i: 1000, iface: time.UTC, want: time.Unix(0, 1000).In(time.UTC)},
		{t: TimeType, i: 1000, want: time.Unix(0, 1000)},
		{t: Uint64Type, i: 42, want: uint64(42)},
		{t: Uint32Type, i: 42, want: uint32(42)},
		{t: Uint16Type, i: 42, want: uint16(42)},
		{t: Uint8Type, i: 42, want: uint8(42)},
		{t: UintptrType, i: 42, want: uintptr(42)},
		{t: ReflectType, iface: users(2), want: users(2)},
		{t: NamespaceType, want: map[string]interface{}{}},
		{t: StringerType, iface: users(2), want: "2 users"},
		{t: StringerType, iface: &obj{}, want: "obj"},
		{t: StringerType, iface: (*obj)(nil), want: "nil obj"},
		{t: SkipType, want: interface{}(nil)},
		{t: StringerType, iface: (*url.URL)(nil), want: "<nil>"},
		{t: StringerType, iface: (*users)(nil), want: "<nil>"},
		{t: ErrorType, iface: (*errObj)(nil), want: "<nil>"},
	}

	for _, tt := range tests {
		enc := NewMapObjectEncoder()
		f := Field{Key: "k", Type: tt.t, Integer: tt.i, Interface: tt.iface, String: tt.s}
		f.AddTo(enc)
		assert.Equal(t, tt.want, enc.Fields["k"], "Unexpected output from field %+v.", f)

		delete(enc.Fields, "k")
		assert.Equal(t, 0, len(enc.Fields), "Unexpected extra fields present.")

		assert.True(t, f.Equals(f), "Field does not equal itself")
	}
}

func TestInlineMarshaler(t *testing.T) {
	enc := NewMapObjectEncoder()

	topLevelStr := Field{Key: "k", Type: StringType, String: "s"}
	topLevelStr.AddTo(enc)

	inlineObj := Field{Key: "ignored", Type: InlineMarshalerType, Interface: users(10)}
	inlineObj.AddTo(enc)

	nestedObj := Field{Key: "nested", Type: ObjectMarshalerType, Interface: users(11)}
	nestedObj.AddTo(enc)

	assert.Equal(t, map[string]interface{}{
		"k":     "s",
		"users": 10,
		"nested": map[string]interface{}{
			"users": 11,
		},
	}, enc.Fields)
}

func TestEquals(t *testing.T) {
	// Values outside the UnixNano range were encoded incorrectly (#737, #803).
	timeOutOfRangeHigh := time.Unix(0, math.MaxInt64).Add(time.Nanosecond)
	timeOutOfRangeLow := time.Unix(0, math.MinInt64).Add(-time.Nanosecond)
	timeOutOfRangeHighNano := time.Unix(0, timeOutOfRangeHigh.UnixNano())
	timeOutOfRangeLowNano := time.Unix(0, timeOutOfRangeLow.UnixNano())
	require.False(t, timeOutOfRangeHigh.Equal(timeOutOfRangeHighNano), "should be different as value is >  UnixNano range")
	require.False(t, timeOutOfRangeHigh.Equal(timeOutOfRangeHighNano), "should be different as value is <  UnixNano range")

	tests := []struct {
		a, b Field
		want bool
	}{
		{
			a:    Int16("a", 1),
			b:    Int32("a", 1),
			want: false,
		},
		{
			a:    String("k", "a"),
			b:    String("k", "a"),
			want: true,
		},
		{
			a:    String("k", "a"),
			b:    String("k2", "a"),
			want: false,
		},
		{
			a:    String("k", "a"),
			b:    String("k", "b"),
			want: false,
		},
		{
			a:    Time("k", time.Unix(1000, 1000)),
			b:    Time("k", time.Unix(1000, 1000)),
			want: true,
		},
		{
			a:    Time("k", time.Unix(1000, 1000).In(time.UTC)),
			b:    Time("k", time.Unix(1000, 1000).In(time.FixedZone("TEST", -8))),
			want: false,
		},
		{
			a:    Time("k", timeOutOfRangeLow),
			b:    Time("k", timeOutOfRangeLowNano),
			want: false,
		},
		{
			a:    Time("k", timeOutOfRangeHigh),
			b:    Time("k", timeOutOfRangeHighNano),
			want: false,
		},
		{
			a:    Time("k", time.Unix(1000, 1000)),
			b:    Time("k", time.Unix(1000, 2000)),
			want: false,
		},
		{
			a:    Binary("k", []byte{1, 2}),
			b:    Binary("k", []byte{1, 2}),
			want: true,
		},
		{
			a:    Binary("k", []byte{1, 2}),
			b:    Binary("k", []byte{1, 3}),
			want: false,
		},
		{
			a:    ByteString("k", []byte("abc")),
			b:    ByteString("k", []byte("abc")),
			want: true,
		},
		{
			a:    ByteString("k", []byte("abc")),
			b:    ByteString("k", []byte("abd")),
			want: false,
		},
		{
			a:    Ints("k", []int{1, 2}),
			b:    Ints("k", []int{1, 2}),
			want: true,
		},
		{
			a:    Ints("k", []int{1, 2}),
			b:    Ints("k", []int{1, 3}),
			want: false,
		},
		{
			a:    Object("k", users(10)),
			b:    Object("k", users(10)),
			want: true,
		},
		{
			a:    Object("k", users(10)),
			b:    Object("k", users(20)),
			want: false,
		},
		{
			a:    Any("k", map[string]string{"a": "b"}),
			b:    Any("k", map[string]string{"a": "b"}),
			want: true,
		},
		{
			a:    Any("k", map[string]string{"a": "b"}),
			b:    Any("k", map[string]string{"a": "d"}),
			want: false,
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.a.Equals(tt.b), "a.Equals(b) a: %#v b: %#v", tt.a, tt.b)
		assert.Equal(t, tt.want, tt.b.Equals(tt.a), "b.Equals(a) a: %#v b: %#v", tt.a, tt.b)
	}
}
