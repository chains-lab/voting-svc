-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TYPE petition_status AS ENUM (
    'processed',  -- in processed by moderation
    'declined',   -- declined by moderation
    'published',  -- active and open for signatures
    'withdrawn',  -- withdrawn by the initiator
    'approved',   -- approved by addressed user
    'rejected'    -- rejected by addressed user
);

CREATE TABLE "petitions" (
    "id"            UUID            PRIMARY KEY NOT NULL,
    "city_id"       UUID            NOT NULL,
    "title"         VARCHAR(255)    NOT NULL,
    "description"   VARCHAR(8192)   NOT NULL,
    "initiator_id"  UUID            NOT NULL,
    "address_to_id" UUID, -- if NULL the petition is addressed to the city government
    "location"      GEOMETRY(Point, 4326), -- location of the petition, if NULL the petition is not location-specific
    "status"        petition_status NOT NULL,
    "signatures"    INT             NOT NULL DEFAULT 0 CHECK (signatures >= 0),
    "goal"          INT             NOT NULL DEFAULT 0 CHECK (goal >= 0),
    "end_date"      TIMESTAMP       NOT NULL,
    "created_at"    TIMESTAMP       NOT NULL,
    "updated_at"    TIMESTAMP       NOT NULL
)

CREATE TABLE "petition_signatures" (
    "id"          UUID      PRIMARY KEY NOT NULL,
    "petition_id" UUID      NOT NULL REFERENCES "petitions" ("id") ON DELETE CASCADE,
    "user_id"     UUID      NOT NULL,
    "created_at"  TIMESTAMP NOT NULL,
    UNIQUE ("petition_id", "user_id")
);

CREATE TYPE proposal_status AS ENUM (
    'processed',  -- in processed by moderation
    'declined',   -- declined by moderation
    'published',  -- active and open for signatures
    'withdrawn',  -- withdrawn by the initiator
    'approved',   -- approved by addressed user
    'rejected'    -- rejected by addressed user
);

CREATE TABLE "proposals" (
    "id"            UUID            PRIMARY KEY NOT NULL,
    "city_id"       UUID            NOT NULL,
    "title"         VARCHAR(255)    NOT NULL,
    "description"   VARCHAR(8192)   NOT NULL,
    "status"        proposal_status NOT NULL,
    "initiator_id"  UUID            NOT NULL,
    "address_to_id" UUID, -- if NULL the proposals is addressed to the city government
    "location"      GEOMETRY(Point, 4326), -- location of the proposal, if NULL the proposal is not location-specific
    "agreed_num"    INT             NOT NULL DEFAULT 0, CHECK (agreed_num >= 0),
    "disagreed_num" INT             NOT NULL DEFAULT 0, CHECK (disagreed_num >= 0),
    "end_date"      TIMESTAMP       NOT NULL,
    "created_at"    TIMESTAMP       NOT NULL,
    "updated_at"    TIMESTAMP       NOT NULL
);

CREATE TABLE "proposal_votes" (
    "id"          UUID      PRIMARY KEY NOT NULL,
    "proposal_id" UUID      NOT NULL REFERENCES "proposals" ("id") ON DELETE CASCADE,
    "user_id"     UUID      NOT NULL,
    "vote"        BOOLEAN   NOT NULL, -- TRUE for agree, FALSE for disagree
    "created_at"  TIMESTAMP NOT NULL,
    UNIQUE ("proposal_id", "user_id")
);

CREATE TYPE poll_status AS ENUM (
    'processed',  -- in processed by moderation
    'declined',   -- declined by moderation
    'published',  -- active and open for signatures
    'withdrawn',  -- withdrawn by the initiator
);

CREATE TABLE "polls" (
    "id"           UUID            PRIMARY KEY NOT NULL,
    "city_id"      UUID            NOT NULL,
    "title"        VARCHAR(255)    NOT NULL,
    "description"  VARCHAR(8192)   NOT NULL,
    "status"       poll_status     NOT NULL,
    "initiator_id" UUID            NOT NULL,
    "location"     GEOMETRY(Point, 4326), -- location of the poll, if NULL the poll is not location-specific
    "end_date"     TIMESTAMP       NOT NULL,
    "created_at"   TIMESTAMP       NOT NULL,
    "updated_at"   TIMESTAMP       NOT NULL
);

CREATE TABLE "poll_options" (
    "id"          UUID          PRIMARY KEY NOT NULL,
    "poll_id"     UUID          NOT NULL REFERENCES "polls" ("id") ON DELETE CASCADE,
    "option_text" VARCHAR(255)  NOT NULL,
    "votes_count" INT           NOT NULL DEFAULT 0 CHECK (votes_count >= 0),
    "created_at"  TIMESTAMP     NOT NULL
);

CREATE TABLE "poll_votes" (
    "id"          UUID      PRIMARY KEY NOT NULL,
    "poll_id"     UUID      NOT NULL REFERENCES "polls" ("id") ON DELETE CASCADE,
    "user_id"     UUID      NOT NULL,
    "option_id"   UUID      NOT NULL REFERENCES "poll_options" ("id") ON DELETE CASCADE,
    "created_at"  TIMESTAMP NOT NULL,
    UNIQUE ("poll_id", "user_id")
);

-- +migrate Down
DROP TABLE IF EXISTS "petition_signatures" CASCADE;
DROP TABLE IF EXISTS "petitions" CASCADE;
DROP TABLE IF EXISTS "proposal_votes" CASCADE;
DROP TABLE IF EXISTS "proposals" CASCADE;
DROP TABLE IF EXISTS "poll_votes" CASCADE;
DROP TABLE IF EXISTS "poll_options" CASCADE;
DROP TABLE IF EXISTS "polls" CASCADE;

DROP TYPE IF EXISTS petition_status CASCADE;
DROP TYPE IF EXISTS proposal_status CASCADE;
DROP TYPE IF EXISTS poll_status CASCADE;

DELETE EXTENSION IF EXISTS postgis CASCADE;
DELETE EXTENSION IF EXISTS "uuid-ossp" CASCADE;