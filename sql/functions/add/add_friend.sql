CREATE OR REPLACE FUNCTION add_friend(display_order INT, name VARCHAR, sport_type_id INT) RETURNS BOOLEAN
AS $$ BEGIN
INSERT INTO friends (display_order, name, sport_type_id, year)
SELECT add_friend.display_order, add_friend.name, add_friend.sport_type_id, year
FROM stats AS s
WHERE s.active
AND s.sport_type_id = add_friend.sport_type_id;
RETURN FOUND;
END $$
LANGUAGE plpgsql;
