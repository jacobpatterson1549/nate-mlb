CREATE OR REPLACE FUNCTION add_user(username VARCHAR, password TEXT) RETURNS BOOLEAN
AS $$
WITH inserted AS (
INSERT INTO users (username, password)
SELECT add_user.username, add_user.password
RETURNING username)
SELECT COUNT(*) > 0 FROM inserted
$$
LANGUAGE SQL;
