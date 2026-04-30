-- +goose Up
CREATE TABLE members (
    id              CHAR(26)        NOT NULL,
    team_id         CHAR(26)        NOT NULL,
    name            VARCHAR(100)    NOT NULL,
    email           VARCHAR(255)    NOT NULL,
    role            VARCHAR(20)     NOT NULL DEFAULT 'member',
    api_key_hash    VARCHAR(255)    NOT NULL,
    api_key_prefix  VARCHAR(10)     NOT NULL,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ     NULL,

    CONSTRAINT pk_members PRIMARY KEY (id),
    CONSTRAINT fk_members_team FOREIGN KEY (team_id) REFERENCES teams(id),
    CONSTRAINT chk_members_role CHECK (role IN ('admin', 'member'))
);

CREATE UNIQUE INDEX idx_members_api_key_hash ON members (api_key_hash) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_members_team_email ON members (team_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_members_team_id ON members (team_id);

-- +goose Down
DROP TABLE IF EXISTS members;
