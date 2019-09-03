CREATE OR REPLACE FUNCTION get_user_password(username VARCHAR) RETURNS SETOF TEXT
AS $$ BEGIN
RETURN QUERY
SELECT password FROM users AS u
WHERE u.username = get_user_password.username;
END $$
LANGUAGE plpgsql;
