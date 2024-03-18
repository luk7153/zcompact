package merge

import (
	"bytes"
	"fmt"
	"github.com/luk7152/zcompact/constaint"
	"github.com/luk7152/zcompact/utils"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"go.etcd.io/etcd/pkg/fileutil"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Logic struct {
	LogicType    string
	FunctionName string
	Request      string
	ResponseType string
	ReturnString string
}

func MergeLogic(fd, handlerName string, group spec.Group, dir, nameStyle string) error {
	fds := strings.Split(fd, "/")
	folder := ""
	for idx, v := range fds {
		if idx == 0 {
			folder = path.Join(folder, constaint.LogicDir)
		} else if idx == 1 {
			continue
		} else {
			folder = path.Join(folder, v)
		}
	}

	//group下的所有routes
	var routes = group.Routes
	parentPkg, err := utils.GetParentPackage(dir)
	if err != nil {
		return err
	}

	filename, err := format.FileNamingFormat(nameStyle, handlerName+"logics")
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

	text, err := pathx.LoadTemplate(constaint.Category, constaint.LogicTemplateFile, constaint.LogicTemplate)
	if err != nil {
		return err
	}

	var funcs []string
	var imports []string
	//处理每一个route
	var logic = getLogicNameNew(handlerName)
	for _, route := range routes {
		fun, _ := getHandlerBaseName(route)
		if hasExist && utils.FuncExist(filename, fun) {
			continue
		}
		var responseString string
		var returnString string
		var requestString string
		if len(route.ResponseTypeName()) > 0 {
			resp := utils.ResponseGoTypeName(route, constaint.TypesPacket)
			responseString = "(resp " + resp + ", err error)"
			returnString = "return"
		} else {
			responseString = "error"
			returnString = "return nil"
		}
		if len(route.RequestTypeName()) > 0 {
			requestString = "req *" + utils.RequestGoTypeName(route, constaint.TypesPacket)
		}

		logicObj := Logic{
			LogicType:    logic,
			FunctionName: fun,
			Request:      requestString,
			ResponseType: responseString,
			ReturnString: returnString,
		}

		buffer := new(bytes.Buffer)
		err = template.Must(template.New("logicTemplate").Parse(text)).Execute(buffer, logicObj)
		if err != nil {
			return err
		}

		funcs = append(funcs, buffer.String())

		for _, item := range genLogicImports(route, parentPkg) {
			if !stringx.Contains(imports, item) {
				imports = append(imports, item)
			}
		}
	}

	buffer := new(bytes.Buffer)
	if !hasExist {
		importsStr := strings.Join(imports, "\n\t")
		err = template.Must(template.New("logicImports").Parse(constaint.LogicStruct)).Execute(buffer, map[string]string{
			"imports": importsStr,
			"logic":   logic,
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

func findLogicPackage(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(constaint.GroupProperty)
	if len(folder) > 0 {
		parts := strings.Split(folder, "/")
		return parts[len(parts)-1]
	}

	folder = group.GetAnnotation(constaint.GroupProperty)
	if len(folder) > 0 {
		parts := strings.Split(folder, "/")
		return parts[len(parts)-1]
	}

	return "logic"
}

func findLogicPackageNew(group spec.Group, route spec.Route) string {
	//folder := route.GetAnnotation(constaint.GroupProperty)
	//if len(folder) > 0 {
	//	parts := strings.Split(folder, "/")
	//	return parts[len(parts)-1]
	//}
	//
	//folder = group.GetAnnotation(constaint.GroupProperty)
	//if len(folder) > 0 {
	//	parts := strings.Split(folder, "/")
	//	return parts[len(parts)-1]
	//}

	return "logic"
}

func GetLogicName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Logic"
}

func getLogicNameNew(handler string) string {
	//handler, err := getHandlerBaseName(route)
	//if err != nil {
	//	panic(err)
	//}

	return strings.Title(strings.TrimSpace(handler)) + "Logic"
}

func genLogicImports(route spec.Route, parentPkg string) []string {
	var imports []string
	imports = append(imports, "\"context\"")
	imports = append(imports, fmt.Sprintf("\"%s\"", utils.JoinPackages(parentPkg, constaint.ContextDir)))
	if len(route.RequestTypeName()) > 0 {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", utils.JoinPackages(parentPkg, constaint.TypesDir)))
	}
	imports = append(imports, fmt.Sprintf("\"%s/core/logx\"", constaint.ProjectOpenSourceUrl))

	return imports
}

func GetLogicFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(constaint.GroupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(constaint.GroupProperty)
		if len(folder) == 0 {
			return constaint.LogicDir
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(constaint.LogicDir, folder)
}

func getLogicFolderPathNew(group spec.Group, route spec.Route) (string, string) {
	folder := route.GetAnnotation(constaint.GroupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(constaint.GroupProperty)
		if len(folder) == 0 {
			return constaint.LogicDir, ""
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")

	fds := strings.Split(folder, "/")
	p, n := constaint.LogicDir, ""
	n = fds[len(fds)-1]

	for _, v := range fds[:len(fds)-1] {
		p = path.Join(p, v)
	}
	return p, n
}
