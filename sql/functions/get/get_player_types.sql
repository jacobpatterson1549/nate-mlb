CREATE OR REPLACE FUNCTION get_player_types() RETURNS SETOF player_types
AS $$ BEGIN
RETURN QUERY
SELECT id, sport_type_id, name, description
FROM player_types
ORDER BY sport_type_id ASC, id ASC;
END $$
LANGUAGE plpgsql;
