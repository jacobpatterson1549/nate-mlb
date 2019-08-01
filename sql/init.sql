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
	( id SERIAL PRIMARY KEY
	, name VARCHAR(20) UNIQUE NOT NULL
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

INSERT INTO key_store (k, v)
        VALUES
          ('admin', 'invalid_hssh_value')
          ('etl', '{}')
        ;

INSERT INTO player_types (id, name)
	VALUES
          (1, 'teams')
        , (2, 'hitting')
        , (3, 'pitching')
        ;

INSERT INTO friends (id, name, display_order)
	VALUES
          (1, 'Bob', 1)
        , (2, 'W', 2)
        , (3, 'Nate', 3)
        , (4, 'Sam', 4)
        , (5, 'Steve', 5)
        , (6, 'Mike', 6)
        ;

INSERT INTO players (player_type_id, player_id, friend_id, display_order)
	VALUES
-- teams: (-- name_display_full)
          (1, 111, 1, 1) -- Boston Red Sox
        , (1, 112, 1, 2) -- Chicago Cubs
        , (1, 142, 1, 3) -- Minnesota Twins
        , (1, 136, 1, 4) -- Seattle Mariners
        , (1, 118, 1, 5) -- Kansas City Royal
        , (1, 147, 2, 1) -- New York Yankees
        , (1, 143, 2, 2) -- Philadelphia Phillies
        , (1, 115, 2, 3) -- Colorado Rockies
        , (1, 134, 2, 4) -- Pittsburgh Pirates
        , (1, 146, 2, 5) -- Miami Marlins
        , (1, 158, 3, 1) -- Milwaukee Brewers
        , (1, 138, 3, 2) -- St. Louis Cardinals
        , (1, 108, 3, 3) -- Los Angeles Angels
        , (1, 145, 3, 4) -- Chicago White Sox
        , (1, 137, 3, 5) -- San Francisco Giants
        , (1, 117, 4, 1) -- Houston Astros
        , (1, 139, 4, 2) -- Tampa Bay Rays
        , (1, 121, 4, 3) -- New York Mets
        , (1, 141, 4, 4) -- Toronto Blue Jays
        , (1, 137, 4, 5) -- San Francisco Giants
        , (1, 119, 5, 1) -- Los Angeles Dodgers
        , (1, 144, 5, 2) -- Atlanta Braves
        , (1, 109, 5, 3) -- Arizona Diamondbacks
        , (1, 113, 5, 4) -- Cincinnati Reds
        , (1, 116, 5, 5) -- Detroit Tigers
        , (1, 114, 6, 1) -- Cleveland Indians
        , (1, 120, 6, 2) -- Washington Nationals
        , (1, 133, 6, 3) -- Oakland Athletics
        , (1, 135, 6, 4) -- San Diego Padres
        , (1, 137, 6, 5) -- San Francisco Giants
-- hitters: (-- name_display_first_last)
        , (2, 502110, 1, 1) -- J.D. Martinez
        , (2, 605141, 1, 2) -- Mookie Betts
        , (2, 608070, 1, 3) -- Jose Ramirez
        , (2, 547180, 2, 1) -- Bryce Harper
        , (2, 656555, 2, 2) -- Rhys Hoskins
        , (2, 660670, 2, 3) -- Ronald Acuna Jr.
        , (2, 545361, 3, 1) -- Mike Trout
        , (2, 571448, 3, 2) -- Nolan Arenado
        , (2, 592518, 3, 3) -- Manny Machado
        , (2, 501981, 4, 1) -- Khris Davis
        , (2, 608336, 4, 2) -- Joey Gallo
        , (2, 596019, 4, 3) -- Francisco Lindor
        , (2, 519317, 5, 1) -- Giancarlo Stanton
        , (2, 429665, 5, 2) -- Edwin Encarnacion
        , (2, 592885, 5, 3) -- Christian Yelich
        , (2, 592450, 6, 1) -- Aaron Judge
        , (2, 596115, 6, 2) -- Trevor Story
        , (2, 502671, 6, 3) -- Paul Goldschmidt
-- pitchers:
        , (3, 605483, 1, 1) -- Blake Snell
        , (3, 622663, 1, 2) -- Luis Severino
        , (3, 605400, 1, 3) -- Aaron Nola
        , (3, 594798, 2, 1) -- Jacob deGrom
        , (3, 621111, 2, 2) -- Walker Buehler
        , (3, 572020, 2, 3) -- James Paxton
        , (3, 453286, 3, 1) -- Max Scherzer
        , (3, 519144, 3, 2) -- Rick Porcello
        , (3, 605400, 3, 3) -- Aaron Nola
        , (3, 446372, 4, 1) -- Corey Kluber
        , (3, 471911, 4, 2) -- Carlos Carrasco
        , (3, 452657, 4, 3) -- Jon Lester
        , (3, 519242, 5, 1) -- Chris Sale
        , (3, 456034, 5, 2) -- David Price
        , (3, 592789, 5, 3) -- Noah Syndergaard
        , (3, 434378, 6, 1) -- Justin Verlander
        , (3, 543037, 6, 2) -- Gerrit Cole
        , (3, 545333, 6, 3) -- Trevor Bauer
        ;
