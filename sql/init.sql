-- $ sudo -u postgres psql
-- CREATE DATABASE nate_mlb_db;
-- CREATE user nate WITH ENCRYPTED PASSWORD 'Have19_each+Iowa';
-- GRANT ALL PRIVILEGES ON DATABASE nate_mlb_db to nate;
-- $ \q
-- $ PGPASSWORD=Have19_each%Iowa psql nate -h 127.0.0.1 -d nate_mlb_db

CREATE TABLE stats
        ( year INT PRIMARY KEY
        , active BOOLEAN
        , etl_timestamp TIMESTAMP
        , etl_json TEXT
        , CONSTRAINT active_true_or_null CHECK (active)
        , CONSTRAINT active_only_one UNIQUE (active)
        , CONSTRAINT valid_year CHECK (year >= 2000 AND year <= 3000)
        );

CREATE TABLE friends
	( id SERIAL PRIMARY KEY
	, name VARCHAR(20) UNIQUE NOT NULL
	, display_order INT DEFAULT 0
        , year INT NOT NULL
        , FOREIGN KEY (year) REFERENCES stats (year) ON DELETE CASCADE
	);

CREATE TABLE player_types
	( id INT PRIMARY KEY
	, name VARCHAR(20) UNIQUE NOT NULL
        , description TEXT
	);

CREATE TABLE players
	( id SERIAL PRIMARY KEY
	, player_type_id INT
	, player_id INT NOT NULL
	, friend_id INT NOT NULL
	, display_order INT DEFAULT 0
        , year INT NOT NULL
	, UNIQUE (player_type_id, player_id, friend_id)
	, FOREIGN KEY (player_type_id) REFERENCES player_types (id) ON DELETE RESTRICT
	, FOREIGN KEY (friend_id) REFERENCES friends (id) ON DELETE CASCADE
        , FOREIGN KEY (year) REFERENCES stats (year) ON DELETE CASCADE
	);

CREATE TABLE users
	( username VARCHAR(20) PRIMARY KEY
        , password TEXT
        );

INSERT INTO stats (year, active) VALUES (2019, TRUE);

INSERT INTO friends (id, name, display_order, year)
	VALUES
          (1, 'Bob',   0, 2019)
        , (2, 'W',     1, 2019)
        , (3, 'Nate',  2, 2019)
        , (4, 'Sam',   3, 2019)
        , (5, 'Steve', 4, 2019)
        , (6, 'Mike',  5, 2019)
        ;
SELECT setVal('friends_id_seq', COALESCE((SELECT MAX(id)+1 FROM friends), 1));

INSERT INTO player_types (id, name, description)
	VALUES
          (1, 'teams', 'wins')
        , (2, 'hitting', 'home runs')
        , (3, 'pitching', 'wins')
        ;

INSERT INTO players (player_type_id, player_id, friend_id, display_order, year)
	VALUES
-- teams: (-- name_display_full)
          (1, 111, 1, 0, 2019) -- Boston Red Sox
        , (1, 112, 1, 1, 2019) -- Chicago Cubs
        , (1, 142, 1, 2, 2019) -- Minnesota Twins
        , (1, 136, 1, 3, 2019) -- Seattle Mariners
        , (1, 118, 1, 4, 2019) -- Kansas City Royal
        , (1, 147, 2, 0, 2019) -- New York Yankees
        , (1, 143, 2, 1, 2019) -- Philadelphia Phillies
        , (1, 115, 2, 2, 2019) -- Colorado Rockies
        , (1, 134, 2, 3, 2019) -- Pittsburgh Pirates
        , (1, 146, 2, 4, 2019) -- Miami Marlins
        , (1, 158, 3, 0, 2019) -- Milwaukee Brewers
        , (1, 138, 3, 1, 2019) -- St. Louis Cardinals
        , (1, 108, 3, 2, 2019) -- Los Angeles Angels
        , (1, 145, 3, 3, 2019) -- Chicago White Sox
        , (1, 137, 3, 4, 2019) -- San Francisco Giants
        , (1, 117, 4, 0, 2019) -- Houston Astros
        , (1, 139, 4, 1, 2019) -- Tampa Bay Rays
        , (1, 121, 4, 2, 2019) -- New York Mets
        , (1, 141, 4, 3, 2019) -- Toronto Blue Jays
        , (1, 137, 4, 4, 2019) -- San Francisco Giants
        , (1, 119, 5, 0, 2019) -- Los Angeles Dodgers
        , (1, 144, 5, 1, 2019) -- Atlanta Braves
        , (1, 109, 5, 2, 2019) -- Arizona Diamondbacks
        , (1, 113, 5, 3, 2019) -- Cincinnati Reds
        , (1, 116, 5, 4, 2019) -- Detroit Tigers
        , (1, 114, 6, 0, 2019) -- Cleveland Indians
        , (1, 120, 6, 1, 2019) -- Washington Nationals
        , (1, 133, 6, 2, 2019) -- Oakland Athletics
        , (1, 135, 6, 3, 2019) -- San Diego Padres
        , (1, 137, 6, 4, 2019) -- San Francisco Giants
-- hitters: (-- name_display_first_last)
        , (2, 502110, 1, 0, 2019) -- J.D. Martinez
        , (2, 605141, 1, 1, 2019) -- Mookie Betts
        , (2, 608070, 1, 2, 2019) -- Jose Ramirez
        , (2, 547180, 2, 0, 2019) -- Bryce Harper
        , (2, 656555, 2, 1, 2019) -- Rhys Hoskins
        , (2, 660670, 2, 2, 2019) -- Ronald Acuna Jr.
        , (2, 545361, 3, 0, 2019) -- Mike Trout
        , (2, 571448, 3, 1, 2019) -- Nolan Arenado
        , (2, 592518, 3, 2, 2019) -- Manny Machado
        , (2, 501981, 4, 0, 2019) -- Khris Davis
        , (2, 608336, 4, 1, 2019) -- Joey Gallo
        , (2, 596019, 4, 2, 2019) -- Francisco Lindor
        , (2, 519317, 5, 0, 2019) -- Giancarlo Stanton
        , (2, 429665, 5, 1, 2019) -- Edwin Encarnacion
        , (2, 592885, 5, 2, 2019) -- Christian Yelich
        , (2, 592450, 6, 0, 2019) -- Aaron Judge
        , (2, 596115, 6, 1, 2019) -- Trevor Story
        , (2, 502671, 6, 2, 2019) -- Paul Goldschmidt
-- pitchers:
        , (3, 605483, 1, 0, 2019) -- Blake Snell
        , (3, 622663, 1, 1, 2019) -- Luis Severino
        , (3, 605400, 1, 2, 2019) -- Aaron Nola
        , (3, 594798, 2, 0, 2019) -- Jacob deGrom
        , (3, 621111, 2, 1, 2019) -- Walker Buehler
        , (3, 572020, 2, 2, 2019) -- James Paxton
        , (3, 453286, 3, 0, 2019) -- Max Scherzer
        , (3, 519144, 3, 1, 2019) -- Rick Porcello
        , (3, 605400, 3, 2, 2019) -- Aaron Nola
        , (3, 446372, 4, 0, 2019) -- Corey Kluber
        , (3, 471911, 4, 1, 2019) -- Carlos Carrasco
        , (3, 452657, 4, 2, 2019) -- Jon Lester
        , (3, 519242, 5, 0, 2019) -- Chris Sale
        , (3, 456034, 5, 1, 2019) -- David Price
        , (3, 592789, 5, 2, 2019) -- Noah Syndergaard
        , (3, 434378, 6, 0, 2019) -- Justin Verlander
        , (3, 543037, 6, 1, 2019) -- Gerrit Cole
        , (3, 545333, 6, 2, 2019) -- Trevor Bauer
        ;
SELECT setVal('players_id_seq', COALESCE((SELECT MAX(id)+1 FROM players), 1));

INSERT INTO users (username, password)
        VALUES
          ('admin', 'invalid_hash_value')
        ;
