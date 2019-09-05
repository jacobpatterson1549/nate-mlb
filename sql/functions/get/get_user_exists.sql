CREATE OR REPLACE FUNCTION get_user_exists(username VARCHAR, OUT username_exists BOOLEAN) RETURNS BOOLEAN
AS $$ BEGIN
    SELECT EXISTS (
        SELECT u.username
        FROM users AS u
        WHERE u.username = get_user_exists.username
    )
    INTO get_user_exists.username_exists;
END $$
LANGUAGE plpgsql;
