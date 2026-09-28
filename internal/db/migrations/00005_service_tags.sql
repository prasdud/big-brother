-- +goose Up
ALTER TABLE services ADD COLUMN tags TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE services DROP COLUMN tags;
