-- $ sudo -u postgres psql
-- CREATE DATABASE nate_mlb_db;
-- CREATE user nate WITH ENCRYPTED PASSWORD 'Have19_each+Iowa';
-- GRANT ALL PRIVILEGES ON DATABASE nate_mlb_db to nate;
-- \q
-- $ PGPASSWORD=Have19_each%Iowa psql nate -h 127.0.0.1 -d nate_mlb_db

CREATE TABLE users
	( username VARCHAR(20) PRIMARY KEY
	, password TEXT
	);
INSERT INTO users (username, password)
	VALUES ('admin', 'invalid_hash_value');

--DROP TABLE sport_types, stats, friends, player_types, players;

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
	, CONSTRAINT name_sport_type_id_year_unique UNIQUE (name, sport_type_id, year)
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
	, CONSTRAINT player_type_id_player_id_friend_id_unique UNIQUE (player_type_id, player_id, friend_id)
	, FOREIGN KEY (player_type_id) REFERENCES player_types (id) ON DELETE RESTRICT
	, FOREIGN KEY (friend_id) REFERENCES friends (id) ON DELETE CASCADE
	);

INSERT INTO sport_types (id, name)
	VALUES (
	   1, 'mlb')
	, (2, 'nfl')
	;
SELECT setVal('sport_types_id_seq', COALESCE((SELECT MAX(id)+1 FROM sport_types), 1));

INSERT INTO stats (sport_type_id, year, active)
	VALUES (
	   1, 2019, TRUE)
	, (2, 2018, NULL)
	, (2, 2019, TRUE)
	;

INSERT INTO friends (id, name, display_order, sport_type_id, year)
	VALUES (
	   1, 'Bob',   0, 1, 2019)
	, (2, 'W',     1, 1, 2019)
	, (3, 'Nate',  2, 1, 2019)
	, (4, 'Sam',   3, 1, 2019)
	, (5, 'Steve', 4, 1, 2019)
	, (6, 'Mike',  5, 1, 2019)
	, ( 7, 'Nate',  0, 2, 2019)
	, ( 8, 'Sam',   1, 2, 2019)
	, ( 9, 'Viet',  2, 2, 2019)
	, (10, 'W',     3, 2, 2019)
	, (11, 'Bob',   4, 2, 2019)
	, (12, 'Steve', 5, 2, 2019)
	;
SELECT setVal('friends_id_seq', COALESCE((SELECT MAX(id)+1 FROM friends), 1));

INSERT INTO player_types (id, sport_type_id, name, description)
	VALUES (
	   1, 1, 'Teams', 'Wins')
	, (2, 1, 'Hitting', 'Home Runs')
	, (3, 1, 'Pitching', 'Wins')
	, (4, 2, 'Teams', 'Wins')
	, (5, 2, 'Quarterbacks', 'Touchdown (passes+runs)')
	, (6, 2, 'Runningbacks & Recievers', 'Touchdowns')
	;

INSERT INTO players (player_type_id, player_id, friend_id, display_order)
	VALUES (
-- teams:  name_display_full
	   1, 111, 1, 0) -- Boston Red Sox
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
-- hitters: name_display_first_last
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
-- pitchers: name_display_first_last
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
-- nfl teams:
	, (4, 17,  7, 0) -- Los Angeles Rams
	, (4, 13,  7, 1) -- Houston Texans
	, (4,  1,  7, 2) -- Atlanta Falcons
	, (4, 12,  7, 3) -- Tennessee Titans
	, (4, 31,  7, 4) -- Tampa Bay Buccaneers
	, (4, 28,  8, 0) -- Los Angeles Chargers
	, (4, 14,  8, 1) -- Indianapolis Colts
	, (4, 11,  8, 2) -- Green Bay Packers
	, (4,  4,  8, 3) -- Carolina Panthers
	, (4,  3,  8, 4) -- Buffalo Bills
	, (4, 22,  9, 0) -- New Orleans Saints
	, (4, 25,  9, 1) -- Philadelphia Eagles
	, (4,  2,  9, 2) -- Baltimore Ravens
	, (4, 18,  9, 3) -- Oakland Raiders
	, (4, 10,  9, 4) -- Detroit Lions
	, (4, 16, 10, 0) -- Kansas City Chiefs
	, (4, 15, 10, 1) -- Jacksonville Jaguars
	, (4, 20, 10, 2) -- Minnesota Vikings
	, (4,  9, 10, 3) -- Denver Broncos
	, (4, 26, 10, 4) -- Arizona Cardinals
	, (4,  7, 11, 0) -- Cleveland Browns
	, (4,  8, 11, 1) -- Dallas Cowboys
	, (4, 30, 11, 2) -- Seattle Seahawks
	, (4, 29, 11, 3) -- San Francisco 49ers
	, (4, 23, 11, 4) -- New York Giants
	, (4, 21, 12, 0) -- New England Patriots
	, (4,  5, 12, 1) -- Chicago Bears
	, (4, 27, 12, 2) -- Pittsburgh Steelers
	, (4, 24, 12, 3) -- New York Jets
	, (4, 32, 12, 4) -- Washington Redskins
-- quarterbacks
	, (5, 2558125, 07, 0) -- Patrick Mahomes
	, (5, 2506109, 07, 1) -- Ben Roethlisberger
	, (5, 79860,   07, 2) -- Matthew Stafford
	, (5, 2504775, 08, 0) -- Drew Brees
	, (5, 2533031, 08, 1) -- Andrew Luck
	, (5, 2555260, 08, 2) -- Dak Prescott
	, (5, 2558063, 09, 0) -- Deshaun Watson
	, (5, 2555259, 09, 1) -- Carson Wentz
	, (5, 2562382, 09, 2) -- Kyler Murray
	, (5, 2560800, 10, 0) -- Baker Mayfield
	, (5, 2560757, 10, 1) -- Lamar Jackson
	, (5, 2532820, 10, 2) -- Kirk Cousins
	, (5, 2504211, 11, 0) -- Tom Brady
	, (5, 2555334, 11, 1) -- Jared Goff
	, (5, 310,     11, 2) -- Matt Ryan
	, (5, 2506363, 12, 0) -- Aaron Rodgers
	, (5, 2532975, 12, 1) -- Russell Wilson
	, (5, 2506121, 12, 2) -- Philip Rivers
-- running backs and wide receivers
	, (6, 2552475, 07, 0) -- Todd Gurley
	, (6, 2557997, 07, 1) -- Christian McCaffrey
	, (6, 2561021, 07, 2) -- Nick Chubb
	, (6, 2540165, 08, 0) -- DeAndre Hopkins
	, (6, 2543496, 08, 1) -- Odell Beckham
	, (6, 2495454, 08, 2) -- Julio Jones
	, (6, 2540175, 09, 0) -- Le'Veon Bell
	, (6, 2553435, 09, 1) -- David Johnson
	, (6, 2557976, 09, 2) -- Joe Mixon
	, (6, 2560968, 10, 0) -- Saquon Barkley
	, (6, 2540258, 10, 1) -- Travis Kelce
	, (6, 2556214, 10, 2) -- Tyreek Hill
	, (6, 2555224, 11, 0) -- Ezekiel Elliott
	, (6, 2557978, 11, 1) -- James Conner
	, (6, 2543495, 11, 2) -- Davante Adams
	, (6, 2558019, 12, 0) -- Alvin Kamara
	, (6, 2561358, 12, 1) -- J.J. Jones
	, (6, 2552487, 12, 2) -- Amari Cooper
	;
