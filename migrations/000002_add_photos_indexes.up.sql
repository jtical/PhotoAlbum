-- Filename: migrations/000003_add_photos_indexes.up.sql

CREATE INDEX IF NOT EXISTS photos_title_idx ON photos USING GIN(to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS photos_photo_idx ON photos USING GIN(to_tsvector('simple', photo));
CREATE INDEX IF NOT EXISTS photos_description_idx ON photos USING GIN(to_tsvector('simple', description));