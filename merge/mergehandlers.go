package merge

import (
	"bytes"
	"fmt"
	"github.com/luk7152/zcompact/constaint"
	"github.com/luk7152/zcompact/utils"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"

	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"go.etcd.io/etcd/pkg/fileutil"
)

type Handler struct {
	HandlerName  string
	RequestType  string
	LogicType    string
	Call         string
	HasResp      bool
	HasRequest   bool
	LogicPackage string
}

func gen(folder string, group spec.Group, dir, nameStyle string) error {
	var routes = group.Routes
	parentPkg, err := utils.GetParentPackage(dir)
	if err != nil {
		return err
	}

	filename, err := format.FileNamingFormat(nameStyle, "handlers")
	if err != nil {
		return err
	}

	filename = filename + ".go"
	filename = filepath.Join(dir, folder, filename)
	var fp *os.File

	var hasExist = true
	if !fileutil.Exist(filename) {
		_, err = os.Create(filename)
		if err != nil {
			return err
		}
		hasExist = false
	}

	fp, err = os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()

	text, err := pathx.LoadTemplate(constaint.Category, constaint.HandlerTemplateFile, constaint.HandlerTemplate)
	if err != nil {
		return err
	}

	var funcs []string
	var imports []string
	for _, route := range routes {
		var handler = GetHandlerName(route, folder)
		if hasExist && utils.FuncExist(filename, handler) {
			continue
		}
		handleObj := Handler{
			HandlerName:  handler,
			RequestType:  strings.Title(route.RequestTypeName()),
			LogicType:    strings.Title(GetLogicName(route)),
			Call:         strings.Title(strings.TrimSuffix(handler, "Handler")),
			HasResp:      len(route.ResponseTypeName()) > 0,
			HasRequest:   len(route.RequestTypeName()) > 0,
			LogicPackage: findLogicPackage(group, route),
		}

		buffer := new(bytes.Buffer)
		err = template.Must(template.New("handlerTemplate").Parse(text)).Execute(buffer, handleObj)
		if err != nil {
			return err
		}

		funcs = append(funcs, buffer.String())

		for _, item := range genHandlerImports(group, route, parentPkg) {
			if !stringx.Contains(imports, item) {
				imports = append(imports, item)
			}
		}
	}

	buffer := new(bytes.Buffer)
	if !hasExist {
		importsStr := strings.Join(imports, "\n\t")
		err = template.Must(template.New("handlerImports").Parse(constaint.HandlerImports)).Execute(buffer, map[string]string{
			"packages": importsStr,
		})
		if err != nil {
			return err
		}
	}

	formatCode := utils.FormatCode(strings.ReplaceAll(buffer.String(), "&#34;", "\"") + strings.Join(funcs, "\n\n"))
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if len(content) > 0 {
		formatCode = string(content) + "\n" + formatCode
	}
	_, err = fp.WriteString(formatCode)
	return err
}

func MergeHandler(folder, handlerName string, group spec.Group, dir, nameStyle string) error {
	var routes = group.Routes
	parentPkg, err := utils.GetParentPackage(dir)
	if err != nil {
		return err
	}

	filename, err := format.FileNamingFormat(nameStyle, handlerName+"handlers")
	if err != nil {
		return err
	}

	filename = filename + ".go"
	filename = filepath.Join(dir, folder, filename)
	var fp *os.File

	var hasExist = true
	if !fileutil.Exist(filename) {
		_, err = os.Create(filename)
		if err != nil {
			return err
		}
		hasExist = false
	}

	fp, err = os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer fp.Close()

	text, err := pathx.LoadTemplate(constaint.Category, constaint.HandlerTemplateFile, constaint.HandlerTemplate)
	if err != nil {
		return err
	}

	var funcs []string
	var imports []string
	for _, route := range routes {
		var handler = GetHandlerName(route, folder)
		if hasExist && utils.FuncExist(filename, handler) {
			continue
		}
		handleObj := Handler{
			HandlerName:  handler,
			RequestType:  strings.Title(route.RequestTypeName()),
			LogicType:    strings.Title(getLogicNameNew(handlerName)),
			Call:         strings.Title(strings.TrimSuffix(handler, "Handler")),
			HasResp:      len(route.ResponseTypeName()) > 0,
			HasRequest:   len(route.RequestTypeName()) > 0,
			LogicPackage: findLogicPackageNew(group, route),
		}

		buffer := new(bytes.Buffer)
		err = template.Must(template.New("handlerTemplate").Parse(text)).Execute(buffer, handleObj)
		if err != nil {
			return err
		}

		funcs = append(funcs, buffer.String())

		for _, item := range genHandlerImports(group, route, parentPkg) {
			if !stringx.Contains(imports, item) {
				imports = append(imports, item)
			}
		}
	}

	buffer := new(bytes.Buffer)
	if !hasExist {
		importsStr := strings.Join(imports, "\n\t")
		err = template.Must(template.New("handlerImports").Parse(constaint.HandlerImports)).Execute(buffer, map[string]string{
			"packages": importsStr,
		})
		if err != nil {
			return err
		}
	}

	formatCode := utils.FormatCode(strings.ReplaceAll(buffer.String(), "&#34;", "\"") + strings.Join(funcs, "\n\n"))
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if len(content) > 0 {
		formatCode = string(content) + "\n" + formatCode
	}
	_, err = fp.WriteString(formatCode)
	return err
}

func GetHandlerName(route spec.Route, folder string) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	handler = handler + "Handler"
	if folder != constaint.HandlerDir {
		handler = strings.Title(handler)
	}
	return handler
}

func GetHandlerFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(constaint.GroupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(constaint.GroupProperty)
		if len(folder) == 0 {
			return constaint.HandlerDir
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(constaint.HandlerDir, folder)
}

func GetHandlerFolderPathNew(group spec.Group, route spec.Route) (string, string) {
	folder := route.GetAnnotation(constaint.GroupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(constaint.GroupProperty)
		if len(folder) == 0 {
			return constaint.HandlerDir, folder + ""
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")

	fds := strings.Split(folder, "/")
	p, n := constaint.HandlerDir, ""
	n = fds[len(fds)-1]

	for _, v := range fds[:len(fds)-1] {
		p = path.Join(p, v)
	}

	return p, n
}

func getHandlerBaseName(route spec.Route) (string, error) {
	handler := route.Handler
	handler = strings.TrimSpace(handler)
	handler = strings.TrimSuffix(handler, "handler")
	handler = strings.TrimSuffix(handler, "Handler")
	return handler, nil
}

func genHandlerImports(group spec.Group, route spec.Route, parentPkg string) []string {
	logicPath, _ := getLogicFolderPathNew(group, route)
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"",
		utils.JoinPackages(parentPkg, logicPath)))
	imports = append(imports, fmt.Sprintf("\"%s\"", utils.JoinPackages(parentPkg, constaint.ContextDir)))
	if len(route.RequestTypeName()) > 0 {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", utils.JoinPackages(parentPkg, constaint.TypesDir)))
	}
	imports = append(imports, fmt.Sprintf("\"%s/rest/httpx\"", constaint.ProjectOpenSourceUrl))

	return imports
}
