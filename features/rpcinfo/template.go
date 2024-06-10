package rpcinfo

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"github.com/castbox/go-guru/pkg/goguru/cache"
)

//go:embed service_template.tpl
var serviceTemplate string

type serviceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	//Metadata    string // api/helloworld/helloworld.proto
	Methods    []*methodDesc
	MethodSets map[string]*methodDesc
}

type methodDesc struct {
	// method
	Name         string
	OriginalName string // The parsed original name
	Num          int
	Request      string
	Reply        string
	Comment      string
	CacheRule    *cache.CacheRule
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("service").Parse(strings.TrimSpace(serviceTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
