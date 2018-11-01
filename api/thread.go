package api

import (
	"db-forum/database"
	"db-forum/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/valyala/fasthttp"
	"golang.org/x/tools/container/intsets"
)

func CreateThread(ctx *fasthttp.RequestCtx, forumName string) {
	var thread models.Thread
	body := ctx.PostBody()
	if err := json.Unmarshal(body, &thread); err != nil {
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	thread.Forum = forumName
	user, err := database.GetUserByUsername(thread.Author)
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find user"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	thread.Author = user.Nickname
	forum, err := database.GetForum(thread.Forum)
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find forum"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	thread.Forum = forum.Slug
	if thread.Slug != "" {
		existsThread, err := database.GetThreadBySlug(thread.Slug)
		if err != nil {
			if err != database.ErrNotFound {
				log.Println(err.Error())
				WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
				return
			}
		}
		if existsThread != nil {
			WriteResponse(ctx, http.StatusConflict, existsThread)
			return
		}
	}
	newThread, err := database.CreateThread(&thread)
	if err != nil {
		if err == database.ErrDuplicate {
			WriteResponse(ctx, http.StatusConflict, newThread)
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	WriteResponse(ctx, http.StatusCreated, newThread)
}

func GetThread(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	var thread *models.Thread
	var err error
	if govalidator.IsNumeric(slug) {
		thread, err = database.GetThread(slug, slug)
	} else {
		thread, err = database.GetThreadBySlug(slug)
	}

	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find forum by slug: " + slug})
			return
		}
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	WriteResponse(ctx, http.StatusOK, thread)
}

func GetForumThreads(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	limit := ctx.QueryArgs().Peek("limit")
	desc := ctx.QueryArgs().Peek("desc")
	since := ctx.QueryArgs().Peek("since")
	queryDesc, querySince := string(desc), string(since)
	var queryLimit int
	var err error
	if string(limit) == "" {
		queryLimit = intsets.MaxInt
	} else {
		queryLimit, err = strconv.Atoi(string(limit))
		if err != nil {
			WriteResponse(ctx, http.StatusBadRequest, models.Error{err.Error()})
		}
	}

	if queryDesc == "true" {
		queryDesc = "DESC"
	} else {
		queryDesc = "ASC"
	}

	_, err = database.GetForum(slug)
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find forum by slug: " + slug})
			return
		}
	}

	threads, err := database.GetForumThreads(slug, querySince, queryDesc, queryLimit)
	if err != nil {
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}

	WriteResponse(ctx, http.StatusOK, (*threads))
}

func UpdateThread(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	body := ctx.PostBody()
	var thread *models.Thread
	var postThread models.Thread
	if err := json.Unmarshal(body, &postThread); err != nil {
		WriteResponse(ctx, http.StatusBadRequest, models.Error{err.Error()})
		return
	}
	var err error
	if govalidator.IsNumeric(slug) {
		thread, err = database.GetThread(slug, slug)
	} else {
		thread, err = database.GetThreadBySlug(slug)
	}
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{err.Error()})
			return
		}
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	thread.Title, thread.Message = postThread.Title, postThread.Message
	resThread, err := database.UpdateThread(thread)
	if err != nil {
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	WriteResponse(ctx, http.StatusOK, resThread)
}

func VoteThread(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	body := ctx.PostBody()
	var voice models.Vote
	if err := json.Unmarshal(body, &voice); err != nil {
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusBadRequest, models.Error{err.Error()})
		return
	}
	user, err := database.GetUserByUsername(voice.Nickname)
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find user"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	voice.Nickname = user.Nickname
	var thread *models.Thread
	if govalidator.IsNumeric(slug) {
		thread, err = database.GetThread(slug, slug)
	} else {
		thread, err = database.GetThreadBySlug(slug)
	}
	if err != nil {
		if err == database.ErrNotFound {
			WriteResponse(ctx, http.StatusNotFound, models.Error{"Can't find thread"})
			return
		}
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	slug = thread.Slug
	voice.ThreadId = thread.ID
	newVote, err := database.VoteThread(&voice)
	if err != nil {
		log.Println(err.Error())
		WriteResponse(ctx, http.StatusInternalServerError, models.Error{err.Error()})
		return
	}
	thread.Votes = newVote
	WriteResponse(ctx, http.StatusOK, thread)
}
