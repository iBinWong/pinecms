package router

import (
	"io/ioutil"
	"path/filepath"

	"github.com/xiusin/pine"
	"github.com/xiusin/pinecms/src/config"
)

func InitStatics(app *pine.Application) {
	for _, static := range config.App().Statics {
		app.Static(static.Route, filepath.FromSlash(static.Path), 1)
	}
}

func InitRouter(app *pine.Application) {

	// 前端路由注册
	//app.Handle(new(frontend.FescController))
	//app.Handle(new(frontend.IndexController))
	app.GET("/", func(ctx *pine.Context) {
		ctx.Redirect("/admin/")
	})

	app.GET("/admin", func(ctx *pine.Context) {
		if byts, err := ioutil.ReadFile("admin/dist/index.html"); err != nil {
			ctx.Abort(500, err.Error())
		} else {
			_ = ctx.WriteHTMLBytes(byts)
		}
	})

}