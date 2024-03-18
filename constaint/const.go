package constaint

const (
	Category                   = "api"
	HandlerTemplateFile        = "handler.tpl"
	LogicTemplateFile          = "logic.tpl"
	RoutesTemplateFile         = "routes.tpl"
	RoutesAdditionTemplateFile = "route-addition.tpl"

	GroupProperty        = "group"
	Interval             = "internal/"
	HandlerDir           = Interval + "handler"
	LogicDir             = Interval + "logic"
	TypesPacket          = "types"
	TypesDir             = Interval + TypesPacket
	ContextDir           = Interval + "svc"
	PkgSep               = "/"
	ProjectOpenSourceUrl = "github.com/zeromicro/go-zero"
	HandlerImports       = `package handler

import (
	"net/http"

	{{.packages}}
)
`
	HandlerTemplate = `func {{.HandlerName}}(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}{{end}}

		l := {{.LogicPackage}}.New{{.LogicType}}(r.Context(), ctx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
		if err != nil {
			httpx.Error(w, err)
		} else {
			{{if .HasResp}}httpx.OkJson(w, resp){{else}}httpx.Ok(w){{end}}
		}
	}
}`

	LogicStruct = `package logic

import (
	{{.imports}}
)

type {{.logic}} struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func New{{.logic}}(ctx context.Context, svcCtx *svc.ServiceContext) *{{.logic}} {
	return &{{.logic}}{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
`
	LogicTemplate = `func (l *{{.LogicType}}) {{.FunctionName}}({{.Request}}) {{.ResponseType}} {
	// todo: add your logic here and delete this line

	{{.ReturnString}}
}`
)
