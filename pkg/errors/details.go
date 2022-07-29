package errors

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/md"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type Reason interface {
	Reason() string
	Domain() string
	Code() code.Code
}

type Message interface {
	Message(language string) string
}

func NewReason(reason Reason, meta map[string]string) *errdetails.ErrorInfo {
	return &errdetails.ErrorInfo{
		Reason:   reason.Reason(),
		Domain:   reason.Domain(),
		Metadata: meta,
	}
}

func NewLocalizedMsg(ctx context.Context, msg Message) *errdetails.LocalizedMessage {
	var languages []string
	if meta, ok := md.FromInContext(ctx); ok {
		if values, ok := meta["language"]; ok {
			languages = append(values, languages...)
		}
	}
	return NewLocalizedMsgWithLang(languages, msg)
}

func NewLocalizedMsgWithLang(languages []string, msg Message) *errdetails.LocalizedMessage {
	languages = append(languages, "zh-CN")
	for _, language := range languages {
		localMsg := msg.Message(language)
		if len(localMsg) > 0 {
			return &errdetails.LocalizedMessage{
				Locale:  language,
				Message: localMsg,
			}
		}
	}
	return nil
}

func NewDebugInfo(stacks []string, msg string) *errdetails.DebugInfo {
	return &errdetails.DebugInfo{
		StackEntries: stacks,
		Detail:       msg,
	}
}
