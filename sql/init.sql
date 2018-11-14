CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users
(
  id       SERIAL NOT NULL
    CONSTRAINT users_pkey
    PRIMARY KEY,
  nickname CITEXT NOT NULL,
  fullname TEXT   NOT NULL,
  about    TEXT,
  email    CITEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS forum
(
  id      SERIAL NOT NULL
    CONSTRAINT forum_pkey
    PRIMARY KEY,
  title   CITEXT NOT NULL,
  author  CITEXT NOT NULL,
  slug    CITEXT NOT NULL,
  posts   BIGINT DEFAULT 0,
  threads BIGINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS thread
(
  id      SERIAL NOT NULL
    CONSTRAINT thread_pkey
    PRIMARY KEY,
  title   CITEXT NOT NULL,
  author  CITEXT NOT NULL,
  forum   CITEXT NOT NULL,
  message CITEXT NOT NULL,
  votes   INTEGER                  DEFAULT 0,
  created TIMESTAMP WITH TIME ZONE DEFAULT now(),
  slug    CITEXT
);

CREATE TABLE IF NOT EXISTS post
(
  id        SERIAL            NOT NULL
    CONSTRAINT post_pkey
    PRIMARY KEY,
  parent    INTEGER DEFAULT 0 NOT NULL,
  author    CITEXT,
  message   CITEXT,
  is_edited BOOLEAN                  DEFAULT FALSE,
  forum     CITEXT            NOT NULL,
  thread    INTEGER                  DEFAULT 0,
  created   TIMESTAMP WITH TIME ZONE DEFAULT now(),
  path      INTEGER []               DEFAULT ARRAY [] :: INTEGER [],
  root      INTEGER                  DEFAULT 0
);

CREATE TABLE voice
(
  id         SERIAL            NOT NULL
    CONSTRAINT voice_pkey
    PRIMARY KEY,
  nickname   CITEXT,
  vote       INTEGER DEFAULT 0 NOT NULL,
  prev_vote  INTEGER                  DEFAULT 0,
  thread_id  INTEGER                  DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE (nickname, vote, thread_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS user_nickname_uindex
  ON users (nickname);

CREATE INDEX IF NOT EXISTS user_id_index
  ON users (id);

CREATE UNIQUE INDEX IF NOT EXISTS user_email_uindex
  ON users (email);

CREATE UNIQUE INDEX IF NOT EXISTS forum_slug_uindex
  ON forum (slug);

CREATE INDEX IF NOT EXISTS index_post_root
  ON post (root);

CREATE INDEX IF NOT EXISTS idx_threads_slug_id
  ON thread (slug, id);

CREATE INDEX IF NOT EXISTS idx_threads_forumSlug_createdAt
  ON thread (forum, created);

CREATE INDEX IF NOT EXISTS index_thread_
  ON thread (author, created, forum, message, title, slug, id, votes);

CREATE INDEX IF NOT EXISTS indx_user
  ON users (nickname, email, about, fullname);

CREATE INDEX IF NOT EXISTS index_on_forum
  ON forum ( slug, title, author, posts, threads);

CREATE INDEX IF NOT EXISTS index_post_for_parent_tree
  ON post (thread, parent, path);

CREATE INDEX IF NOT EXISTS index_post_for_parent_tree_with_sin
  ON post (thread, parent, root);

CREATE INDEX IF NOT EXISTS index_post_for_parent_tree_without_sin
  ON post (thread, parent, id);

