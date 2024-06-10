package pool

import (
	"github.com/planetscale/vtprotobuf/generator"
	"google.golang.org/protobuf/compiler/protogen"
)

func init() {
	generator.RegisterFeature("setter", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
		return &setter{GeneratedFile: gen}
	})
}

type setter struct {
	*generator.GeneratedFile
	once bool
}

var _ generator.FeatureGenerator = (*setter)(nil)

func (p *setter) GenerateFile(file *protogen.File) bool {
	for _, message := range file.Messages {
		p.message(message)
	}
	return p.once
}

func (p *setter) message(message *protogen.Message) {
	for _, nested := range message.Messages {
		p.message(nested)
	}

	p.once = true
	ccTypeName := message.GoIdent

	for _, field := range message.Fields {
		fieldName := field.GoName

		goType, pointer := p.FieldGoType(field)
		if pointer {
			goType = "*" + goType
		}
		// Generate a naive setter.
		// I don't pretend to understand everything I am eliding by this simplicity.
		p.P("func (m *", ccTypeName, ") Set"+fieldName+"(v "+goType+") {")
		p.P("m." + fieldName + " = v")
		p.P("}")
		p.P()
	}
}
