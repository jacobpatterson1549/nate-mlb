CREATE OR REPLACE FUNCTION set_friend(display_order INT, name VARCHAR, id INT) RETURNS BOOLEAN
AS $$
WITH updated AS (
UPDATE friends AS f
SET display_order = set_friend.display_order, name = set_friend.name
WHERE f.id = set_friend.id
RETURNING f.id)
SELECT COUNT(*) > 0 FROM updated
$$
LANGUAGE SQL;
