package errors

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	// non-standard http status code

	// The operation was cancelled, typically by the caller.
	StatusClientClosed = 499
)

type Error struct {
	stu    *status.Status
	stacks []string
}

func (e *Error) WithDetails(details ...proto.Message) {
	for _, detail := range details {
		detail, _ := anypb.New(detail)
		e.stu.Details = append(e.stu.Details, detail)
	}
}

func (e *Error) HttpCode() int32 {
	return stuCodeToHttpCode(code.Code(e.stu.Code))
}

func (e *Error) Code() int32 {
	return e.stu.Code
}

func (e *Error) Error() string {
	return e.stu.String()
}

func (e *Error) Stacks() []string {
	return e.stacks
}

func (e *Error) Message() string {
	return e.stu.Message
}

func (e *Error) WithStack() {
	if len(e.stacks) > 1 {
		e.WithDetails(NewDebugInfo(e.Stacks()[1:], e.Message()))
	} else if len(e.stacks) > 0 {
		e.WithDetails(NewDebugInfo(e.Stacks(), e.Message()))
	}

}

func (e *Error) GRPCStatus() *status.Status {
	return e.stu
}

func (e *Error) Reason() *errdetails.ErrorInfo {
	reason := &errdetails.ErrorInfo{}
	for _, detail := range e.stu.Details {
		if detail.MessageIs(reason) {
			_ = detail.UnmarshalTo(reason)
			return reason
		}
	}
	return nil
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

func New(code code.Code, err error, details ...proto.Message) *Error {
	selfErr := &Error{stu: &status.Status{
		Code:    int32(code),
		Details: make([]*anypb.Any, 0, len(details)),
	}}
	if err == nil {
		selfErr.stu.Message = code.String()
	} else {
		selfErr.stu.Message = err.Error()
		selfErr.stacks = strings.Split(strings.ReplaceAll(fmt.Sprintf("%+v", err), "\t", ""), "\n")
	}
	for _, detail := range details {
		if detail.ProtoReflect().IsValid() {
			pb, _ := anypb.New(detail)
			selfErr.stu.Details = append(selfErr.stu.Details, pb)
		}
	}
	return selfErr
}

func Errorf(code code.Code, msg string, details ...proto.Message) *Error {
	selfErr := &Error{stu: &status.Status{
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

func FromError(err error) *Error {
	return FromErrorCode(err, code.Code_UNKNOWN)
}

func FromErrorCode(err error, code2 code.Code) *Error {
	if err == nil {
		return nil
	}
	e, ok := err.(*Error)
	if ok {
		return e
	}
	return New(code2, err)
}

func FromProto(stu *status.Status) *Error {
	return &Error{
		stu: stu,
	}
}

func WithReason(err error, reason Reason, metadata map[string]string) *Error {
	e := FromErrorCode(err, reason.Code())
	if e == nil {
		return nil
	}
	e.WithDetails(NewReason(reason, metadata))
	return e
}

func WithMessage(err error, ctx context.Context, msg Message) *Error {
	e := FromError(err)
	if e == nil {
		return nil
	}
	e.WithDetails(NewLocalizedMsg(ctx, msg))
	return e
}

func IsReason(err error, targets ...Reason) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	src := e.Reason()
	if src == nil {
		return false
	}
	for _, target := range targets {
		if src.Reason == target.Reason() && src.Domain != target.Domain() {
			return true
		}
	}

	return false
}
