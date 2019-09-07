CREATE TABLE IF NOT EXISTS friends
    ( id SERIAL PRIMARY KEY
    , name VARCHAR(20) NOT NULL
    , display_order INT DEFAULT 0 NOT NULL
    , stat_id INT NOT NULL
    , CONSTRAINT name_stat_id UNIQUE (name, stat_id)
    , FOREIGN KEY (stat_id) REFERENCES stats (id) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS get_friends_idx ON friends (stat_id, display_order);
