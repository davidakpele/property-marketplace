CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TYPE listing_type AS ENUM ('rent', 'sale', 'shortlet');

CREATE TABLE IF NOT EXISTS listings (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title        VARCHAR(500) NOT NULL,
    description  TEXT,
    price        NUMERIC(15, 2) NOT NULL,
    type         listing_type NOT NULL,
    bedrooms     SMALLINT NOT NULL CHECK (bedrooms >= 0),
    address      VARCHAR(500) NOT NULL,
    latitude     DOUBLE PRECISION NOT NULL,
    longitude    DOUBLE PRECISION NOT NULL,
    location     GEOGRAPHY(POINT, 4326) GENERATED ALWAYS AS (
                     ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography
                 ) STORED,
    agent_id     UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listings_agent_id   ON listings(agent_id);
CREATE INDEX IF NOT EXISTS idx_listings_type        ON listings(type);
CREATE INDEX IF NOT EXISTS idx_listings_price       ON listings(price);
CREATE INDEX IF NOT EXISTS idx_listings_bedrooms    ON listings(bedrooms);
CREATE INDEX IF NOT EXISTS idx_listings_location    ON listings USING GIST(location);
