-- $ sudo -u postgres psql
-- CREATE DATABASE nate_mlb_db;
-- CREATE user nate WITH ENCRYPTED PASSWORD 'Have19_each+Iowa';
-- GRANT ALL PRIVILEGES ON DATABASE nate_mlb_db to nate;
-- $ \q
-- $ PGPASSWORD=Have19_each%Iowa psql nate -h 127.0.0.1 -d nate_mlb_db

CREATE TABLE IF NOT EXISTS friends
	( id SERIAL PRIMARY KEY
	, name VARCHAR(20) UNIQUE NOT NULL
	, display_order INT DEFAULT 0
	);

CREATE TABLE IF NOT EXISTS player_types
	( id INT PRIMARY KEY
	, name VARCHAR(20) UNIQUE NOT NULL
        , description TEXT
	);

CREATE TABLE IF NOT EXISTS players
	( id SERIAL PRIMARY KEY
	, player_type_id INT
	, player_id INT NOT NULL
	, friend_id INT NOT NULL
	, display_order INT DEFAULT 0
	, UNIQUE (player_type_id, player_id, friend_id)
	, FOREIGN KEY (player_type_id) REFERENCES player_types (id)
	, FOREIGN KEY (friend_id) REFERENCES friends (id)
	);

CREATE TABLE IF NOT EXISTS key_store
	( k VARCHAR(20) PRIMARY KEY
        , v TEXT
        );

CREATE TABLE IF NOT EXISTS stats
        ( year INT PRIMARY KEY
        , active BOOLEAN
        , etl_timestamp TIMESTAMP
        , etl_json TEXT
        , CONSTRAINT active_true_or_null CHECK (active)
        , CONSTRAINT active_only_one UNIQUE (active)
        );

INSERT INTO stats (year) VALUES (2019);

INSERT INTO key_store (k, v)
        VALUES
          ('admin', 'invalid_hash_value')
        ;

INSERT INTO player_types (id, name, description)
	VALUES
          (1, 'teams', 'sum of team wins')
        , (2, 'hitting' 'sum of top two batters'' home run counts')
        , (3, 'pitching', 'sum of top two pitchers'' win counts')
        ;

INSERT INTO friends (id, name, display_order)
	VALUES
          (1, 'Bob',   0)
        , (2, 'W',     1)
        , (3, 'Nate',  2)
        , (4, 'Sam',   3)
        , (5, 'Steve', 4)
        , (6, 'Mike',  5)
        ;

INSERT INTO players (player_type_id, player_id, friend_id, display_order)
	VALUES
-- teams: (-- name_display_full)
          (1, 111, 1, 0) -- Boston Red Sox
        , (1, 112, 1, 1) -- Chicago Cubs
        , (1, 142, 1, 2) -- Minnesota Twins
        , (1, 136, 1, 3) -- Seattle Mariners
        , (1, 118, 1, 4) -- Kansas City Royal
        , (1, 147, 2, 0) -- New York Yankees
        , (1, 143, 2, 1) -- Philadelphia Phillies
        , (1, 115, 2, 2) -- Colorado Rockies
        , (1, 134, 2, 3) -- Pittsburgh Pirates
        , (1, 146, 2, 4) -- Miami Marlins
        , (1, 158, 3, 0) -- Milwaukee Brewers
        , (1, 138, 3, 1) -- St. Louis Cardinals
        , (1, 108, 3, 2) -- Los Angeles Angels
        , (1, 145, 3, 3) -- Chicago White Sox
        , (1, 137, 3, 4) -- San Francisco Giants
        , (1, 117, 4, 0) -- Houston Astros
        , (1, 139, 4, 1) -- Tampa Bay Rays
        , (1, 121, 4, 2) -- New York Mets
        , (1, 141, 4, 3) -- Toronto Blue Jays
        , (1, 137, 4, 4) -- San Francisco Giants
        , (1, 119, 5, 0) -- Los Angeles Dodgers
        , (1, 144, 5, 1) -- Atlanta Braves
        , (1, 109, 5, 2) -- Arizona Diamondbacks
        , (1, 113, 5, 3) -- Cincinnati Reds
        , (1, 116, 5, 4) -- Detroit Tigers
        , (1, 114, 6, 0) -- Cleveland Indians
        , (1, 120, 6, 1) -- Washington Nationals
        , (1, 133, 6, 2) -- Oakland Athletics
        , (1, 135, 6, 3) -- San Diego Padres
        , (1, 137, 6, 4) -- San Francisco Giants
-- hitters: (-- name_display_first_last)
        , (2, 502110, 1, 0) -- J.D. Martinez
        , (2, 605141, 1, 1) -- Mookie Betts
        , (2, 608070, 1, 2) -- Jose Ramirez
        , (2, 547180, 2, 0) -- Bryce Harper
        , (2, 656555, 2, 1) -- Rhys Hoskins
        , (2, 660670, 2, 2) -- Ronald Acuna Jr.
        , (2, 545361, 3, 0) -- Mike Trout
        , (2, 571448, 3, 1) -- Nolan Arenado
        , (2, 592518, 3, 2) -- Manny Machado
        , (2, 501981, 4, 0) -- Khris Davis
        , (2, 608336, 4, 1) -- Joey Gallo
        , (2, 596019, 4, 2) -- Francisco Lindor
        , (2, 519317, 5, 0) -- Giancarlo Stanton
        , (2, 429665, 5, 1) -- Edwin Encarnacion
        , (2, 592885, 5, 2) -- Christian Yelich
        , (2, 592450, 6, 0) -- Aaron Judge
        , (2, 596115, 6, 1) -- Trevor Story
        , (2, 502671, 6, 2) -- Paul Goldschmidt
-- pitchers:
        , (3, 605483, 1, 0) -- Blake Snell
        , (3, 622663, 1, 1) -- Luis Severino
        , (3, 605400, 1, 2) -- Aaron Nola
        , (3, 594798, 2, 0) -- Jacob deGrom
        , (3, 621111, 2, 1) -- Walker Buehler
        , (3, 572020, 2, 2) -- James Paxton
        , (3, 453286, 3, 0) -- Max Scherzer
        , (3, 519144, 3, 1) -- Rick Porcello
        , (3, 605400, 3, 2) -- Aaron Nola
        , (3, 446372, 4, 0) -- Corey Kluber
        , (3, 471911, 4, 1) -- Carlos Carrasco
        , (3, 452657, 4, 2) -- Jon Lester
        , (3, 519242, 5, 0) -- Chris Sale
        , (3, 456034, 5, 1) -- David Price
        , (3, 592789, 5, 2) -- Noah Syndergaard
        , (3, 434378, 6, 0) -- Justin Verlander
        , (3, 543037, 6, 1) -- Gerrit Cole
        , (3, 545333, 6, 2) -- Trevor Bauer
        ;
