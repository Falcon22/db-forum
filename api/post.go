package api

import (
	"db-forum/database"
	"db-forum/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/valyala/fasthttp"
	"golang.org/x/tools/container/intsets"
)

func CreatePost(ctx *fasthttp.RequestCtx) {
	var posts []models.Post
	posts = make([]models.Post, 0)
	body := ctx.PostBody()
	slug := ctx.UserValue("slug").(string)
	if err := json.Unmarshal(body, &posts); err != nil {
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusBadRequest, models.Error{err.Error()})
	}
	resPosts, err := database.CreatePosts(&posts, slug)
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find user"})
			return
		}
		if err == database.ErrDuplicate {
			WriteResponse(ctx, http.StatusConflict, models.Error{"ti priomniy"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	WriteResponse(ctx, http.StatusCreated, resPosts)
}

func GetPost(ctx *fasthttp.RequestCtx) {
	var err error
	var posts *[]models.Post
	var thread *models.Thread
	slug := ctx.UserValue("slug").(string)
	limit, sort, since, desc := string(ctx.QueryArgs().Peek("limit")),
		string(ctx.QueryArgs().Peek("sort")),
		string(ctx.QueryArgs().Peek("since")),
		string(ctx.QueryArgs().Peek("desc"))
	if limit == "" {
		limit = strconv.Itoa(intsets.MaxInt)
	}
	if govalidator.IsNumeric(slug) {
		thread, err = database.GetThread(slug, slug)
	} else {
		thread, err = database.GetThreadBySlug(slug)
	}
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find thread by slug " + slug})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	switch sort {
	case "flat":
		posts, err = database.GetPostsFlat(thread.ID, limit, since, desc)
	case "tree":
		posts, err = database.GetPostsTree(thread.ID, limit, since, desc)
	case "parent_tree":
		posts, err = database.GetPostsParentTree(thread.ID, limit, since, desc)
	default:
		posts, err = database.GetPostsFlat(thread.ID, limit, since, desc)
	}
	if err != nil {
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	WriteResponse(ctx, http.StatusOK, posts)
}

func GetPostDetails(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	related := string(ctx.QueryArgs().Peek("related"))
	params := make([]string, 0)
	params = append(params, strings.Split(related, ",")...)
	id, err := strconv.Atoi(slug)
	if err != nil {
		WriteResponse(ctx, http.StatusBadRequest, models.Error{err.Error()})
		return
	}
	post, err := database.GetPostByID(int64(id))
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"can't find post"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	var postFull models.PostFull
	for _, param := range params {
		switch param {
		case "user":
			postFull.Author, err = database.GetUserByUsername(post.Author)
		case "forum":
			postFull.Forum, err = database.GetForum(post.Forum)
		case "thread":
			postFull.Thread, err = database.GetThreadByIDint32(post.Thread)
		}
	}
	postFull.Post = post
	WriteResponse(ctx, http.StatusOK, postFull)
}

func UpdatePost(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	id, err := strconv.Atoi(slug)
	if err != nil {
		WriteResponse(ctx, http.StatusBadRequest, models.Error{err.Error()})
		return
	}
	var post models.Post
	body := ctx.PostBody()
	if err := post.UnmarshalJSON(body); err != nil {
		WriteResponse(ctx, http.StatusBadRequest, models.Error{err.Error()})
		return
	}
	post.ID = int64(id)
	newPost, err := database.UpdatePost(&post)
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"can't find post"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}

	WriteResponse(ctx, http.StatusOK, newPost)
}
