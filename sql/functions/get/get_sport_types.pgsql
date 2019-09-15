CREATE OR REPLACE FUNCTION get_sport_types() RETURNS SETOF sport_types
AS $$
SELECT id, name, url
FROM sport_types
ORDER BY id ASC;
$$
LANGUAGE SQL;
