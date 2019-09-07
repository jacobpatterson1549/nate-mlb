CREATE TABLE IF NOT EXISTS friends
    ( id SERIAL PRIMARY KEY
    , name VARCHAR(20) NOT NULL
    , display_order INT DEFAULT 0 NOT NULL
    , sport_type_id INT NOT NULL
    , year INT NOT NULL
    , CONSTRAINT name_sport_type_id_year_unique UNIQUE (name, sport_type_id, year)
    , FOREIGN KEY (sport_type_id, year) REFERENCES stats (sport_type_id, year) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS get_friends_idx ON friends (sport_type_id, year, display_order);
