CREATE TABLE IF NOT EXISTS stats
    ( id SERIAL PRIMARY KEY
    , sport_type_id INT
    , year INT NOT NULL
    , active BOOLEAN
    , etl_timestamp TIMESTAMP
    , etl_json JSON
    , CONSTRAINT sport_year_unique UNIQUE (sport_type_id, year)
    , CONSTRAINT active_true_or_null CHECK (active)
    , CONSTRAINT active_only_one UNIQUE (active, sport_type_id)
    , CONSTRAINT valid_year CHECK (year >= 2000 AND year <= 3000)
    , FOREIGN KEY (sport_type_id) REFERENCES sport_types (id) ON DELETE RESTRICT
    );

CREATE INDEX IF NOT EXISTS get_active_year_idx ON stats (sport_type_id) WHERE active;

CREATE INDEX IF NOT EXISTS get_years_idx ON stats (sport_type_id, year);

-- an active year is required for all SportTypes
INSERT INTO stats (sport_type_id, year, active)
    SELECT sport_type_id, year, active FROM ( VALUES
      (1, 2019, TRUE)
    , (2, 2019, TRUE)
    ) new_stats (sport_type_id, year, active)
    WHERE NOT EXISTS (SELECT * FROM stats WHERE sport_type_id BETWEEN 1 AND 2)
    ;
