// Copyright 2013 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"text/template"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

type netrpcPlugin struct{ *generator.Generator }

func (p *netrpcPlugin) Name() string                { return "netrpc" }
func (p *netrpcPlugin) Init(g *generator.Generator) { p.Generator = g }

func (p *netrpcPlugin) GenerateImports(file *generator.FileDescriptor) {
	if len(file.Service) > 0 {
		p.P(`import "net/rpc"`)
	}
}

func (p *netrpcPlugin) Generate(file *generator.FileDescriptor) {
	for _, svc := range file.Service {
		p.genServiceInterface(file, svc)
		p.genServiceServer(file, svc)
		p.genServiceClient(file, svc)
	}
}

func (p *netrpcPlugin) genServiceInterface(
	file *generator.FileDescriptor,
	svc *descriptor.ServiceDescriptorProto,
) {
	const serviceInterfaceTmpl = `
type {{.ServiceName}}Interface interface {
	{{.CallMethodList}}
}
`
	const callMethodTmpl = `
{{.MethodName}}(in {{.ArgsType}}, out *{{.ReplyType}}) error`

	// gen call method list
	var callMethodList string
	for _, m := range svc.Method {
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(callMethodTmpl))
		t.Execute(out, &struct{ ServiceName, MethodName, ArgsType, ReplyType string }{
			ServiceName: generator.CamelCase(svc.GetName()),
			MethodName:  generator.CamelCase(m.GetName()),
			ArgsType:    p.TypeName(p.ObjectNamed(m.GetInputType())),
			ReplyType:   p.TypeName(p.ObjectNamed(m.GetOutputType())),
		})
		callMethodList += out.String()

		p.RecordTypeUse(m.GetInputType())
		p.RecordTypeUse(m.GetOutputType())
	}

	// gen all interface code
	{
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(serviceInterfaceTmpl))
		t.Execute(out, &struct{ ServiceName, CallMethodList string }{
			ServiceName:    generator.CamelCase(svc.GetName()),
			CallMethodList: callMethodList,
		})
		p.P(out.String())
	}
}

func (p *netrpcPlugin) genServiceServer(
	file *generator.FileDescriptor,
	svc *descriptor.ServiceDescriptorProto,
) {
	const serviceHelperFunTmpl = `
func Register{{.ServiceName}}(srv *rpc.Server, x {{.ServiceName}}) error {
	if err := srv.RegisterName("{{.ServiceName}}", x); err != nil {
		return err
	}
	return nil
}
`
	{
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(serviceHelperFunTmpl))
		t.Execute(out, &struct{ PackageName, ServiceName, ServiceRegisterName string }{
			PackageName: file.GetPackage(),
			ServiceName: generator.CamelCase(svc.GetName()),
		})
		p.P(out.String())
	}
}

func (p *netrpcPlugin) genServiceClient(
	file *generator.FileDescriptor,
	svc *descriptor.ServiceDescriptorProto,
) {
	const clientHelperFuncTmpl = `
type {{.ServiceName}}Client struct {
	*rpc.Client
}

var _ {{.ServiceName}}Interface = (*{{.ServiceName}}Client)(nil)

func Dial{{.ServiceName}}(network, address string) (*{{.ServiceName}}Client, error) {
	c, err := rpc.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &{{.ServiceName}}Client{Client: c}, nil
}

{{.MethodList}}
`
	const clientMethodTmpl = `
func (p *{{.ServiceName}}Client) {{.MethodName}}(in {{.ArgsType}}, out *{{.ReplyType}}) error {
	return p.Client.Call("{{.ServiceName}}.{{.MethodName}}", in, out)
}
`

	// gen client method list
	var methodList string
	for _, m := range svc.Method {
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(clientMethodTmpl))
		t.Execute(out, &struct{ ServiceName, ServiceRegisterName, MethodName, ArgsType, ReplyType string }{
			ServiceName:         generator.CamelCase(svc.GetName()),
			ServiceRegisterName: file.GetPackage() + "." + generator.CamelCase(svc.GetName()),
			MethodName:          generator.CamelCase(m.GetName()),
			ArgsType:            p.TypeName(p.ObjectNamed(m.GetInputType())),
			ReplyType:           p.TypeName(p.ObjectNamed(m.GetOutputType())),
		})
		methodList += out.String()
	}

	// gen all client code
	{
		out := bytes.NewBuffer([]byte{})
		t := template.Must(template.New("").Parse(clientHelperFuncTmpl))
		t.Execute(out, &struct{ PackageName, ServiceName, MethodList string }{
			PackageName: file.GetPackage(),
			ServiceName: generator.CamelCase(svc.GetName()),
			MethodList:  methodList,
		})
		p.P(out.String())
	}
}

func init() {
	generator.RegisterPlugin(new(netrpcPlugin))
}
