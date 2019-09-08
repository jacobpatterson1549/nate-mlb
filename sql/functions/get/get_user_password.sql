CREATE OR REPLACE FUNCTION get_user_password(username VARCHAR, OUT password TEXT) RETURNS SETOF TEXT
AS $$
SELECT u.password FROM users AS u
WHERE u.username = get_user_password.username;
$$
LANGUAGE SQL;
