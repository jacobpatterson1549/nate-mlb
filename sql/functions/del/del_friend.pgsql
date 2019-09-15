CREATE OR REPLACE FUNCTION del_friend(id INT) RETURNS BOOLEAN
AS $$
WITH deleted AS (
DELETE FROM friends AS f
WHERE f.id = del_friend.id
RETURNING f.id)
SELECT COUNT(*) > 0 FROM deleted
$$
LANGUAGE SQL;
