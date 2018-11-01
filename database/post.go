package database

import (
	"database/sql"
	"db-forum/models"

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
	for _, post := range *posts {
		newPost := post
		user, err := GetUserByUsername(post.Author)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		newPost.Author, newPost.Thread, newPost.Forum = user.Nickname, thread.ID, thread.Forum
		if err := tx.Stmt(db.CreatePostStmt).QueryRow(post.Parent, user.Nickname, post.Message, thread.Forum, thread.ID).Scan(&newPost.ID, &newPost.Created); err != nil {
			tx.Rollback()
			return nil, errors.Wrap(err, "can't insert into post")
		}
		resPosts = append(resPosts, newPost)
	}
	if _, err := db.pg.Exec(updateForumPostsCount, thread.Forum, len(resPosts)); err != nil {
		return &resPosts, errors.Wrap(err, "can't update forum")
	}
	tx.Commit()
	for _, post := range resPosts {
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
	return &resPosts, nil
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
			getPostsFlat += " AND id < $2 ORDER BY created DESC, id DESC LIMIT $3;"
		} else {
			getPostsFlat += " AND id > $2 ORDER BY created, id ASC LIMIT $3;"
		}
		rows, err = db.pg.Query(getPostsFlat, thread, since, limit)
	} else {
		if desc == "true" {
			getPostsFlat += " ORDER BY created DESC, id DESC LIMIT $2;"
		} else {
			getPostsFlat += " ORDER BY created, id LIMIT $2;"
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
