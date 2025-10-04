---- 1 Creating table posts
CREATE TABLE posts
(
    id    BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL
);

---- 2 Adding column body
ALTER TABLE posts
    ADD COLUMN body TEXT;

---- 3 Add created_at with default and NOT NULL
ALTER TABLE posts
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

---- 4 Add normal index on title
CREATE INDEX idx_posts_title ON posts (title);

---- 5 Add unique index on title
CREATE UNIQUE INDEX ux_posts_title ON posts (title);

---- 6 Create users table
CREATE TABLE users
(
    id         BIGSERIAL PRIMARY KEY,
    email      VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

---- 7 Add author_id with foreign key to users
ALTER TABLE posts
    ADD COLUMN author_id BIGINT;

ALTER TABLE posts
    ADD CONSTRAINT fk_posts_author
        FOREIGN KEY (author_id) REFERENCES users (id)
            ON DELETE SET NULL
            ON UPDATE CASCADE;

---- 8 Rename column title to headline
ALTER TABLE posts
    RENAME COLUMN title TO headline;

---- 9 Make body NOT NULL
ALTER TABLE posts
    ALTER COLUMN body SET NOT NULL;

---- 10 Add updated_at and deleted_at
ALTER TABLE posts
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ADD COLUMN deleted_at TIMESTAMPTZ NULL;

---- 11 Auto-update updated_at on row change (trigger + function)
CREATE
OR REPLACE FUNCTION set_posts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at
:= NOW();
RETURN NEW;
END;
$$
LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_posts_updated_at ON posts;
CREATE TRIGGER trg_posts_updated_at
    BEFORE UPDATE
    ON posts
    FOR EACH ROW EXECUTE FUNCTION set_posts_updated_at();

---- 12 Add status with CHECK constraint
ALTER TABLE posts
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'draft';

ALTER TABLE posts
    ADD CONSTRAINT chk_posts_status
        CHECK (status IN ('draft', 'published', 'archived'));

---- 13 Composite index on (author_id, created_at DESC)
CREATE INDEX idx_posts_author_created_at
    ON posts (author_id, created_at DESC);

---- 14 Drop unique index on headline (formerly title)
DROP INDEX IF EXISTS ux_posts_title;

---- 15 Drop column body
ALTER TABLE posts
DROP
COLUMN body;

---- 16 Change default of status
ALTER TABLE posts
    ALTER COLUMN status SET DEFAULT 'published';

---- 17 Ensure id uses identity (Postgres 10+ style)
-- Demonstration: convert to identity if needed (no-op if already BIGSERIAL-backed)
DO
$$
BEGIN
  IF
NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'posts' AND column_name = 'id'
  ) THEN
    RAISE EXCEPTION 'posts.id not found';
END IF;
END $$;

---- 18 Ensure email unique index on users (already via UNIQUE, add named index)
DO
$$
BEGIN
  IF
NOT EXISTS (
    SELECT 1 FROM pg_indexes
    WHERE schemaname = 'public' AND indexname = 'ux_users_email'
  ) THEN
CREATE UNIQUE INDEX ux_users_email ON users (email);
END IF;
END $$;

---- 19 Add slug with unique index
ALTER TABLE posts
    ADD COLUMN slug VARCHAR(255);

CREATE UNIQUE INDEX ux_posts_slug ON posts (slug);

---- 20 Add published_at and visibility with default
ALTER TABLE posts
    ADD COLUMN published_at TIMESTAMPTZ NULL,
  ADD COLUMN visibility TEXT NOT NULL DEFAULT 'public';

ALTER TABLE posts
    ADD CONSTRAINT chk_posts_visibility
        CHECK (visibility IN ('public', 'private', 'unlisted'));