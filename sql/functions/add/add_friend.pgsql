CREATE OR REPLACE FUNCTION add_friend(display_order INT, name VARCHAR, sport_type_id INT) RETURNS BOOLEAN
AS $$
WITH inserted AS (
INSERT INTO friends (display_order, name, stat_id)
SELECT add_friend.display_order, add_friend.name, s.id
FROM stats AS s
WHERE s.active
AND s.sport_type_id = add_friend.sport_type_id
RETURNING id)
SELECT COUNT(*) > 0 FROM inserted
$$
LANGUAGE SQL;
