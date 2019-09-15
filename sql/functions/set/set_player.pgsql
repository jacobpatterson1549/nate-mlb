CREATE OR REPLACE FUNCTION set_player(display_order INT, id INT) RETURNS BOOLEAN
AS $$
WITH updated AS (
UPDATE players AS p
SET display_order = set_player.display_order
WHERE p.id = set_player.id
RETURNING p.id)
SELECT COUNT(*) > 0 FROM updated
$$
LANGUAGE SQL;
