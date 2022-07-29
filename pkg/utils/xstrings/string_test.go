// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xstrings

import (
	"container/list"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var s = strings.Repeat("a", 1024)

func TestClearByteMemory(t *testing.T) {
	b := []byte("aaa")
	assert.Equal(t, "aaa", string(b))
	ClearByteMemory(b)
	assert.NotEqual(t, "aaa", string(b))
}

func TestMinInt(t *testing.T) {

	a := MinInt(1, 2)
	assert.Equal(t, 1, a)

	a = MinInt(2, 1)
	assert.Equal(t, 1, a)
}

func BenchmarkTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := []byte(s)
		_ = string(b)
	}
}

func BenchmarkTestBlock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := []byte(s)
		_ = Bytes2str(b)
	}
}
func BenchmarkTest1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = []byte(s)

	}
}

func BenchmarkTestBlock2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Str2bytes("asd")
	}
}

const capacity = 2

func array() [capacity]int {
	var d [capacity]int

	for i := 0; i < len(d); i++ {
		d[i] = 1
	}

	return d
}

func slice() []int {
	d := make([]int, capacity)

	for i := 0; i < len(d); i++ {
		d[i] = 1
	}

	return d
}

var l *list.List

func init() {
	l = list.New()
	for i := 0; i < capacity; i++ {
		l.PushBack(i)
	}
}
func List() {

	for e := l.Front(); e != nil; e = e.Next() {
		_ = e.Value.(int)
	}

}
func BenchmarkArray(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = array()
	}
}

func BenchmarkSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = slice()
	}

}
func BenchmarkList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		List()
	}

}

func TestStringFunc(t *testing.T) {

	var strArr = []string{"abc", "def"}

	b := StrInSlice("abc", strArr)
	assert.Equal(t, true, b)

	b = StrInSlice("wer", strArr)
	assert.Equal(t, false, b)

	by := Str2bytes("abc")
	str := Bytes2str(by)

	assert.Equal(t, str, string(by))
}

func TestSplitToTwo(t *testing.T) {
	s1 := "aa"
	s2 := "bb"
	sub := "::"
	s := s1 + sub + s2
	p1, p2 := SplitToTwo(s, sub)
	assert.Equal(t, s1, p1)
	assert.Equal(t, s2, p2)

	p1, p2 = SplitToTwo(s, "/")
	assert.Empty(t, p1)
	assert.Equal(t, s, p2)
}

func TestSplitFirstSep(t *testing.T) {
	s1 := "aa"
	s2 := "bb"
	sub := "::"
	s := s1 + sub + s2
	p1 := SplitFirstSep(s, sub)
	assert.Equal(t, s1, p1)

	p1 = SplitFirstSep(s, "/")
	assert.Empty(t, p1)
}

func SplitToTwoByStringsSplit(s, sep string) (string, string) {
	r := strings.Split(s, sep)
	return r[0], r[1]
}

func SplitFirstSepByStringsSplit(s, sep string) string {
	return strings.Split(s, sep)[0]
}

var testURL = "http://127.0.0.1"

func Benchmark_SplitToTwo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SplitToTwo(testURL, "://")
	}
}

func Benchmark_SplitToTwoByStringsSplit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SplitToTwoByStringsSplit(testURL, "://")
	}
}

func Benchmark_SplitFirstSep(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SplitFirstSep(testURL, "://")
	}
}

func Benchmark_SplitFirstSepByStringsSplit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SplitFirstSepByStringsSplit(testURL, "://")
	}
}

func TestHasPrefix(t *testing.T) {
	type args struct {
		key       string
		pre       string
		delimiter string
	}
	tests := []struct {
		args args
		want bool
	}{
		{args{"app_database_host", "app", "_"}, true},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := HasPrefix(tt.args.key, tt.args.pre, tt.args.delimiter)
			assert.Equal(t, tt.want, got)
		})
	}
}
