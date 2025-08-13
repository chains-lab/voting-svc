-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;

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
    "agreed_num"    INT             NOT NULL DEFAULT 0 CHECK (agreed_num >= 0),
    "disagreed_num" INT             NOT NULL DEFAULT 0 CHECK (disagreed_num >= 0),
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

CREATE OR REPLACE FUNCTION sync_proposal_votes_counter()
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- новый голос
        UPDATE proposals
            SET agreed_num    = agreed_num    + CASE WHEN NEW.vote THEN 1 ELSE 0 END,
                disagreed_num = disagreed_num + CASE WHEN NEW.vote THEN 0 ELSE 1 END
            WHERE id = NEW.proposal_id;
        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        -- удаление голоса
        UPDATE proposals
            SET agreed_num    = GREATEST(agreed_num    - CASE WHEN OLD.vote THEN 1 ELSE 0 END, 0),
                disagreed_num = GREATEST(disagreed_num - CASE WHEN OLD.vote THEN 0 ELSE 1 END, 0)
            WHERE id = OLD.proposal_id;
        RETURN OLD;

    ELSIF TG_OP = 'UPDATE' THEN
        -- если меняется значение vote у того же proposal_id
        IF NEW.vote IS DISTINCT FROM OLD.vote THEN
        UPDATE proposals
            SET agreed_num    = GREATEST(agreed_num    + CASE WHEN NEW.vote THEN 1 ELSE -1 END, 0),
                disagreed_num = GREATEST(disagreed_num + CASE WHEN NEW.vote THEN -1 ELSE 1 END, 0)
            WHERE id = NEW.proposal_id;
        END IF;

        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER proposal_votes_after_ins
    AFTER INSERT ON proposal_votes
    FOR EACH ROW
    EXECUTE FUNCTION sync_proposal_votes_counter();

CREATE TRIGGER proposal_votes_after_del
    AFTER DELETE ON proposal_votes
    FOR EACH ROW
    EXECUTE FUNCTION sync_proposal_votes_counter();

CREATE TRIGGER proposal_votes_after_upd
    AFTER UPDATE OF proposal_id, vote ON proposal_votes
    FOR EACH ROW
    EXECUTE FUNCTION sync_proposal_votes_counter();

-- +migrate Down
DROP TABLE IF EXISTS "proposal_votes" CASCADE;
DROP TABLE IF EXISTS "proposals" CASCADE;

DROP TYPE IF EXISTS proposal_status;

DROP EXTENSION IF EXISTS postgis;
DROP EXTENSION IF EXISTS "uuid-ossp";