CREATE OR REPLACE FUNCTION set_player(display_order INT, id INT) RETURNS BOOLEAN
AS $$ BEGIN
UPDATE players AS p
SET display_order = set_player.display_order
WHERE p.id = set_player.id;
RETURN FOUND;
END $$
LANGUAGE plpgsql;
