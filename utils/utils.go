package utils

import (
	"bytes"
	"fmt"
	"github.com/luk7153/zcompact/constaint"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/api/util"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/golang"
	"github.com/zeromicro/go-zero/tools/goctl/util/ctx"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"go/ast"
	goformat "go/format"
	"go/parser"
	"go/token"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func FuncExist(filename, funcName string) bool {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return false
	}

	set := token.NewFileSet()
	packs, err := parser.ParseFile(set, filename, string(data), parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, d := range packs.Decls {
		if fn, isFn := d.(*ast.FuncDecl); isFn {
			if fn.Name.String() == funcName {
				return true
			}
		}
	}
	return false
}

func FormatCode(code string) string {
	ret, err := goformat.Source([]byte(code))
	if err != nil {
		return code
	}

	return string(ret)
}
func JoinPackages(pkgs ...string) string {
	return strings.Join(pkgs, constaint.PkgSep)
}

func GetParentPackage(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	projectCtx, err := ctx.Prepare(abs)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(projectCtx.Path, strings.TrimPrefix(projectCtx.WorkDir, projectCtx.Dir))), nil
}

func ResponseGoTypeName(r spec.Route, pkg ...string) string {
	if r.ResponseType == nil {
		return ""
	}

	resp := GolangExpr(r.ResponseType, pkg...)
	switch r.ResponseType.(type) {
	case spec.DefineStruct:
		if !strings.HasPrefix(resp, "*") {
			return "*" + resp
		}
	}

	return resp
}
func RequestGoTypeName(r spec.Route, pkg ...string) string {
	if r.RequestType == nil {
		return ""
	}

	return GolangExpr(r.RequestType, pkg...)
}

func GolangExpr(ty spec.Type, pkg ...string) string {
	switch v := ty.(type) {
	case spec.PrimitiveType:
		return v.RawName
	case spec.DefineStruct:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}

		if len(pkg) == 0 {
			return v.RawName
		}

		return fmt.Sprintf("%s.%s", pkg[0], strings.Title(v.RawName))
	case spec.ArrayType:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}

		if len(pkg) == 0 {
			return v.RawName
		}

		return fmt.Sprintf("[]%s", GolangExpr(v.Value, pkg...))
	case spec.MapType:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}

		if len(pkg) == 0 {
			return v.RawName
		}

		return fmt.Sprintf("map[%s]%s", v.Key, GolangExpr(v.Value, pkg...))
	case spec.PointerType:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}

		if len(pkg) == 0 {
			return v.RawName
		}

		return fmt.Sprintf("*%s", GolangExpr(v.Type, pkg...))
	case spec.InterfaceType:
		return v.RawName
	}

	return ""
}

type FileGenConfig struct {
	Dir             string
	Subdir          string
	Filename        string
	TemplateName    string
	Category        string
	TemplateFile    string
	BuiltinTemplate string
	Data            any
}

func GenFile(c FileGenConfig) error {
	fp, created, err := util.MaybeCreateFile(c.Dir, c.Subdir, c.Filename)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	var text string
	if len(c.Category) == 0 || len(c.TemplateFile) == 0 {
		text = c.BuiltinTemplate
	} else {
		text, err = pathx.LoadTemplate(c.Category, c.TemplateFile, c.BuiltinTemplate)
		if err != nil {
			return err
		}
	}

	t := template.Must(template.New(c.TemplateName).Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, c.Data)
	if err != nil {
		return err
	}
	code := golang.FormatCode(strings.ReplaceAll(buffer.String(), "&#34;", "\""))
	_, err = fp.WriteString(code)
	return err
}
