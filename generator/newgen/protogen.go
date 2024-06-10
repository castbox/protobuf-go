package newgen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/tools/imports"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

// Options 是对 protogen.Options 的继承
type Options struct {
	protogen.Options
}

func NewCommandLineOpts(set func(name, value string) error) *Options {
	return &Options{
		Options: protogen.Options{
			ParamFunc: set,
		},
	}
}

// Run executes a function as a protoc plugin.
// 调用了我们自己写的 run 方法，使用继承 protogen.Plugin 的 Plugin 结构体指针作为传入参数
//
// It reads a [pluginpb.CodeGeneratorRequest] message from [os.Stdin], invokes the plugin
// function, and writes a [pluginpb.CodeGeneratorResponse] message to [os.Stdout].
//
// If a failure occurs while reading or writing, Run prints an error to
// [os.Stderr] and calls [os.Exit](1).
func (opts Options) Run(f func(*Plugin) error) {
	if err := run(opts, f); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), err)
		os.Exit(1)
	}
}

func run(opts Options, f func(*Plugin) error) error {
	if len(os.Args) > 1 {
		return fmt.Errorf("unknown argument %q (this program should be run by protoc, not directly)", os.Args[1])
	}
	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(in, req); err != nil {
		return err
	}
	gen, err := opts.New(req)
	// 将 gen 包裹在 ourGen 里，目的是用到被重载的 Plugin.Response() 方法
	ourGen := &Plugin{
		Plugin: *gen,
	}
	if err != nil {
		return err
	}
	if err = f(ourGen); err != nil {
		// Errors from the plugin function are reported by setting the
		// error field in the CodeGeneratorResponse.
		//
		// In contrast, errors that indicate a problem in protoc
		// itself (unparsable input, I/O errors, etc.) are reported
		// to stderr.
		gen.Error(err)
	}
	// 使用我们自己的方法
	resp := ourGen.FormattedResponse()
	out, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err = os.Stdout.Write(out); err != nil {
		return err
	}
	return nil
}

// Plugin 是对 protogen.Plugin 的继承
type Plugin struct {
	protogen.Plugin
}

// FormattedResponse 调用了 protogen.Plugin 的 Response 方法
// 然后对其中的每一个文件的内容，进行 goimports 处理，主要目的是优化 imports ，去掉不必要的 imports
func (gen *Plugin) FormattedResponse() *pluginpb.CodeGeneratorResponse {
	resp := gen.Plugin.Response()
	for _, f := range resp.File {
		content := *f.Content
		//fmt.Fprint(os.Stderr, content)
		output, err := imports.Process("", []byte(content), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "goimports content err: %v\n", err)
			return resp
		}
		//fmt.Fprintf(os.Stderr, "goimports content: %v\n", string(output))
		formattedContent := string(output)
		f.Content = &formattedContent
	}
	return resp
}
