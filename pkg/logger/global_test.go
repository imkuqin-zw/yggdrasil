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
	"errors"
	"testing"
	"time"
)

func Test_logger(t *testing.T) {
	Debug("Debug", "Debug")
	Debugf("this %s", "Debugf")

	Info("Info", "Info")
	Infof("this %s", "Infof")

	Warn("Warn", "Earn")
	Warnf("this %s", "Warnf")

	Error("Error", "Error")
	Errorf("this %s", "Errorf")
	DebugField("access", String("fd", "fd"), Int("fd", 456))
	InfoField("access", String("fd", "fd"), Int("fd", 456))
	ErrorField("access", String("fd", "fd"), Int("fd", 456), Err(errors.New("fdasfdf")))

}

func Test(t *testing.T) {
	start := time.Now()
	defer func() {
		Debugf("free: %.3f", time.Since(start).Seconds())
	}()
	time.Sleep(time.Millisecond * 3)
}

func Test_WithFields(t *testing.T) {
	lg := WithFields(String("mod", "test"))
	lg.Debug("Debug")
	lg.Info("Info")
	lg.Warn("Warn")
	lg.Error("Error")
	//lg.Fatal("Fatal")

	lg.Debugf("Debug %s", "test")
	lg.Infof("Info %s", "test")
	lg.Warnf("Warn %s", "test")
	lg.Errorf("Error %s", "test")
	//lg.Fatalf("Fatal %s", "test")

	lg.DebugField("Debug", String("k", "test"))
	lg.InfoField("Info", String("k", "test"))
	lg.WarnField("Warn", String("k", "test"))
	lg.ErrorField("Error", String("k", "test"))
	lg.FatalField("Fatal", String("k", "test"))
}
