package api

import (
	"db-forum/database"
	"db-forum/models"
	"encoding/json"
	"log"
	"net/http"

	"strconv"

	"github.com/valyala/fasthttp"
	"golang.org/x/tools/container/intsets"
)

func CreateForum(ctx *fasthttp.RequestCtx) {
	var forum models.Forum
	body := ctx.PostBody()
	if err := json.Unmarshal(body, &forum); err != nil {

	}
	log.Println(forum)
	forumAuthor, err := database.GetUserByUsername(forum.User)
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find user"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	forum.User = forumAuthor.Nickname
	newForum, err := database.CreateForum(&forum)
	if err != nil {
		if err == database.ErrDuplicate {
			WriteResponse(ctx, http.StatusConflict, newForum)
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	log.Println(newForum)
	WriteResponse(ctx, http.StatusCreated, newForum)
}

func GetForum(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	forum, err := database.GetForum(slug)
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find forum"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	WriteResponse(ctx, http.StatusOK, forum)
}

func CreateForumThread(ctx *fasthttp.RequestCtx) {

}

func GetForumUsers(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	limit := string(ctx.QueryArgs().Peek("limit"))
	since := string(ctx.QueryArgs().Peek("since"))
	desc := string(ctx.QueryArgs().Peek("desc"))
	if limit == "" {
		limit = strconv.Itoa(intsets.MaxInt)
	}
	forum, err := database.GetForum(slug)
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"can't find forum"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	slug = forum.Slug
	users, err := database.GetForumUsers(slug, limit, since, desc)
	if err != nil {
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	WriteResponse(ctx, http.StatusOK, users)
}
