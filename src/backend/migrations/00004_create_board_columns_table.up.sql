-- +goose Up
CREATE TABLE board_columns (
    id          CHAR(26)        NOT NULL,
    board_id    CHAR(26)        NOT NULL,
    name        VARCHAR(50)     NOT NULL,
    position    INTEGER         NOT NULL DEFAULT 0,
    status      VARCHAR(20)     NOT NULL DEFAULT 'todo',
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ     NULL,

    CONSTRAINT pk_board_columns PRIMARY KEY (id),
    CONSTRAINT fk_columns_board FOREIGN KEY (board_id) REFERENCES boards(id),
    CONSTRAINT chk_columns_status CHECK (status IN ('todo', 'in_progress', 'review', 'done')),
    CONSTRAINT chk_columns_position CHECK (position >= 0)
);

CREATE INDEX idx_columns_board_id ON board_columns (board_id);
CREATE UNIQUE INDEX idx_columns_board_position ON board_columns (board_id, position) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS board_columns;
