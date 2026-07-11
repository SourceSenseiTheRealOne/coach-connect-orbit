CREATE TABLE profiles (
  id uuid PRIMARY KEY,
  clerk_subject varchar(256) NOT NULL UNIQUE,
  display_name varchar(80) NOT NULL,
  avatar_url varchar(2048),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT profiles_timestamps CHECK (updated_at >= created_at)
);

CREATE TABLE posts (
  id uuid PRIMARY KEY,
  author_id uuid NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  body varchar(8800) NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT posts_body_nonempty CHECK (length(btrim(body)) > 0),
  CONSTRAINT posts_timestamps CHECK (updated_at >= created_at)
);
CREATE INDEX posts_feed_idx ON posts (created_at DESC, id DESC);
CREATE INDEX posts_author_id_idx ON posts (author_id);

CREATE TABLE post_likes (
  id uuid PRIMARY KEY,
  post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  profile_id uuid NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL
);
CREATE UNIQUE INDEX post_likes_post_profile_unique_idx ON post_likes (post_id, profile_id);
CREATE INDEX post_likes_profile_idx ON post_likes (profile_id, created_at DESC);

CREATE TABLE saved_posts (
  id uuid PRIMARY KEY,
  post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  profile_id uuid NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL
);
CREATE UNIQUE INDEX saved_posts_post_profile_unique_idx ON saved_posts (post_id, profile_id);
CREATE INDEX saved_posts_private_feed_idx ON saved_posts (profile_id, created_at DESC, id DESC);

CREATE TABLE comments (
  id uuid PRIMARY KEY,
  post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  author_id uuid NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  body varchar(8800) NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT comments_body_nonempty CHECK (length(btrim(body)) > 0),
  CONSTRAINT comments_timestamps CHECK (updated_at >= created_at)
);
CREATE INDEX comments_post_feed_idx ON comments (post_id, created_at DESC, id DESC);
CREATE INDEX comments_author_id_idx ON comments (author_id);

CREATE TABLE post_media (
  id uuid PRIMARY KEY,
  post_id uuid NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  object_key varchar(512) NOT NULL,
  public_url varchar(2048) NOT NULL,
  alt_text varchar(240) NOT NULL,
  mime_type varchar(64) NOT NULL,
  width integer NOT NULL,
  height integer NOT NULL,
  created_at timestamptz NOT NULL,
  CONSTRAINT post_media_width_positive CHECK (width > 0),
  CONSTRAINT post_media_height_positive CHECK (height > 0)
);
CREATE UNIQUE INDEX post_media_post_id_idx ON post_media (post_id);
