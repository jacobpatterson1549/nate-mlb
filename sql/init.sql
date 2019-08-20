-- $ sudo -u postgres psql
-- CREATE DATABASE nate_mlb_db;
-- CREATE user nate WITH ENCRYPTED PASSWORD 'Have19_each+Iowa';
-- GRANT ALL PRIVILEGES ON DATABASE nate_mlb_db to nate;
-- \q
-- $ PGPASSWORD=Have19_each%Iowa psql nate -h 127.0.0.1 -d nate_mlb_db

--DROP TABLE sport_types, stats, friends, player_types, players;--, users;

CREATE TABLE sport_types
	( id SERIAL PRIMARY KEY
	, name TEXT UNIQUE NOT NULL
	);

CREATE TABLE stats
	( id SERIAL PRIMARY KEY
	, sport_type_id INT
	, year INT NOT NULL
	, active BOOLEAN
	, etl_timestamp TIMESTAMP
	, etl_json TEXT
	, CONSTRAINT sport_year_unique UNIQUE (sport_type_id, year)
	, CONSTRAINT active_true_or_null CHECK (active)
	, CONSTRAINT active_only_one UNIQUE (active, sport_type_id)
	, CONSTRAINT valid_year CHECK (year >= 2000 AND year <= 3000)
	, FOREIGN KEY (sport_type_id) REFERENCES sport_types (id) ON DELETE RESTRICT
	);

CREATE TABLE friends
	( id SERIAL PRIMARY KEY
	, name VARCHAR(20) NOT NULL
	, display_order INT DEFAULT 0 NOT NULL
	, sport_type_id INT
	, year INT NOT NULL
	, CONSTRAINT name_year_unique UNIQUE (year, name)
	, FOREIGN KEY (sport_type_id, year) REFERENCES stats (sport_type_id, year) ON DELETE CASCADE
	);

CREATE TABLE player_types
	( id INT PRIMARY KEY
	, sport_type_id INT
	, name VARCHAR(30) NOT NULL
	, description TEXT
	, FOREIGN KEY (sport_type_id) REFERENCES sport_types (id) ON DELETE CASCADE
	, CONSTRAINT sport_type_id_name_unique UNIQUE (sport_type_id, name)
	);

CREATE TABLE players
	( id SERIAL PRIMARY KEY
	, player_type_id INT
	, player_id INT NOT NULL
	, friend_id INT NOT NULL
	, display_order INT DEFAULT 0
	, sport_type_id INT
	, year INT NOT NULL
	, CONSTRAINT player_type_id_player_id_friend_id_unique UNIQUE (player_type_id, player_id, friend_id)
	, FOREIGN KEY (player_type_id) REFERENCES player_types (id) ON DELETE RESTRICT
	, FOREIGN KEY (friend_id) REFERENCES friends (id) ON DELETE CASCADE
	, FOREIGN KEY (sport_type_id, year) REFERENCES stats (sport_type_id, year) ON DELETE CASCADE
	);

CREATE TABLE users
	( username VARCHAR(20) PRIMARY KEY
	, password TEXT
	);

INSERT INTO sport_types (id, name) VALUES (
	   1, 'mlb')
	, (2, 'nfl')
	;
SELECT setVal('sport_types_id_seq', COALESCE((SELECT MAX(id)+1 FROM sport_types), 1));

INSERT INTO stats (sport_type_id, year, active)
	VALUES (
	   1, 2019, TRUE)
	, (2, 2018, TRUE))
INSERT INTO friends (id, name, display_order, sport_type_id, year)
	VALUES (
	   1, 'Bob',   0, 1, 2019)
	, (2, 'W',     1, 1, 2019)
	, (3, 'Nate',  2, 1, 2019)
	, (4, 'Sam',   3, 1, 2019)
	, (5, 'Steve', 4, 1, 2019)
	, (6, 'Mike',  5, 1, 2019)
	, ( 7, 'Mark',   0, 2, 2018)
	, ( 8, 'Nate',   1, 2, 2018)
	, ( 9, 'Warren', 2, 2, 2018)
	, (10, 'Viet',   3, 2, 2018)
	, (11, 'Sam',    4, 2, 2018)

	;
SELECT setVal('friends_id_seq', COALESCE((SELECT MAX(id)+1 FROM friends), 1));

INSERT INTO player_types (id, sport_type_id, name, description)
	VALUES (
	   1, 1, 'Teams', 'Wins')
	, (2, 1, 'Hitting', 'Home Runs')
	, (3, 1, 'Pitching', 'Wins')
	, (4, 2, 'Teams', 'Wins')
	, (5, 2, 'Quarterbacks', 'Touchdown passes + runs')
	, (6, 2, 'Runningbacks & Wide Recievers', 'Touchdowns')
	;

INSERT INTO players (player_type_id, player_id, friend_id, display_order, sport_type_id, year)
	VALUES (
-- teams:  name_display_full
	   1, 111, 1, 0, 1, 2019) -- Boston Red Sox
	, (1, 112, 1, 1, 1, 2019) -- Chicago Cubs
	, (1, 142, 1, 2, 1, 2019) -- Minnesota Twins
	, (1, 136, 1, 3, 1, 2019) -- Seattle Mariners
	, (1, 118, 1, 4, 1, 2019) -- Kansas City Royal
	, (1, 147, 2, 0, 1, 2019) -- New York Yankees
	, (1, 143, 2, 1, 1, 2019) -- Philadelphia Phillies
	, (1, 115, 2, 2, 1, 2019) -- Colorado Rockies
	, (1, 134, 2, 3, 1, 2019) -- Pittsburgh Pirates
	, (1, 146, 2, 4, 1, 2019) -- Miami Marlins
	, (1, 158, 3, 0, 1, 2019) -- Milwaukee Brewers
	, (1, 138, 3, 1, 1, 2019) -- St. Louis Cardinals
	, (1, 108, 3, 2, 1, 2019) -- Los Angeles Angels
	, (1, 145, 3, 3, 1, 2019) -- Chicago White Sox
	, (1, 137, 3, 4, 1, 2019) -- San Francisco Giants
	, (1, 117, 4, 0, 1, 2019) -- Houston Astros
	, (1, 139, 4, 1, 1, 2019) -- Tampa Bay Rays
	, (1, 121, 4, 2, 1, 2019) -- New York Mets
	, (1, 141, 4, 3, 1, 2019) -- Toronto Blue Jays
	, (1, 137, 4, 4, 1, 2019) -- San Francisco Giants
	, (1, 119, 5, 0, 1, 2019) -- Los Angeles Dodgers
	, (1, 144, 5, 1, 1, 2019) -- Atlanta Braves
	, (1, 109, 5, 2, 1, 2019) -- Arizona Diamondbacks
	, (1, 113, 5, 3, 1, 2019) -- Cincinnati Reds
	, (1, 116, 5, 4, 1, 2019) -- Detroit Tigers
	, (1, 114, 6, 0, 1, 2019) -- Cleveland Indians
	, (1, 120, 6, 1, 1, 2019) -- Washington Nationals
	, (1, 133, 6, 2, 1, 2019) -- Oakland Athletics
	, (1, 135, 6, 3, 1, 2019) -- San Diego Padres
	, (1, 137, 6, 4, 1, 2019) -- San Francisco Giants
-- hitters: name_display_first_last
	, (2, 502110, 1, 0, 1, 2019) -- J.D. Martinez
	, (2, 605141, 1, 1, 1, 2019) -- Mookie Betts
	, (2, 608070, 1, 2, 1, 2019) -- Jose Ramirez
	, (2, 547180, 2, 0, 1, 2019) -- Bryce Harper
	, (2, 656555, 2, 1, 1, 2019) -- Rhys Hoskins
	, (2, 660670, 2, 2, 1, 2019) -- Ronald Acuna Jr.
	, (2, 545361, 3, 0, 1, 2019) -- Mike Trout
	, (2, 571448, 3, 1, 1, 2019) -- Nolan Arenado
	, (2, 592518, 3, 2, 1, 2019) -- Manny Machado
	, (2, 501981, 4, 0, 1, 2019) -- Khris Davis
	, (2, 608336, 4, 1, 1, 2019) -- Joey Gallo
	, (2, 596019, 4, 2, 1, 2019) -- Francisco Lindor
	, (2, 519317, 5, 0, 1, 2019) -- Giancarlo Stanton
	, (2, 429665, 5, 1, 1, 2019) -- Edwin Encarnacion
	, (2, 592885, 5, 2, 1, 2019) -- Christian Yelich
	, (2, 592450, 6, 0, 1, 2019) -- Aaron Judge
	, (2, 596115, 6, 1, 1, 2019) -- Trevor Story
	, (2, 502671, 6, 2, 1, 2019) -- Paul Goldschmidt
-- pitchers: name_display_first_last
	, (3, 605483, 1, 0, 1, 2019) -- Blake Snell
	, (3, 622663, 1, 1, 1, 2019) -- Luis Severino
	, (3, 605400, 1, 2, 1, 2019) -- Aaron Nola
	, (3, 594798, 2, 0, 1, 2019) -- Jacob deGrom
	, (3, 621111, 2, 1, 1, 2019) -- Walker Buehler
	, (3, 572020, 2, 2, 1, 2019) -- James Paxton
	, (3, 453286, 3, 0, 1, 2019) -- Max Scherzer
	, (3, 519144, 3, 1, 1, 2019) -- Rick Porcello
	, (3, 605400, 3, 2, 1, 2019) -- Aaron Nola
	, (3, 446372, 4, 0, 1, 2019) -- Corey Kluber
	, (3, 471911, 4, 1, 1, 2019) -- Carlos Carrasco
	, (3, 452657, 4, 2, 1, 2019) -- Jon Lester
	, (3, 519242, 5, 0, 1, 2019) -- Chris Sale
	, (3, 456034, 5, 1, 1, 2019) -- David Price
	, (3, 592789, 5, 2, 1, 2019) -- Noah Syndergaard
	, (3, 434378, 6, 0, 1, 2019) -- Justin Verlander
	, (3, 543037, 6, 1, 1, 2019) -- Gerrit Cole
	, (3, 545333, 6, 2, 1, 2019) -- Trevor Bauer
-- nfl teams:
	, (4, 21,  7, 0) -- New England Patriots
	, (4, 30,  7, 1) -- Seattle Seahawks
	, (4,  1,  7, 2) -- Atlanta Falcons
	, (4, 10,  7, 3) -- Detroit Lions
	, (4,  6,  7, 4) -- Cincinnati Bengals
	, (4, 24,  7, 5) -- New York Jets
	, (4, 20,  8, 0) -- Minnesota Vikings
	, (4, 27,  8, 1) -- Pittsburgh Steelers
	, (4,  4,  8, 2) -- Carolina Panthers
	, (4, 29,  8, 3) -- San Francisco 49ers
	, (4,  5,  8, 4) -- Chicago Bears
	, (4,  3,  8, 5) -- Buffalo Bills
	, (4, 15,  9, 0) -- Jacksonville Jaguars
	, (4, 22,  9, 1) -- New Orleans Saints
	, (4, 12,  9, 2) -- Tennessee Titans
	, (4,  9,  9, 3) -- Denver Broncos
	, (4, 31,  9, 4) -- Tampa Bay Buccaneers
	, (4,  7,  9, 5) -- Cleveland Browns
	, (4, 25, 10, 0) -- Philadelphia Eagles
	, (4, 28, 10, 1) -- Los Angeles Chargers
	, (4, 13, 10, 2) -- Houston Texans
	, (4,  8, 10, 3) -- Dallas Cowboys
	, (4, 23, 10, 4) -- New York Giants
	, (4, 14, 10, 5) -- Indianapolis Colts
	, (4, 17, 11, 0) -- Los Angeles Rams
	, (4, 11, 11, 1) -- Green Bay Packers
	, (4, 18, 11, 2) -- Oakland Raiders
	, (4, 16, 11, 3) -- Kansas City Chiefs
	, (4,  2, 11, 4) -- Baltimore Ravens
	, (4, 32, 11, 5) -- Washington Redskins
-- quarterbacks
	, (5, 2504775,  7, 0) -- Drew Brees
	, (5, 2506109,  7, 1) -- Ben Roethlisberger
	, (5, 2555334,  7, 2) -- Jared Goff
	, (5, 2532975,  8, 0) -- Russell Wilson
	, (5,   79860,  8, 1) -- Matthew Stafford
	, (5, 2495455,  8, 2) -- Cam Newton
	, (5, 2558063,  9, 0) -- Deshaun Watson
	, (5, 2532820,  9, 1) -- Kirk Cousins
	, (5, 2552466,  9, 2) -- Marcus Mariota
	, (5, 2506363, 10, 0) -- Aaron Rodgers
	, (5, 2506121, 10, 1) -- Philip Rivers
	, (5, 2543801, 10, 2) -- Jimmy Garoppolo
	, (5, 2504211, 11, 0) -- Tom Brady
	, (5, 2555259, 11, 1) -- Carson Wentz
	, (5, 2555334, 11, 2) -- Jared Goff
	;

INSERT INTO users (username, password)
	VALUES ('admin', 'invalid_hash_value');
