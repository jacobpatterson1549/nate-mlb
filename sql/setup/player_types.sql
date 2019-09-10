CREATE TABLE IF NOT EXISTS player_types
    ( id INT PRIMARY KEY
    , sport_type_id INT NOT NULL
    , name VARCHAR(30) NOT NULL
    , description TEXT
    , score_type VARCHAR(30) NOT NULL
    , CONSTRAINT sport_type_id_name_unique UNIQUE (sport_type_id, name)
    , FOREIGN KEY (sport_type_id) REFERENCES sport_types (id) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS get_player_types_idx ON player_types (sport_type_id, id);

INSERT INTO player_types (id, sport_type_id, name, description, score_type)
    SELECT id, sport_type_id, name, description, score_type FROM ( VALUES
      (1, 1, 'Teams', 'Wins', 'Wins')
    , (2, 1, 'Hitting', 'Home Runs', 'HRs')
    , (3, 1, 'Pitching', 'Wins', 'Wins')
    , (4, 2, 'Teams', 'Wins', 'Wins')
    , (5, 2, 'Quarterbacks', 'Touchdown (passes+runs)', 'TDs')
    , (6, 2, 'Misc', 'Touchdowns (RB/WR/TE) (Rushing/Receiving)', 'TDs')
    ) new_player_types (id, sport_type_id, name, description, score_type)
    WHERE NOT EXISTS (SELECT * FROM player_types WHERE id BETWEEN 1 AND 6)
    ;
  