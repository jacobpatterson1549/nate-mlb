CREATE TABLE IF NOT EXISTS player_types
    ( id INT PRIMARY KEY
    , sport_type_id INT
    , name VARCHAR(30) NOT NULL
    , description TEXT
    , CONSTRAINT sport_type_id_name_unique UNIQUE (sport_type_id, name)
    , FOREIGN KEY (sport_type_id) REFERENCES sport_types (id) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS get_player_types_idx ON player_types (sport_type_id, id);

INSERT INTO player_types (id, sport_type_id, name, description)
    SELECT id, sport_type_id, name, description FROM ( VALUES
      (1, 1, 'Teams', 'Wins')
    , (2, 1, 'Hitting', 'Home Runs')
    , (3, 1, 'Pitching', 'Wins')
    , (4, 2, 'Teams', 'Wins')
    , (5, 2, 'Quarterbacks', 'Touchdown (passes+runs)')
    , (6, 2, 'Misc', 'Touchdowns (RB/WR/TE) (Rushing/Receiving)')
    ) new_player_types (id, sport_type_id, name, description)
    WHERE NOT EXISTS (SELECT * FROM player_types WHERE id BETWEEN 1 AND 6)
    ;
  