CREATE TABLE IF NOT EXISTS stats
    ( id SERIAL PRIMARY KEY
    , sport_type_id INT NOT NULL
    , year INT NOT NULL
    , active BOOLEAN
    , etl_timestamp TIMESTAMP
    , etl_json JSONB
    , CONSTRAINT sport_year_unique UNIQUE (sport_type_id, year)
    , CONSTRAINT active_true_or_null CHECK (active)
    , CONSTRAINT active_only_one UNIQUE (active, sport_type_id)
    , CONSTRAINT valid_year CHECK (year >= 2000 AND year <= 3000)
    , FOREIGN KEY (sport_type_id) REFERENCES sport_types (id) ON DELETE RESTRICT
    );

CREATE INDEX IF NOT EXISTS get_active_year_idx ON stats (sport_type_id) WHERE active;

CREATE INDEX IF NOT EXISTS get_years_idx ON stats (sport_type_id, year);
