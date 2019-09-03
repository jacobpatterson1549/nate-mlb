CREATE OR REPLACE FUNCTION set_user_password(username VARCHAR, password TEXT) RETURNS BIGINT
    AS $$
    WITH update_cte (update_count) AS (
        UPDATE users
        SET password = set_user_password.password
        WHERE username = set_user_password.username
        RETURNING 1
    )
    SELECT COUNT(update_count) FROM update_cte
    $$
    LANGUAGE SQL;
