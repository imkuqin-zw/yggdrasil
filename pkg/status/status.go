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

package status

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	// non-standard http errors code

	// The operation was cancelled, typically by the caller.
	StatusClientClosed = 499
)

type Status struct {
	stu    *status.Status
	stacks []string
}

func (e *Status) WithDetails(details ...proto.Message) {
	if e == nil || e.stu == nil {
		return
	}
	for _, detail := range details {
		detail, _ := anypb.New(detail)
		e.stu.Details = append(e.stu.Details, detail)
	}
}

func (e *Status) HttpCode() int32 {
	if e == nil || e.stu == nil {
		return http.StatusOK
	}
	return stuCodeToHttpCode(code.Code(e.stu.Code))
}

func (e *Status) Code() int32 {
	if e == nil || e.stu == nil {
		return int32(code.Code_OK)
	}
	return e.stu.Code
}

func (e *Status) Err() error {
	if e.Code() == int32(code.Code_OK) {
		return nil
	}
	return e
}

func (e *Status) Error() string {
	if e == nil || e.stu == nil {
		return ""
	}
	return e.stu.String()
}

func (e *Status) Stacks() []string {
	if e == nil {
		return nil
	}
	return e.stacks
}

func (e *Status) Message() string {
	if e == nil || e.stu == nil {
		return ""
	}
	return e.stu.Message
}

func (e *Status) WithStack() {
	if len(e.stacks) > 1 {
		e.WithDetails(NewDebugInfo(e.Stacks()[1:], e.Message()))
	} else if len(e.stacks) > 0 {
		e.WithDetails(NewDebugInfo(e.Stacks(), e.Message()))
	}
}

func (e *Status) Status() *status.Status {
	if e == nil || e.stu == nil {
		return nil
	}
	return proto.Clone(e.stu).(*status.Status)
}

func (e *Status) Reason() *errdetails.ErrorInfo {
	if e != nil {
		reason := &errdetails.ErrorInfo{}
		for _, detail := range e.stu.Details {
			if detail.MessageIs(reason) {
				_ = detail.UnmarshalTo(reason)
				return reason
			}
		}
	}
	return nil
}

func (e *Status) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.Message())
			for i := 0; i < len(e.stacks); i += 2 {
				_, _ = io.WriteString(s, "\n")
				_, _ = io.WriteString(s, e.stacks[i])
				_, _ = io.WriteString(s, "\n\t")
				_, _ = io.WriteString(s, e.stacks[i+1])
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Message())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}

func stuCodeToHttpCode(stuCode code.Code) int32 {
	switch stuCode {
	case code.Code_OK:
		return http.StatusOK
	case code.Code_CANCELLED:
		return StatusClientClosed
	case code.Code_UNKNOWN:
		return http.StatusInternalServerError
	case code.Code_INVALID_ARGUMENT:
		return http.StatusBadRequest
	case code.Code_DEADLINE_EXCEEDED:
		return http.StatusGatewayTimeout
	case code.Code_NOT_FOUND:
		return http.StatusNotFound
	case code.Code_ALREADY_EXISTS:
		return http.StatusConflict
	case code.Code_PERMISSION_DENIED:
		return http.StatusForbidden
	case code.Code_UNAUTHENTICATED:
		return http.StatusUnauthorized
	case code.Code_RESOURCE_EXHAUSTED:
		return http.StatusTooManyRequests
	case code.Code_FAILED_PRECONDITION:
		return http.StatusBadRequest
	case code.Code_ABORTED:
		return http.StatusConflict
	case code.Code_OUT_OF_RANGE:
		return http.StatusBadRequest
	case code.Code_UNIMPLEMENTED:
		return http.StatusNotImplemented
	case code.Code_INTERNAL:
		return http.StatusInternalServerError
	case code.Code_UNAVAILABLE:
		return http.StatusServiceUnavailable
	case code.Code_DATA_LOSS:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}

func httpCodeToStuCode(httpCode int32) code.Code {
	switch httpCode {
	case http.StatusOK:
		return code.Code_OK
	case StatusClientClosed:
		return code.Code_CANCELLED
	case http.StatusBadRequest:
		return code.Code_INVALID_ARGUMENT
	case http.StatusGatewayTimeout:
		return code.Code_DEADLINE_EXCEEDED
	case http.StatusNotFound:
		return code.Code_NOT_FOUND
	case http.StatusForbidden:
		return code.Code_PERMISSION_DENIED
	case http.StatusUnauthorized:
		return code.Code_UNAUTHENTICATED
	case http.StatusTooManyRequests:
		return code.Code_RESOURCE_EXHAUSTED
	case http.StatusConflict:
		return code.Code_ABORTED
	case http.StatusNotImplemented:
		return code.Code_UNIMPLEMENTED
	case http.StatusInternalServerError:
		return code.Code_INTERNAL
	case http.StatusServiceUnavailable:
		return code.Code_UNAVAILABLE
	}
	return http.StatusInternalServerError
}

func New(code code.Code, err error, details ...proto.Message) *Status {
	selfErr := &Status{stu: &status.Status{
		Code:    int32(code),
		Details: make([]*anypb.Any, 0, len(details)),
	}}
	if err == nil {
		selfErr.stu.Message = code.String()
	} else {
		selfErr.stu.Message = err.Error()
		selfErr.stacks = strings.Split(strings.ReplaceAll(fmt.Sprintf("%+v", err), "\t", ""), "\n")
		if len(selfErr.stacks) > 0 {
			selfErr.stacks = selfErr.stacks[1:]
		}
	}
	for _, detail := range details {
		if detail.ProtoReflect().IsValid() {
			pb, _ := anypb.New(detail)
			selfErr.stu.Details = append(selfErr.stu.Details, pb)
		}
	}
	return selfErr
}

func Errorf(code code.Code, msg string, details ...proto.Message) *Status {
	selfErr := &Status{stu: &status.Status{
		Code:    int32(code),
		Message: msg,
		Details: make([]*anypb.Any, 0, len(details)),
	}}
	for _, detail := range details {
		if detail.ProtoReflect().IsValid() {
			pb, _ := anypb.New(detail)
			selfErr.stu.Details = append(selfErr.stu.Details, pb)
		}
	}
	return selfErr
}

func FromReason(err error, reason Reason, metadata map[string]string) *Status {
	e := FromErrorCode(err, reason.Code())
	if e == nil {
		return nil
	}
	e.WithDetails(NewReason(reason, metadata))
	return e
}

func WithMessage(err error, ctx context.Context, msg Message) *Status {
	e := FromError(err)
	if e == nil {
		return nil
	}
	e.WithDetails(NewLocalizedMsg(ctx, msg))
	return e
}

func FromHttpCode(httpCode int32, err error, details ...proto.Message) *Status {
	return New(httpCodeToStuCode(httpCode), err, details...)
}

func coverError(err error) (*Status, bool) {
	if err == nil {
		return nil, true
	}
	s, ok := errors.Unwrap(err).(*Status)
	if ok {
		return s, true
	}
	return nil, false
}

func FromError(err error) *Status {
	return FromErrorCode(err, code.Code_UNKNOWN)
}

func CoverError(err error) (*Status, bool) {
	st, ok := coverError(err)
	if ok {
		return st, ok
	}
	return New(code.Code_UNKNOWN, err), false
}

func FromErrorCode(err error, code2 code.Code) *Status {
	st, ok := coverError(err)
	if ok {
		return st
	}
	return New(code2, err)
}

// FromContextError converts a context reason or wrapped context reason into a
// Status.  It returns a Status with codes.OK if err is nil, or a Status with
// codes.Unknown if err is non-nil and not a context reason.
func FromContextError(err error) *Status {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return New(code.Code_DEADLINE_EXCEEDED, err)
	}
	if errors.Is(err, context.Canceled) {
		return New(code.Code_CANCELLED, err)
	}
	return New(code.Code_UNKNOWN, err)
}

func FromProto(stu *status.Status) *Status {
	return &Status{stu: stu}
}

func IsReason(err error, targets ...Reason) bool {
	e, ok := coverError(err)
	if !ok {
		return false
	}
	src := e.Reason()
	if src == nil {
		return false
	}
	for _, target := range targets {
		if src.Reason == target.Reason() && src.Domain == target.Domain() {
			return true
		}
	}

	return false
}

func IsCode(err error, code2 code.Code) bool {
	e, ok := coverError(err)
	if !ok {
		return false
	}
	return e.Code() == int32(code2)
}
