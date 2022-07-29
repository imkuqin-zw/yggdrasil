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
