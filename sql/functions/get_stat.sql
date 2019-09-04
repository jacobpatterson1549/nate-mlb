CREATE OR REPLACE FUNCTION get_stat(sport_type_id INT) RETURNS SETOF stats
AS $$ BEGIN
RETURN QUERY
SELECT s.id, s.sport_type_id, s.year, s.active, s.etl_timestamp, s.etl_json
FROM stats AS s
WHERE s.sport_type_id = get_stat.sport_type_id
AND s.active;
END $$
LANGUAGE plpgsql;
