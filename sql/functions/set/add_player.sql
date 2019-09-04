CREATE OR REPLACE FUNCTION add_player(display_order INT, player_type_id INT, source_id INT, friend_id INT, sport_type_id INT) RETURNS BOOLEAN
AS $$ BEGIN
INSERT INTO players (display_order, player_type_id, source_id, friend_id)
SELECT add_player.display_order, add_player.player_type_id, add_player.source_id, add_player.friend_id
FROM stats AS s
WHERE s.active
AND s.sport_type_id = add_player.sport_type_id;
RETURN FOUND;
END $$
LANGUAGE plpgsql;
