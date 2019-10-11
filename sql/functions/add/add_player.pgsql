CREATE OR REPLACE FUNCTION add_player(display_order INT, player_type_id INT, source_id INT, friend_id INT) RETURNS BOOLEAN
AS $$
WITH inserted AS (
INSERT INTO players (display_order, player_type_id, source_id, friend_id)
SELECT add_player.display_order, add_player.player_type_id, add_player.source_id, add_player.friend_id
FROM stats AS s
JOIN player_types AS pt ON add_player.player_type_id = pt.id
WHERE s.active
AND s.sport_type_id = pt.sport_type_id
RETURNING id)
SELECT COUNT(*) > 0 FROM inserted
$$
LANGUAGE SQL;
