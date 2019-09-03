CREATE OR REPLACE FUNCTION get_user_password(username VARCHAR) RETURNS SETOF TEXT
    AS $$
    SELECT password FROM users
    WHERE username = get_user_password.username
    $$
    LANGUAGE SQL;