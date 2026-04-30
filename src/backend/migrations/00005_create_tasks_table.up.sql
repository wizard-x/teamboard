-- +goose Up
CREATE TABLE tasks (
    id           CHAR(26)        NOT NULL,
    column_id    CHAR(26)        NOT NULL,
    board_id     CHAR(26)        NOT NULL,
    title        VARCHAR(200)    NOT NULL,
    description  TEXT            NULL,
    status       VARCHAR(20)     NOT NULL DEFAULT 'todo',
    priority     VARCHAR(20)     NOT NULL DEFAULT 'medium',
    position     INTEGER         NOT NULL DEFAULT 0,
    assignee_id  CHAR(26)        NULL,
    due_date     TIMESTAMPTZ     NULL,
    created_by   CHAR(26)        NOT NULL,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ     NULL,

    CONSTRAINT pk_tasks PRIMARY KEY (id),
    CONSTRAINT fk_tasks_column FOREIGN KEY (column_id) REFERENCES board_columns(id),
    CONSTRAINT fk_tasks_board FOREIGN KEY (board_id) REFERENCES boards(id),
    CONSTRAINT fk_tasks_assignee FOREIGN KEY (assignee_id) REFERENCES members(id),
    CONSTRAINT fk_tasks_creator FOREIGN KEY (created_by) REFERENCES members(id),
    CONSTRAINT chk_tasks_status CHECK (status IN ('todo', 'in_progress', 'review', 'done')),
    CONSTRAINT chk_tasks_priority CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_tasks_position CHECK (position >= 0)
);

CREATE INDEX idx_tasks_column_id ON tasks (column_id);
CREATE INDEX idx_tasks_board_id ON tasks (board_id);
CREATE INDEX idx_tasks_assignee_id ON tasks (assignee_id);
CREATE INDEX idx_tasks_column_position ON tasks (column_id, position) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_status ON tasks (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_due_date ON tasks (due_date) WHERE deleted_at IS NULL AND due_date IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS tasks;
