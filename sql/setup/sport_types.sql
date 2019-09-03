CREATE TABLE IF NOT EXISTS sport_types
    ( id INT PRIMARY KEY
    , name TEXT UNIQUE NOT NULL
    , url TEXT UNIQUE NOT NULL
    );

INSERT INTO sport_types (id, name, url)
    SELECT id, name, url FROM ( VALUES
      (1, 'MLB', 'mlb')
    , (2, 'NFL', 'nfl')
    ) new_sport_types (id, name, url)
    WHERE NOT EXISTS (SELECT * FROM sport_types WHERE id BETWEEN 1 AND 2)
    ;
