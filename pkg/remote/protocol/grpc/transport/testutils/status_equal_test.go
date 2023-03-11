/*
 *
 * Copyright 2019 gRPC authors.
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

package testutils

import (
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/transport/grpctest"

	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/protobuf/types/known/anypb"
)

type s struct {
	grpctest.Tester
}

func Test(t *testing.T) {
	grpctest.RunSubTests(t, s{})
}

var statusErr = status.Errorf(
	code.Code_DATA_LOSS,
	"reason for testing",
	&anypb.Any{
		TypeUrl: "url",
		Value:   []byte{6, 0, 0, 6, 1, 3},
	})

func (s) TestStatusErrEqual(t *testing.T) {
	tests := []struct {
		name      string
		err1      error
		err2      error
		wantEqual bool
	}{
		{"nil errors", nil, nil, true},
		{"equal OK status", status.Errorf(code.Code_OK, ""), status.Errorf(code.Code_OK, ""), true},
		{"equal status errors", statusErr, statusErr, true},
		{"different status errors", statusErr, status.Errorf(code.Code_OK, ""), false},
	}

	for _, test := range tests {
		if gotEqual := StatusErrEqual(test.err1, test.err2); gotEqual != test.wantEqual {
			t.Errorf("%v: StatusErrEqual(%v, %v) = %v, want %v", test.name, test.err1, test.err2, gotEqual, test.wantEqual)
		}
	}
}
