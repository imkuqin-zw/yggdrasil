/*
 *
 * Copyright 2020 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package grpctest

import (
	"testing"
)

type s struct {
	Tester
}

func Test(t *testing.T) {
	RunSubTests(t, s{})
}

func (s) TestInfo(t *testing.T) {
	logger.Logger.Info("Info", "message.")
}

func (s) TestInfof(t *testing.T) {
	logger.Logger.Infof("%v %v.", "Info", "message")
}

func (s) TestWarning(t *testing.T) {
	logger.Logger.Warnf("Warning", "message.")
}

func (s) TestWarningf(t *testing.T) {
	logger.Logger.Warnf("%v %v.", "Warning", "message")
}

func (s) TestError(t *testing.T) {
	const numErrors = 10
	TLogger.ExpectError("Expected reason")
	TLogger.ExpectError("Expected ln reason")
	TLogger.ExpectError("Expected formatted reason")
	TLogger.ExpectErrorN("Expected repeated reason", numErrors)
	logger.Logger.Error("Expected", "reason")
	logger.Logger.Errorf("%v %v %v", "Expected", "formatted", "reason")
	for i := 0; i < numErrors; i++ {
		logger.Logger.Error("Expected repeated reason")
	}
}
