CREATE OR REPLACE FUNCTION del_player(id INT) RETURNS BOOLEAN
AS $$ BEGIN
DELETE FROM players AS p
WHERE p.id = del_player.id;
RETURN FOUND;
END $$
LANGUAGE plpgsql;