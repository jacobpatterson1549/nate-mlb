CREATE OR REPLACE FUNCTION get_sport_types() RETURNS SETOF sport_types
AS $$ BEGIN
RETURN QUERY
SELECT id, name, url
FROM sport_types
ORDER BY id ASC;
END $$
LANGUAGE plpgsql;
