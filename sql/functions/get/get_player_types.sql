CREATE OR REPLACE FUNCTION get_player_types() RETURNS SETOF player_types
AS $$
SELECT id, sport_type_id, name, description, score_type
FROM player_types
ORDER BY sport_type_id ASC, id ASC;
$$
LANGUAGE SQL;
