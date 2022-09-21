CREATE OR REPLACE FUNCTION del_player(id INT, sport_type_id INT) RETURNS BOOLEAN
AS $$
WITH deleted AS (
DELETE FROM players AS p
WHERE p.id = del_player.id
RETURNING p.id)
SELECT COUNT(*) > 0 FROM deleted
$$
LANGUAGE SQL;
