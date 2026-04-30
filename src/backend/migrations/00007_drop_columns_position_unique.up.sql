-- +goose Up
DROP INDEX IF EXISTS idx_columns_board_position;

-- +goose Down
CREATE UNIQUE INDEX idx_columns_board_position ON board_columns (board_id, position) WHERE deleted_at IS NULL;
