-- Filename: migrations/000003_add_photos_indexes.down.sql
DROP INDEX IF EXISTS photos_title_idx;
DROP INDEX IF EXISTS photos_photo_idx;
DROP INDEX IF EXISTS photos_description_idx;