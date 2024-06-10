{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

func Register{{.ServiceType}}GuruServer(srv {{.ServiceType}}Server) error {
	var err error
	rpcinfo.RegisterServiceDesc(&{{$svrType}}_ServiceDesc, srv)
	{{- range .Methods}} {{ if ne .CacheRule nil }}
    err = rpcinfo.RegisterMethodDesc("{{$svrName}}", "{{.Name}}", &cache.CacheRule{
            Cachable: {{.CacheRule.Cachable}},
            Name:     "{{.CacheRule.Name}}",
            Ttl:      "{{.CacheRule.Ttl}}",
            Key:      "{{.CacheRule.Key}}",
    }) {{ else }}
    err = rpcinfo.RegisterMethodDesc("{{$svrName}}", "{{.Name}}", nil) {{ end }}
    if err != nil {
        return err
    }
	{{- end}}
	return nil
}
