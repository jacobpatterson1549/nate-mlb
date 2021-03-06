CREATE OR REPLACE FUNCTION get_players(sport_type_id INT) RETURNS SETOF players
AS $$
SELECT p.id, p.player_type_id, p.source_id, p.friend_id, p.display_order
FROM stats AS s
JOIN friends AS f ON s.id = f.stat_id
JOIN players AS p ON f.id = p.friend_id
WHERE s.active
AND s.sport_type_id = get_players.sport_type_id
ORDER BY p.player_type_id ASC, p.friend_id ASC, p.display_order ASC;
$$
LANGUAGE SQL;
