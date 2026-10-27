package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/jenken827/protoc-gen-sgorm/internal/gen"
)

func main() {
	var flags flag.FlagSet
	pkg := flags.String("package", "", "生成文件的 Go 包名（默认沿用输入 proto 的 go_package 名）")
	suffix := flags.String("suffix", ".sgorm.go", "生成文件后缀")

	protogen.Options{ParamFunc: flags.Set}.Run(func(plugin *protogen.Plugin) error {
		// proto3 optional 必须声明支持，否则 protoc/buf 拒绝生成。
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		return gen.Generate(plugin, gen.Options{
			Package: *pkg,
			Suffix:  *suffix,
		})
	})
}
