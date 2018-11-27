package database

import (
	"database/sql"
	"db-forum/models"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var createPost = `INSERT INTO post (parent, author, message, forum, thread) 
VALUES ($1, $2, $3, $4, $5) RETURNING id, created;`

var updatePostPath = `UPDATE post SET root = $2, path = $3 WHERE id = $1;`
var updateForumPostsCount = `UPDATE forum SET posts = posts + $2 WHERE slug = $1; `

func CreatePost(post *models.Post) (*models.Post, error) {
	newPost := *post
	if err := db.CreatePostStmt.QueryRow(post.Parent, post.Author, post.Message, post.Forum, post.Thread).Scan(&newPost.ID, &newPost.Created); err != nil {
		return nil, errors.Wrap(err, "can't insert into post")
	}
	return &newPost, nil
}

var getPath = `SELECT path FROM post WHERE id = $1 AND thread = $2;`

func CreatePosts(posts *[]models.Post, threadSlug string) (*[]models.Post, error) {
	tx, err := db.pg.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "can't start transaction")
	}
	resPosts := make([]models.Post, 0)
	var thread *models.Thread
	if govalidator.IsNumeric(threadSlug) {
		thread, err = GetThreadByID(threadSlug)
	} else {
		thread, err = GetThreadBySlug(threadSlug)
	}
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	query := `insert into post (parent, message, thread, author, forum) values `
	queryEnd := " returning id, is_edited, created"
	var queryValues []string

	args := make([]interface{}, 0, len(*posts)*5)
	parents := make([]string, 0, len(*posts))

	if len(*posts) == 0 {
		return &resPosts, nil
	}

	if len(*posts) < 100 {
		for i, post := range *posts {
			author, err := GetUserByUsername(post.Author)
			if err != nil || author == nil {
				return nil, ErrNotFound
			}
			(*posts)[i].Author = author.Nickname
		}
	}

	if len(*posts) == 100 {
		for _, post := range *posts {
			if post.Parent != 0 {
				parents = append(parents, strconv.Itoa(int(post.Parent)))
			}
			args = append(args, post.Parent, post.Message, thread.ID, post.Author, thread.Forum)
		}
	} else {
		for _, post := range *posts {
			if post.Parent != 0 {
				parents = append(parents, strconv.Itoa(int(post.Parent)))
			}
			queryValues = append(queryValues, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", len(args)+1, len(args)+2, len(args)+3, len(args)+4, len(args)+5))
			args = append(args, post.Parent, post.Message, thread.ID, post.Author, thread.Forum)
		}
	}

	if len(parents) != 0 {
		rows, err := tx.Query(fmt.Sprint(`select thread from post where id in (`, strings.Join(parents, ","), ")"))
		hasP := false

		for rows.Next() {
			hasP = true

			var tId int32
			err = rows.Scan(&tId)
			if err != nil {
				log.Println(err)
				log.Println(tId)
			}

			if tId != thread.ID {
				return nil, ErrDuplicate
			}
		}

		if !hasP {
			return nil, ErrDuplicate
		}

	}

	query += strings.Join(queryValues, ",") + queryEnd







	rows, err := tx.Query(query, args...)
	var par []string
	var nopar []string

	a := make(map[string]bool)
	for i, post := range *posts {
		if rows.Next() {
			err = rows.Scan(&((*posts)[i].ID), &((*posts)[i].IsEdited), &((*posts)[i].Created))
			if post.Parent != 0 {
				par = append(par, strconv.Itoa(int(post.ID)))
			} else {
				nopar = append(nopar, strconv.Itoa(int(post.ID)))
			}

			a["'"+post.Author+"'"] = true
			(*posts)[i].Forum = thread.Forum
			(*posts)[i].Thread = thread.ID
		}
	}
	rows.Close()

	auth := make([]string, 0, len(a))
	for key := range a {
		auth = append(auth, key)
	}

	if err := rows.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		log.Println("error on main query")
		log.Println(err)
		return nil, ErrDuplicate
	}
	tx.Exec(updateForumPostsCount, thread.Forum, len(*posts))

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	for _, post := range *posts {
		var root int64
		sqlPath := make([]sql.NullInt64, 0)
		if post.Parent != 0 {
			if err = db.pg.QueryRow(getPath, post.Parent, thread.ID).Scan(pq.Array(&sqlPath)); err != nil {
				tx.Rollback()
				if err == sql.ErrNoRows {
					return nil, ErrDuplicate
				}
				return nil, errors.Wrap(err, "can't get path from parent post")
			}
			root = sqlPath[0].Int64
		} else {
			root = post.ID
		}
		sqlPath = append(sqlPath, sql.NullInt64{post.ID, true})
		updateStmt, err := db.pg.Prepare(updatePostPath)
		if err != nil {
			return nil, errors.Wrap(err, "can't prepare post path")
		}
		if _, err = updateStmt.Exec(post.ID, root, pq.Array(sqlPath)); err != nil {
			return nil, errors.Wrap(err, "can't update post path")
		}
	}

	return posts, nil
}

var getPostByID = `SELECT id, parent, author, message, is_edited, forum, thread, created 
FROM post WHERE id = $1;`

func GetPostByID(id int64) (*models.Post, error) {
	var post models.Post
	if err := db.GetPostByIDStmt.QueryRow(id).Scan(&post.ID, &post.Parent, &post.Author, &post.Message, &post.IsEdited, &post.Forum, &post.Thread, &post.Created); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "can't select from post")
	}
	return &post, nil
}

func GetPostsFlat(thread int32, limit string, since string, desc string) (*[]models.Post, error) {
	posts := make([]models.Post, 0)
	getPostsFlat := `SELECT id, parent, author, message, forum, thread, created FROM post WHERE thread = $1`
	var rows *sql.Rows
	var err error
	if since != "" {
		if desc == "true" {
			getPostsFlat += " AND id < $2 ORDER BY id DESC LIMIT $3;"
		} else {
			getPostsFlat += " AND id > $2 ORDER BY id ASC LIMIT $3;"
		}
		rows, err = db.pg.Query(getPostsFlat, thread, since, limit)
	} else {
		if desc == "true" {
			getPostsFlat += " ORDER BY id DESC LIMIT $2;"
		} else {
			getPostsFlat += " ORDER BY id LIMIT $2;"
		}
		rows, err = db.pg.Query(getPostsFlat, thread, limit)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return &posts, nil
		}
		return nil, errors.Wrap(err, "can't select from posts")
	}
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Parent, &post.Author, &post.Message, &post.Forum, &post.Thread, &post.Created); err != nil {
			return nil, errors.Wrap(err, "can't scan rows")
		}
		posts = append(posts, post)
	}
	return &posts, nil
}

func GetPostsTree(thread int32, limit string, since string, desc string) (*[]models.Post, error) {
	posts := make([]models.Post, 0)
	getPostTree := `SELECT id, parent, author, message, forum, thread, created FROM post WHERE thread = $1 `
	var rows *sql.Rows
	var err error
	if since != "" {
		if desc == "true" {
			getPostTree += ` AND path < (SELECT path FROM post WHERE id = $2 ) ORDER BY path DESC LIMIT $3;`
		} else {
			getPostTree += ` AND path > (SELECT path FROM post WHERE id = $2 ) ORDER BY path LIMIT $3;`
		}
		rows, err = db.pg.Query(getPostTree, thread, since, limit)
	} else {
		since = "0"
		if desc == "true" {
			getPostTree += ` ORDER BY path DESC LIMIT $2;`
		} else {
			getPostTree += ` ORDER BY path LIMIT $2;`
		}
		rows, err = db.pg.Query(getPostTree, thread, limit)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return &posts, nil
		}
		return nil, errors.Wrap(err, "can't select from posts")
	}
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Parent, &post.Author, &post.Message, &post.Forum, &post.Thread, &post.Created); err != nil {
			return nil, errors.Wrap(err, "can't scan rows")
		}
		posts = append(posts, post)
	}
	return &posts, nil
}

func GetPostsParentTree(thread int32, limit string, since string, desc string) (*[]models.Post, error) {
	posts := make([]models.Post, 0)
	getPostParentTree := `SELECT id, parent, author, message, forum, thread, created FROM post WHERE root IN (SELECT id FROM post WHERE thread = $1 AND parent = 0 `
	var rows *sql.Rows
	var err error
	if since != "" {
		if desc == "true" {
			getPostParentTree += ` AND root < (SELECT root FROM post WHERE id = $2 ) ORDER BY root DESC LIMIT $3)  ORDER BY root desc, path ;`
		} else {
			getPostParentTree += ` AND path > (SELECT path FROM post WHERE id = $2 ) ORDER BY id LIMIT $3) ORDER BY path;`
		}
		rows, err = db.pg.Query(getPostParentTree, thread, since, limit)
	} else {
		since = "0"
		if desc == "true" {
			getPostParentTree += `ORDER BY root DESC LIMIT $2) ORDER BY root DESC, path;`
		} else {
			getPostParentTree += `ORDER BY id LIMIT $2) ORDER BY path;`
		}
		rows, err = db.pg.Query(getPostParentTree, thread, limit)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return &posts, nil
		}
		return nil, errors.Wrap(err, "can't select from posts")
	}
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Parent, &post.Author, &post.Message, &post.Forum, &post.Thread, &post.Created); err != nil {
			return nil, errors.Wrap(err, "can't scan rows")
		}
		posts = append(posts, post)
	}
	return &posts, nil
}

var updatePost = `UPDATE post SET message = coalesce(coalesce(nullif($2, ''), message)), is_edited = $3 WHERE id = $1 RETURNING message, author, is_edited, thread, created, forum;`

func UpdatePost(post *models.Post) (*models.Post, error) {
	newPost := *post
	oldPost, err := GetPostByID(post.ID)
	if err != nil {
		return nil, err
	}
	if len(post.Message) == 0 {
		post.IsEdited = false
	} else {
		if oldPost != nil {
			if oldPost.Message == post.Message {
				post.IsEdited = false
			} else {
				post.IsEdited = true
			}
		}
	}

	if err := db.pg.QueryRow(updatePost, post.ID, post.Message, post.IsEdited).Scan(&newPost.Message, &newPost.Author, &newPost.IsEdited, &newPost.Thread, &newPost.Created, &newPost.Forum); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "can't update post")
	}
	return &newPost, nil
}
