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

package genrpc

import (
	"fmt"

	error2 "github.com/imkuqin-zw/yggdrasil/proto/yggdrasil/reason"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

const (
	eCodeMin = uint32(code.Code_OK)
	eCodeMax = uint32(code.Code_DATA_LOSS)
)

const (
	codePackage = protogen.GoImportPath("google.golang.org/genproto/googleapis/rpc/code")
)

func GenerateFile(gen *protogen.Plugin, file *protogen.File) {
	var filename string
	filename = file.GeneratedFilenamePrefix + "_reason.pb.go"

	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	generateHeader(g, file)
	generateFileContent(gen, file, g, true)
}

func generateHeader(g *protogen.GeneratedFile, file *protogen.File) {
	g.P("// Code generated by protoc-gen-yggdrasil-reason. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
}

// generateFileContent generates the kratos errors definitions, excluding the package statement.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, isServer bool) {
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the yggdrasil package it is being compiled against.")
	g.P()

	reasons := &Reasons{
		Domain:      string(file.Desc.Package()),
		CodePackage: g.QualifiedGoIdent(codePackage.Ident("")),
	}
	for _, enum := range file.Enums {
		genErrorsReason(enum, reasons)
	}
	// If all enums do not contain 'errors.code', the current file is skipped
	if len(reasons.Reason) == 0 {
		g.Skip()
	}
	g.P(reasons.execute())
}

func genErrorsReason(enum *protogen.Enum, reasons *Reasons) {
	if !proto.HasExtension(enum.Desc.Options(), error2.E_DefaultReason) {
		return
	}
	rw := ReasonWrapper{Name: string(enum.Desc.Name()), Codes: map[int32]uint32{}}
	for _, value := range enum.Values {
		errCode := proto.GetExtension(value.Desc.Options(), error2.E_Code)
		eCode := errCode.(uint32)
		if eCode < eCodeMin || eCode > eCodeMax {
			panic(fmt.Sprintf("code of Enum '%s'.'%s' range must be between %d and %d",
				string(enum.Desc.Name()), string(value.Desc.Name()), eCodeMin, eCodeMax))
		}
		rw.Codes[int32(value.Desc.Number())] = eCode
	}
	reasons.Reason = append(reasons.Reason, rw)
}