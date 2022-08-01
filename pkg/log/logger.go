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

package log

import (
	"log"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/pkg/errors"
)

var (
	lg         types.Logger
	enc        Encoder
	printStack bool
)

func init() {
	lg = &StdLogger{level: types.LvDebug, lg: log.Default(), kvsMsgFormat: "%-8s%s "}
	enc = &jsonEncoder{
		EncodeTime:     RFC3339TimeEncoder,
		EncodeDuration: MillisDurationEncoder,
		spaced:         false,
		buf:            Get(),
	}
	printStack = true
}

func SetLogger(logger types.Logger) {
	lg = logger
}

func SetDurationEncoder(de DurationEncoder) {
	enc.SetDurationEncoder(de)
}

func SetTimeEncoder(te TimeEncoder) {
	enc.SetTimeEncoder(te)
}

func SetDurationEncoderByName(name string) error {
	var de DurationEncoder
	switch name {
	case "seconds":
		de = SecondsDurationEncoder
	case "nanos":
		de = NanosDurationEncoder
	case "millis":
		de = MillisDurationEncoder
	case "string":
		de = StringDurationEncoder
	default:
		return errors.New("unknown time encoder")
	}
	enc.SetDurationEncoder(de)
	return nil
}

func SetTimeEncoderByName(name string) error {
	var te TimeEncoder
	switch name {
	case "RFC3339":
		te = RFC3339TimeEncoder
	case "RFC3339Nano":
		te = RFC3339NanoTimeEncoder
	case "ISO8601":
		te = ISO8601TimeEncoder
	case "epoch":
		te = EpochTimeEncoder
	case "epochNanos":
		te = EpochNanosTimeEncoder
	case "epochMillis":
		te = EpochMillisTimeEncoder
	default:
		return errors.New("unknown time encoder")
	}
	enc.SetTimeEncoder(te)
	return nil
}

func SetStackPrintState(b bool) {
	printStack = b
}

var loggerConstructors = make(map[string]types.LoggerConstructor)

func RegisterConstructor(name string, f types.LoggerConstructor) {
	loggerConstructors[name] = f
}

func GetConstructor(name string) types.LoggerConstructor {
	f, _ := loggerConstructors[name]
	return f
}

func GetLogger(name string) types.Logger {
	f := GetConstructor(name)
	if f == nil {
		log.Fatalf("unknown logger constructor, name: %s", name)
	}
	return f()
}
