CREATE OR REPLACE FUNCTION get_user_password(username VARCHAR) RETURNS SETOF TEXT
AS $$ BEGIN
SELECT password FROM users
WHERE username = get_user_password.username;
END $$
LANGUAGE plpgsql;
