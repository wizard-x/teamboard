-- +goose Up
CREATE TABLE teams (
    id           CHAR(26)        NOT NULL,
    name         VARCHAR(100)    NOT NULL,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ     NULL,

    CONSTRAINT pk_teams PRIMARY KEY (id)
);

CREATE UNIQUE INDEX idx_teams_name ON teams (name) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS teams;
