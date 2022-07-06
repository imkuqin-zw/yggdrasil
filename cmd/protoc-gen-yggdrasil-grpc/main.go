package main

import (
	"flag"

	"github.com/imkuqin-zw/yggdrasil/internal/protogen/gengrpc"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			f.GoImportPath = protogen.GoImportPath(*f.Proto.Options.GoPackage)
			gengrpc.GenerateFiles(gen, f)
		}
		return nil
	})
}
