package rpcinfo

import (
	"strings"

	"github.com/castbox/go-guru/pkg/goguru/cache"
	"github.com/planetscale/vtprotobuf/generator"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	generator.RegisterFeature("rpcinfo", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
		return &rpcinfo{GeneratedFile: gen}
	})
}

const deprecationComment = "// Deprecated: Do not use."

var methodSets = make(map[string]int)

type rpcinfo struct {
	*generator.GeneratedFile
	once bool
}

var _ generator.FeatureGenerator = (*rpcinfo)(nil)

func (g *rpcinfo) GenerateFile(file *protogen.File) bool {
	if len(file.Services) == 0 {
		return true
	}
	for _, service := range file.Services {
		g.service(service)
	}
	return g.once
}

func (g *rpcinfo) service(service *protogen.Service) {
	g.once = true
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	// HTTP Server.
	sd := &serviceDesc{
		ServiceType: service.GoName,
		ServiceName: string(service.Desc.FullName()),
		//Metadata:    file.Desc.Path(),
	}
	for _, method := range service.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}
		rule := proto.GetExtension(method.Desc.Options(), cache.E_Rule).(*cache.CacheRule)
		sd.Methods = append(sd.Methods, g.buildCacheRule(method, rule))
	}
	if len(sd.Methods) != 0 {
		g.P(sd.execute())
	}
}

func (g *rpcinfo) buildCacheRule(m *protogen.Method, rule *cache.CacheRule) *methodDesc {
	md := g.buildMethodDesc(m, rule)
	return md
}

func (g *rpcinfo) buildMethodDesc(m *protogen.Method, rule *cache.CacheRule) *methodDesc {
	defer func() { methodSets[m.GoName]++ }()

	comment := m.Comments.Leading.String() + m.Comments.Trailing.String()
	if comment != "" {
		comment = "// " + m.GoName + strings.TrimPrefix(strings.TrimSuffix(comment, "\n"), "//")
	}
	return &methodDesc{
		Name:         m.GoName,
		OriginalName: string(m.Desc.Name()),
		Num:          methodSets[m.GoName],
		Request:      g.QualifiedGoIdent(m.Input.GoIdent),
		Reply:        g.QualifiedGoIdent(m.Output.GoIdent),
		Comment:      comment,
		CacheRule:    rule,
	}
}
