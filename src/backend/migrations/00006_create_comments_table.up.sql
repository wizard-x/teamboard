-- +goose Up
CREATE TABLE comments (
    id          CHAR(26)        NOT NULL,
    task_id     CHAR(26)        NOT NULL,
    author_id   CHAR(26)        NOT NULL,
    body        VARCHAR(2000)   NOT NULL,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ     NULL,

    CONSTRAINT pk_comments PRIMARY KEY (id),
    CONSTRAINT fk_comments_task FOREIGN KEY (task_id) REFERENCES tasks(id),
    CONSTRAINT fk_comments_author FOREIGN KEY (author_id) REFERENCES members(id)
);

CREATE INDEX idx_comments_task_id ON comments (task_id);
CREATE INDEX idx_comments_author_id ON comments (author_id);
CREATE INDEX idx_comments_task_created ON comments (task_id, created_at) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS comments;
