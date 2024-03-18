package main

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/luk7152/zcompact/merge"
	"github.com/zeromicro/go-zero/tools/goctl/api/gogen"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/golang"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

func main() {
	plugin, err := plugin.NewPlugin()
	if err != nil {
		panic(err)
	}

	if len(plugin.Style) == 0 {
		plugin.Style = "gozero"
	}
	cfg, err := config.NewConfig(plugin.Style)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	err = gogen.DoGenProject(plugin.ApiFilePath, plugin.Dir, strings.TrimSpace(plugin.Style))
	if err != nil {
		panic(err)
	}

	var api = plugin.Api
	for _, group := range api.Service.Groups {
		if len(group.Routes) == 0 {
			continue
		}

		route0 := group.Routes[0]
		folder, name := merge.GetHandlerFolderPathNew(group, route0)
		for _, route := range group.Routes {
			//删除原有的handler.go和路径
			fileHandler, err1 := format.FileNamingFormat(plugin.Style, merge.GetHandlerName(route, folder))
			if err1 != nil {
				panic(err1)
			}

			fileHandler = fileHandler + ".go"
			os.Remove(filepath.Join(plugin.Dir, merge.GetHandlerFolderPath(group, route), fileHandler))
			os.Remove(filepath.Join(plugin.Dir, merge.GetHandlerFolderPath(group, route)))

			//删除logic.go和路径
			fileLogic, err1 := format.FileNamingFormat(plugin.Style, merge.GetLogicName(route))
			if err1 != nil {
				panic(err1)
			}
			fileLogic = fileLogic + ".go"
			os.Remove(filepath.Join(plugin.Dir, merge.GetLogicFolderPath(group, route), fileLogic))
			os.Remove(filepath.Join(plugin.Dir, merge.GetLogicFolderPath(group, route)))

		}

		err = merge.MergeHandler(folder, name, group, plugin.Dir, plugin.Style)
		if err != nil {
			debug.PrintStack()
			panic(err)
		}
		fmt.Println(strings.TrimSpace((name + " handlers merged!")))
		err = merge.MergeLogic(folder, name, group, plugin.Dir, plugin.Style)
		if err != nil {
			debug.PrintStack()
			panic(err)
		}
		fmt.Println(strings.TrimSpace((name + " logics merged!")))
	}
	rootPkg, err := golang.GetParentPackage(plugin.Dir)
	if err != nil {
		panic(err)
	}
	err = merge.GenRoutes(plugin.Dir, rootPkg, cfg, api)
	if err != nil {
		panic(err)
	}
	fmt.Println("routes modified!")
	fmt.Println(color.Green.Render("Done......."))
}
