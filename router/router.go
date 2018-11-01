package router

import (
	"fmt"

	"db-forum/api"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func Create(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Welcome!\n")
}

func routePostOnForum(ctx *fasthttp.RequestCtx) {
	options := ctx.UserValue("options").(string)
	if options == "/create" {
		api.CreateForum(ctx)
		return
	}

	options = options[1 : len(options)-7]
	api.CreateThread(ctx, options)
}

func CreateRouter() *fasthttprouter.Router {
	r := fasthttprouter.New()

	r.POST("/api/user/:nickname/create", api.CreateUser)
	r.GET("/api/user/:nickname/profile", api.GetUser)
	r.POST("/api/user/:nickname/profile", api.UpdateUser)

	r.POST("/api/forum/*options", routePostOnForum)
	r.GET("/api/forum/:slug/details", api.GetForum)
	r.GET("/api/forum/:slug/users", api.GetForumUsers)
	r.GET("/api/forum/:slug/threads", api.GetForumThreads)

	r.GET("/api/thread/:slug", api.GetThread)
	r.POST("/api/thread/:slug/create", api.CreatePost)
	r.GET("/api/thread/:slug/details", api.GetThread)
	r.POST("/api/thread/:slug/details", api.UpdateThread)
	r.POST("/api/thread/:slug/vote", api.VoteThread)

	r.GET("/api/thread/:slug/posts", api.GetPost)
	r.GET("/api/post/:slug/details", api.GetPostDetails)
	r.POST("/api/post/:slug/details", api.UpdatePost)

	r.GET("/api/service/status", api.GetServiceStatus)
	r.POST("/api/service/clear", api.ClearService)
	return r
}
