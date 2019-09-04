CREATE OR REPLACE FUNCTION get_stat(sport_type_id INT) RETURNS SETOF stats
AS $$ BEGIN
RETURN QUERY
SELECT s.id, s.sport_type_id, s.year, s.active, s.etl_timestamp, s.etl_json
FROM stats AS s
WHERE s.active
AND s.sport_type_id = 1;
END $$
LANGUAGE plpgsql;
