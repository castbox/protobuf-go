{{$RepoServiceName := .RepoServiceName}}
{{$MessageName := .CollectionInfo.MessageNameVar}}
{{$CollectionName := .CollectionInfo.CollectionNameVar}}
{{$CollectionColumnMap := .CollectionInfo.CollectionColumnMapVar}}
{{$Indices := .CollectionInfo.Indices}}

type {{$RepoServiceName}} struct {
	collection       *mongo.Collection
	useIdObfuscating bool
	helper           *log.Helper
}

func New{{$RepoServiceName}}(mongoClient *mongodb.Client, logHelper *log.Helper) (*{{$RepoServiceName}}, error) {
	collection := mongoClient.Collection({{$CollectionName}})
	repo := &{{$RepoServiceName}}{
		collection:       collection,
		useIdObfuscating: mongoClient.UseIdObfuscating(),
		helper:           logHelper,
	}
	err := repo.createIndexes(context.Background())
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// createIndexes 创建索引
func (r *{{$RepoServiceName}}) createIndexes(ctx context.Context) error {
    {{if $Indices }}_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
        {{range $index := $Indices}}{{if $index.Fields}}mongo.IndexModel{
            Keys: bson.D{ {{range $field := $index.Fields}}
            {"{{$field.Field}}", {{$field.Type}}},
            {{end}}
            },
            Options: &options.IndexOptions{
                Unique: types.Wrap({{$index.Unique}}),
            },
        },
        {{end}}{{end}}
    })
    if err != nil {
        return err
    }
    {{end}}return nil
}

{{with .MethodMap.Count}}// Count 根据req指定的查询条件获取记录列表
// 支持翻页和排序参数，支持查询条件参数类型自动转换
// 未赋值或或赋值为nil的字段不参与条件查询
func (r *{{$RepoServiceName}}) Count(ctx context.Context, req *{{.InputMethodName}}) (*{{.OutputMethodName}}, error) {
	l := logic.NewCountLogic[*{{.InputMethodName}}, *{{.OutputMethodName}}](r.collection, {{$CollectionColumnMap}}, r.helper)
	return l.Count(ctx, req)
}
{{end}}
{{with .MethodMap.List}}// List 根据req指定的查询条件获取记录列表
// 支持翻页和排序参数，支持查询条件参数类型自动转换
// 未赋值或或赋值为nil的字段不参与条件查询
func (r *{{$RepoServiceName}}) List(ctx context.Context, req *{{.InputMethodName}}) (*{{.OutputMethodName}}, error) {
	l := logic.NewListLogic[*{{.InputMethodName}}, *{{.OutputMethodName}}, *{{$MessageName}}](r.collection, {{$CollectionColumnMap}}, r.helper)
	return l.List(ctx, req)
}
{{end}}
{{with .MethodMap.One}}// One 根据req指定的查询条件获取单条数据
// 支持排序参数，支持查询条件参数类型自动转换
// 未赋值或或赋值为nil的字段不参与条件查询
func (r *{{$RepoServiceName}}) One(ctx context.Context, req *{{.InputMethodName}}) (*{{.OutputMethodName}}, error) {
	l := logic.NewOneLogic[*{{.InputMethodName}}, *{{.OutputMethodName}}, *{{$MessageName}}](r.collection, {{$CollectionColumnMap}}, r.helper)
	return l.One(ctx, req)
}
{{end}}
{{with .MethodMap.Get}}// Get 根据主键/ID查询特定记录
func (r *{{$RepoServiceName}}) Get(ctx context.Context, req *{{.InputMethodName}}) (*{{.OutputMethodName}}, error) {
	l := logic.NewGetLogic[*{{.InputMethodName}}, *{{.OutputMethodName}}](r.collection, r.useIdObfuscating, r.helper)
	return l.Get(ctx, req)
}
{{end}}
{{with .MethodMap.Create}}// Create 插入记录
// 包括表中所有字段，支持字段类型自动转换，支持对非主键且可为空字段不赋值
// 未赋值或赋值为nil的字段将被更新为 NULL 或数据库表指定的DEFAULT
func (r *{{$RepoServiceName}}) Create(ctx context.Context, req *{{.InputMethodName}}) (*{{.OutputMethodName}}, error) {
	l := logic.NewCreateLogic[*{{.InputMethodName}}, *{{.OutputMethodName}}](r.collection, r.useIdObfuscating, r.helper)
	return l.Create(ctx, req)
}
{{end}}
{{with .MethodMap.Update}}// Update 根据主键更新对应记录
// 支持字段类型自动转换，支持对非主键字段赋值/不赋值
// 未赋值或赋值为nil的字段不参与更新（即不会修改原记录的字段值）
func (r *{{$RepoServiceName}}) Update(ctx context.Context, req *{{.InputMethodName}}) (*{{.OutputMethodName}}, error) {
	l := logic.NewUpdateLogic[*{{.InputMethodName}}, *{{.OutputMethodName}}](r.collection, r.useIdObfuscating, r.helper)
	return l.Update(ctx, req)
}
{{end}}
{{with .MethodMap.Upsert}}// Upsert 根据主键（或唯一索引）是否存在且已在req中赋值，更新或插入对应记录。
// 支持字段类型自动转换，支持对非主键字段赋值/不赋值
// 未赋值或赋值为nil的字段不参与更新/插入（即更新时不会修改原记录的字段值）
func (r *{{$RepoServiceName}}) Upsert(ctx context.Context, req *{{.InputMethodName}}) (*{{.OutputMethodName}}, error) {
	l := logic.NewUpsertLogic[*{{.InputMethodName}}, *{{.OutputMethodName}}](r.collection, r.useIdObfuscating, r.helper)
	return l.Upsert(ctx, req)
}
{{end}}
{{with .MethodMap.Delete}}// Delete 根据主键删除对应记录
func (r *{{$RepoServiceName}}) Delete(ctx context.Context, req *{{.InputMethodName}}) (*{{.OutputMethodName}}, error) {
	l := logic.NewDeleteLogic[*{{.InputMethodName}}, *{{.OutputMethodName}}](r.collection, r.useIdObfuscating, r.helper)
	return l.Delete(ctx, req)
}
{{end}}
{{with .MethodMap.DeleteMulti}}// DeleteMulti 根据req指定的条件删除表中记录（可能多条）
// 未赋值或或赋值为nil的字段不参与条件查询
func (r *{{$RepoServiceName}}) DeleteMulti(ctx context.Context, req *{{.InputMethodName}}) (*{{.OutputMethodName}}, error) {
	l := logic.NewDeleteMultiLogic[*{{.InputMethodName}}, *{{.OutputMethodName}}](r.collection, {{$CollectionColumnMap}}, r.helper)
	return l.DeleteMulti(ctx, req)
}
{{end}}

