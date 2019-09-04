CREATE OR REPLACE FUNCTION add_user(username VARCHAR, password TEXT) RETURNS BOOL
AS $$ BEGIN
INSERT INTO users (username, password)
SELECT add_user.username, add_user.password
WHERE NOT EXISTS (SELECT 1 FROM users AS u WHERE u.username = add_user.username);
RETURN FOUND;
END $$
LANGUAGE plpgsql;
