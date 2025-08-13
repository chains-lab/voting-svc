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
);

CREATE TABLE "petition_signatures" (
    "id"          UUID      PRIMARY KEY NOT NULL,
    "petition_id" UUID      NOT NULL REFERENCES "petitions" ("id") ON DELETE CASCADE,
    "user_id"     UUID      NOT NULL,
    "created_at"  TIMESTAMP NOT NULL,
    UNIQUE ("petition_id", "user_id")
);

CREATE OR REPLACE FUNCTION sync_petition_signatures_counter()
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE petitions
            SET signatures = signatures + 1
            WHERE id = NEW.petition_id;
        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        UPDATE petitions
            SET signatures = GREATEST(signatures - 1, 0)
            WHERE id = OLD.petition_id;
        RETURN OLD;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER petition_signatures_after_ins
    AFTER INSERT ON petition_signatures
    FOR EACH ROW
    EXECUTE FUNCTION sync_petition_signatures_counter();

CREATE TRIGGER petition_signatures_after_del
    AFTER DELETE ON petition_signatures
    FOR EACH ROW
    EXECUTE FUNCTION sync_petition_signatures_counter();


-- +migrate Down
DROP TABLE IF EXISTS "petition_signatures" CASCADE;
DROP TABLE IF EXISTS "petitions" CASCADE;

DROP TYPE IF EXISTS petition_status CASCADE;

DROP EXTENSION IF EXISTS postgis CASCADE;
DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;