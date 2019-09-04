CREATE OR REPLACE FUNCTION add_year(sport_type_id INT, year int) RETURNS BOOLEAN
AS $$ BEGIN
INSERT INTO stats (sport_type_id, year)
VALUES (add_year.sport_type_id, add_year.year);
RETURN FOUND;
END $$
LANGUAGE plpgsql;
