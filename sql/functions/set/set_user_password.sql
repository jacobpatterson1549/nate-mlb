CREATE OR REPLACE FUNCTION set_user_password(username VARCHAR, password TEXT) RETURNS BOOLEAN
AS $$
WITH updated AS (
UPDATE users AS u
SET password = set_user_password.password
WHERE u.username = set_user_password.username
RETURNING u.username)
SELECT COUNT(*) > 0 FROM updated
$$
LANGUAGE SQL;
