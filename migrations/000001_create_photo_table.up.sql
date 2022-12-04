-- Filename: migrations/000001_create_photos_table.up.sql

CREATE TABLE IF NOT EXISTS photos (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    photo text NOT NULL,
    description text NOT NULL,
    version integer NOT NULL DEFAULT 1
);