CREATE OR REPLACE FUNCTION set_user_password(username VARCHAR, password TEXT) RETURNS BOOLEAN
AS $$ BEGIN
UPDATE users AS u
SET password = set_user_password.password
WHERE u.username = set_user_password.username;
RETURN FOUND;
END $$
LANGUAGE plpgsql;
