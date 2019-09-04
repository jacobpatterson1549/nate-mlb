CREATE OR REPLACE FUNCTION get_user_password(username VARCHAR, OUT password TEXT) RETURNS SETOF TEXT
AS $$ BEGIN
RETURN QUERY
SELECT u.password FROM users AS u
WHERE u.username = get_user_password.username;
END $$
LANGUAGE plpgsql;
