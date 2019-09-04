CREATE TABLE IF NOT EXISTS players
    ( id SERIAL PRIMARY KEY
    , player_type_id INT
    , source_id INT NOT NULL
    , friend_id INT NOT NULL
    , display_order INT DEFAULT 0 NOT NULL
    , CONSTRAINT player_type_id_source_id_friend_id_unique UNIQUE (player_type_id, source_id, friend_id)
    , FOREIGN KEY (player_type_id) REFERENCES player_types (id) ON DELETE RESTRICT
    , FOREIGN KEY (friend_id) REFERENCES friends (id) ON DELETE CASCADE
    );

--TODO: Is get_players_idx used?
CREATE INDEX IF NOT EXISTS get_players_idx ON players (player_type_id, friend_id, display_order);
