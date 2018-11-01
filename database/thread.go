package database

import (
	"db-forum/models"

	"database/sql"

	"github.com/pkg/errors"
)

var createThread = `INSERT INTO thread (title, author, forum, message, created, slug) VALUES ($1, $2, $3, $4, $5, $6) RETURNING slug, id;`

var updateForumCount = `UPDATE forum SET threads = threads + 1 WHERE slug = $1;`

func CreateThread(thread *models.Thread) (*models.Thread, error) {
	var slug string
	var id int32
	tx, err := db.pg.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "can't start transaction")
	}
	if err := db.CreateThreadStmt.QueryRow(thread.Title, thread.Author, thread.Forum, thread.Message, thread.Created, thread.Slug).Scan(&slug, &id); err != nil {
		existThread, error := GetThreadBySlug(thread.Slug)
		if error == ErrNotFound {
			tx.Rollback()
			return nil, errors.Wrap(err, "can't insert into db")
		}
		tx.Rollback()
		return existThread, ErrDuplicate
	}
	updateForumCountStmt, err := db.pg.Prepare(updateForumCount)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "can't prepare query")
	}
	if _, err := updateForumCountStmt.Exec(thread.Forum); err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "can't exec query")
	}
	tx.Commit()
	thread.Slug = slug
	thread.ID = id
	return thread, nil
}

var getThreadByID = `SELECT id, title, author, forum, message, votes, created, slug FROM thread WHERE id = $1;`

func GetThreadByID(id string) (*models.Thread, error) {
	var thread models.Thread
	if err := db.pg.QueryRow(getThreadByID, id).Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &thread.Created, &thread.Slug); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "can't select from thread")
	}
	return &thread, nil
}

func GetThreadByIDint32(id int32) (*models.Thread, error) {
	var thread models.Thread
	if err := db.pg.QueryRow(getThreadByID, id).Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &thread.Created, &thread.Slug); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "can't select from thread")
	}
	return &thread, nil
}

var getThreadBySlug = `SELECT id, title, author, forum, message, votes, created, slug FROM thread WHERE slug = $1;`

func GetThreadBySlug(slug string) (*models.Thread, error) {
	var thread models.Thread
	if err := db.GetThreadBySlugStmt.QueryRow(slug).Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &thread.Created, &thread.Slug); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "can't select from thread")
	}
	return &thread, nil
}

var getThread = `SELECT id, title, author, forum, message, votes, created, slug FROM thread WHERE id = $1 OR slug = $2;`

func GetThread(id string, slug string) (*models.Thread, error) {
	var thread models.Thread
	if err := db.GetThreadStmt.QueryRow(id, slug).Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &thread.Created, &thread.Slug); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "can't select from thread")
	}
	return &thread, nil
}

var getPrevVote = `SELECT id, nickname, vote, thread_id FROM voice WHERE nickname = $1 AND thread_id = $2 ORDER BY created_at DESC LIMIT 1;`
var createVoteThread = `INSERT INTO voice (nickname, vote, prev_vote, thread_id) VALUES ($1, $2, $3, $4) RETURNING id;`
var updateVoteThread = `UPDATE thread SET votes = votes + $1 WHERE id = $2 RETURNING votes;`

func VoteThread(vote *models.Vote) (newVote int32, err error) {
	var nVote, prevVote models.Vote
	tx, err := db.pg.Begin()
	if err != nil {
		return 0, errors.Wrap(err, "can't start tx")
	}
	if err := db.GetPrevVoteThreadStmt.QueryRow(vote.Nickname, vote.ThreadId).Scan(&prevVote.ID, &prevVote.Nickname, &prevVote.Voice, &prevVote.ThreadId); err != nil {
		if err != sql.ErrNoRows {
			tx.Rollback()
			return 0, errors.Wrap(err, "can't select from voice")
		}
		prevVote.ID = 0
		prevVote.Voice = 0
	}
	if err := db.CreatVoteThreadStmt.QueryRow(vote.Nickname, vote.Voice, prevVote.ID, vote.ThreadId).Scan(&nVote.ID); err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "can't insert into voice")
	}
	if err := db.UpdateVoteThreadStmt.QueryRow(vote.Voice-prevVote.Voice, vote.ThreadId).Scan(&newVote); err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "can't update thread")
	}
	tx.Commit()
	return newVote, nil
}

var updateThread = `UPDATE thread SET title = coalesce(coalesce(nullif($2, ''), title)),
			message = coalesce(coalesce(nullif($3, ''), message))
			WHERE id = $1 RETURNING title, message;`

func UpdateThread(thread *models.Thread) (*models.Thread, error) {
	newThread := *thread
	updateThreadStmt, err := db.pg.Prepare(updateThread)
	if err != nil {
		return nil, errors.Wrap(err, "can't prepare query")
	}
	if err := updateThreadStmt.QueryRow(thread.ID, thread.Title, thread.Message).Scan(&newThread.Title, &newThread.Message); err != nil {
		return nil, errors.Wrap(err, "can't update thread")
	}
	return &newThread, nil
}