CREATE OR REPLACE FUNCTION clr_stat(sport_type_id INT) RETURNS BOOLEAN
AS $$
WITH updated AS (
UPDATE stats AS s
SET etl_timestamp = NULL, etl_json = NULL
WHERE s.active
AND s.sport_type_id = clr_stat.sport_type_id
RETURNING s.id)
SELECT COUNT(*) > 0 FROM updated
$$
LANGUAGE SQL;
