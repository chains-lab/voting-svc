-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TYPE poll_status AS ENUM (
    'processed',  -- in processed by moderation
    'declined',   -- declined by moderation
    'published',  -- active and open for signatures
    'withdrawn'   -- withdrawn by the initiator
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

CREATE OR REPLACE FUNCTION sync_poll_votes_counter()
RETURNS trigger AS $$
BEGIN
	IF TG_OP = 'INSERT' THEN
		-- новый голос: +1 к выбранной опции
        UPDATE poll_options
            SET votes_count = votes_count + 1
            WHERE id = NEW.option_id;
        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        -- удаление голоса: -1 у старой опции
        UPDATE poll_options
            SET votes_count = GREATEST(votes_count - 1, 0)
            WHERE id = OLD.option_id;
        RETURN OLD;

    ELSIF TG_OP = 'UPDATE' THEN
        -- если поменяли option_id (и/или poll_id) — переложить счётчики
        IF NEW.option_id IS DISTINCT FROM OLD.option_id THEN
            UPDATE poll_options
                SET votes_count = GREATEST(votes_count - 1, 0)
                WHERE id = OLD.option_id;

            UPDATE poll_options
                SET votes_count = votes_count + 1
                WHERE id = NEW.option_id;
        END IF;

        RETURN NEW;
    END IF;

RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER poll_votes_after_ins
    AFTER INSERT ON poll_votes
    FOR EACH ROW
    EXECUTE FUNCTION sync_poll_votes_counter();

CREATE TRIGGER poll_votes_after_del
    AFTER DELETE ON poll_votes
    FOR EACH ROW
    EXECUTE FUNCTION sync_poll_votes_counter();

CREATE TRIGGER poll_votes_after_upd
    AFTER UPDATE OF poll_id, option_id ON poll_votes
    FOR EACH ROW
    EXECUTE FUNCTION sync_poll_votes_counter();

-- +migrate Down
DROP TABLE IF EXISTS "poll_votes" CASCADE;
DROP TABLE IF EXISTS "poll_options" CASCADE;
DROP TABLE IF EXISTS "polls" CASCADE;

DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS postgis;