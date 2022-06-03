-- +goose Up
ALTER TABLE current RENAME COLUMN time TO timestamp;
ALTER TABLE power RENAME COLUMN time TO timestamp;
ALTER TABLE energy RENAME COLUMN time TO timestamp;

-- +goose Down
ALTER TABLE current RENAME COLUMN timestamp TO time;
ALTER TABLE power RENAME COLUMN timestamp TO time;
ALTER TABLE energy RENAME COLUMN timestamp TO time;
