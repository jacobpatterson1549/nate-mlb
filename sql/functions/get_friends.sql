CREATE OR REPLACE FUNCTION get_friends(sport_type_id INT) RETURNS SETOF friends
AS $$ BEGIN
RETURN QUERY
SELECT f.id, f.name, f.display_order, f.sport_type_id, f.year
FROM stats AS s
JOIN friends AS f ON s.year = f.year AND s.sport_type_id = f.sport_type_id
WHERE s.sport_type_id = get_friends.sport_type_id
AND s.active
ORDER BY f.display_order ASC;
END $$
LANGUAGE plpgsql;
