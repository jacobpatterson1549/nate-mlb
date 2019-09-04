CREATE OR REPLACE FUNCTION get_friends(sport_type_id INT, OUT id INT, OUT name VARCHAR, OUT display_order INT) RETURNS SETOF RECORD
AS $$ BEGIN
RETURN QUERY
SELECT f.id, f.name, f.display_order
FROM stats AS s
JOIN friends AS f ON s.year = f.year AND s.sport_type_id = f.sport_type_id
WHERE s.active
AND s.sport_type_id = get_friends.sport_type_id
ORDER BY f.display_order ASC;
END $$
LANGUAGE plpgsql;
