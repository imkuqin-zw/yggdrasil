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

package logger

import (
	"bytes"
	"fmt"
	"time"
)

type MemoryWriter struct {
	printTime  bool
	timeEncode func(t time.Time) string
	buf        *bytes.Buffer
}

func NewMemoryWriter(buf *bytes.Buffer, printTime bool, timeEncode func(t time.Time) string) Writer {
	if timeEncode == nil {
		timeEncode = func(t time.Time) string {
			return fmt.Sprintf(`"%s"`, t.Format(time.RFC3339))
		}
	}
	return &MemoryWriter{buf: buf, timeEncode: timeEncode, printTime: printTime}
}

func (l *MemoryWriter) Write(lv Level, t time.Time, msg string, ext ...[]byte) {
	buf := l.buf
	_ = buf.WriteByte('{')
	if !t.IsZero() && l.printTime {
		_, _ = fmt.Fprintf(buf, `"time":%s,`, l.timeEncode(t))
	}
	_, _ = buf.WriteString(`"level":"`)
	_, _ = buf.WriteString(lv.String())
	_, _ = buf.WriteString(`",`)
	_, _ = buf.WriteString(`"msg":"`)
	_, _ = buf.WriteString(msg)
	_, _ = buf.WriteString(`"`)
	if len(ext) > 0 {
		_ = buf.WriteByte(',')
		_, _ = buf.Write(ext[0])
	}
	_ = buf.WriteByte('}')
	buf.WriteByte('\n')
}
