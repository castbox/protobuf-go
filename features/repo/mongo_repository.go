package repo

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/castbox/go-guru/pkg/goguru/orm"
	"github.com/planetscale/vtprotobuf/generator"
	"github.com/samber/lo"
	"github.com/srikrsna/protoc-gen-gotag/tagger"
	"github.com/wesleywu/gcontainer/g"
	"github.com/wesleywu/gcontainer/utils/gstr"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func init() {
	generator.RegisterFeature("repository", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
		return &repository{GeneratedFile: gen}
	})
}

type repository struct {
	*generator.GeneratedFile
	once bool
}

var _ generator.FeatureGenerator = (*repository)(nil)

var (
	repoMethodNames = g.NewHashSetFrom[string](
		[]string{"Count", "List", "One", "Get", "Create", "Update", "Upsert", "Delete", "DeleteMulti"},
	)
)

//go:embed repo_template.tpl
var serviceTemplate string

type collectionInfo struct {
	MessageNameVar         string
	CollectionNameVar      string
	CollectionColumnMapVar string
	Indices                []*index
}

type index struct {
	Fields []*indexField
	Unique bool
}

type indexField struct {
	Field string
	Type  string
}

type repositoryValues struct {
	CollectionInfo  *collectionInfo
	RepoServiceName string
	MethodMap       map[string]repositoryMethodParams
}

type repositoryMethodParams struct {
	InputMethodName  string
	OutputMethodName string
}

func (p *repository) GenerateFile(file *protogen.File) bool {
	var (
		info *collectionInfo
	)
	for _, message := range file.Messages {
		info = p.message(message)
		if info != nil {
			break
		}
	}
	if info == nil {
		return p.once
	}

	//p.Import("context")
	//p.Import("github.com/castbox/go-guru/pkg/infra/mongodb")
	//p.Import("github.com/castbox/go-guru/pkg/infra/mongodb/repo/logic")
	//p.Import("github.com/go-kratos/kratos/v2/log")
	//p.Import("go.mongodb.org/mongo-driver/mongo")
	//p.Import("go.mongodb.org/mongo-driver/mongo/options")

	for _, service := range file.Services {
		p.service(service, info)
	}
	return p.once
}

func (p *repository) message(message *protogen.Message) *collectionInfo {
	for _, nested := range message.Messages {
		p.message(nested)
	}

	p.once = true
	repositorySetting := p.getRepositorySetting(message)

	// 只对标注了 table_name 的 message 生成其 persistence
	if repositorySetting == nil {
		return nil
	}

	ccTypeName := message.GoIdent

	p.P("const (")
	p.P(ccTypeName, "CollectionName = \"", repositorySetting.CollectionName, "\"")
	p.P(")")
	p.P()

	// 生成 ColumnMap
	p.P("var ", ccTypeName, "ColumnMap = map[string]string{")
	for _, field := range message.Fields {
		fieldName := field.GoName

		tagsExt := proto.GetExtension(field.Desc.Options(), tagger.E_Tags)

		bsonFieldName := ""
		if tags, ok := tagsExt.(string); ok {
			// 有 (tagger.tags) = "xxxx" 的注解，尝试从其中解析 key 为 bson 的 value
			bsonFieldName = getBsonFieldName(tags)
		}
		if bsonFieldName == "" {
			bsonFieldName = gstr.CaseSnake(fieldName)
		}
		p.P(fmt.Sprintf("\"%s\": \"%s\",", fieldName, bsonFieldName))
	}
	p.P("}")
	p.P()
	return &collectionInfo{
		MessageNameVar:         message.GoIdent.GoName,
		CollectionNameVar:      ccTypeName.GoName + "CollectionName",
		CollectionColumnMapVar: ccTypeName.GoName + "ColumnMap",
		Indices:                lo.Map(repositorySetting.Indices, transformIndex),
	}
}

func (p *repository) getRepositorySetting(message *protogen.Message) *orm.RepositorySetting {
	if message == nil {
		return nil
	}

	ext := proto.GetExtension(message.Desc.Options(), orm.E_Repository)
	if repositorySetting, ok := ext.(*orm.RepositorySetting); ok {
		return repositorySetting
	}
	return nil
}

func getBsonFieldName(tagsString string) string {
	tags := strings.Split(tagsString, ",")
	for _, tag := range tags {
		if strings.HasPrefix(tag, "bson:") {
			bsonValue := strings.TrimPrefix(tag, "bson:")
			bsonValue = strings.TrimSpace(bsonValue)
			bsonValue = strings.TrimPrefix(bsonValue, "\"")
			bsonValue = strings.TrimSuffix(bsonValue, "\"")
			values := strings.Split(bsonValue, ",")
			return values[0]
		}
	}
	return ""
}

func transformIndex(i *orm.Index, _ int) *index {
	result := &index{
		Fields: make([]*indexField, len(i.Fields)),
	}
	for ii, f := range i.Fields {
		typeStr := ""
		switch f.Type {
		case orm.IndexType_AscIndex:
			typeStr = "1"
		case orm.IndexType_DescIndex:
			typeStr = "-1"
		case orm.IndexType_TextIndex:
			typeStr = "\"text\""
		case orm.IndexType_HashedIndex:
			typeStr = "\"hashed\""
		}
		result.Fields[ii] = &indexField{
			Field: f.Field,
			Type:  typeStr,
		}
	}
	return result
}

func (p *repository) service(service *protogen.Service, info *collectionInfo) {
	p.once = true

	values := &repositoryValues{
		CollectionInfo:  info,
		RepoServiceName: service.GoName,
		MethodMap:       make(map[string]repositoryMethodParams),
	}
	for _, method := range service.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}
		methodName := method.GoName

		if repoMethodNames.Contains(methodName) {
			values.MethodMap[methodName] = repositoryMethodParams{
				InputMethodName:  method.Input.GoIdent.GoName,
				OutputMethodName: method.Output.GoIdent.GoName,
			}
		}
	}
	if len(values.MethodMap) != 0 {
		templateStr, err := p.execute(values)
		if err != nil {
			panic(err)
		}
		p.P(templateStr)
	}
}

func (g *repository) execute(values *repositoryValues) (string, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("service").Parse(strings.TrimSpace(serviceTemplate))
	if err != nil {
		panic(err)
	}
	if err = tmpl.Execute(buf, values); err != nil {
		return "", err
	}
	return buf.String(), nil
}
