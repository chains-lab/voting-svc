-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE issuances (
    id          UUID      PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    election_id UUID      NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    issued_at   TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT uniq_issuance_per_user_per_election
       UNIQUE (user_id, election_id)
);

CREATE TABLE votes (
    id          UUID      PRIMARY KEY DEFAULT uuid_generate_v4(),
    election_id UUID      NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    nullifier   BYTEA     NOT NULL,
    choice      JSONB     NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT uniq_vote_per_election_nullifier
       UNIQUE (election_id, nullifier)
);

