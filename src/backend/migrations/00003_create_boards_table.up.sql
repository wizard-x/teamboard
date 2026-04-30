-- +goose Up
CREATE TABLE boards (
    id           CHAR(26)        NOT NULL,
    team_id      CHAR(26)        NOT NULL,
    name         VARCHAR(100)    NOT NULL,
    description  VARCHAR(500)    NULL,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ     NULL,

    CONSTRAINT pk_boards PRIMARY KEY (id),
    CONSTRAINT fk_boards_team FOREIGN KEY (team_id) REFERENCES teams(id)
);

CREATE INDEX idx_boards_team_id ON boards (team_id);

-- +goose Down
DROP TABLE IF EXISTS boards;
