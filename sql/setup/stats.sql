CREATE TABLE IF NOT EXISTS stats
    ( id SERIAL PRIMARY KEY
    , sport_type_id INT
    , year INT NOT NULL
    , active BOOLEAN
    , etl_timestamp TIMESTAMP
    , etl_json TEXT
    , CONSTRAINT sport_year_unique UNIQUE (sport_type_id, year)
    , CONSTRAINT active_true_or_null CHECK (active)
    , CONSTRAINT active_only_one UNIQUE (active, sport_type_id)
    , CONSTRAINT valid_year CHECK (year >= 2000 AND year <= 3000)
    , FOREIGN KEY (sport_type_id) REFERENCES sport_types (id) ON DELETE RESTRICT
    );

CREATE INDEX IF NOT EXISTS get_years_idx on stats (year);

CREATE INDEX IF NOT EXISTS get_active_year_idx on stats (sport_type_id) WHERE active;
