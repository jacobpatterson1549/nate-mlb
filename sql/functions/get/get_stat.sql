CREATE OR REPLACE FUNCTION get_stat(sport_type_id INT, OUT year INT, OUT etl_timestamp TIMESTAMP, OUT etl_json JSONB) RETURNS SETOF RECORD
AS $$ BEGIN
RETURN QUERY
SELECT s.year, s.etl_timestamp, s.etl_json
FROM stats AS s
WHERE s.active
AND s.sport_type_id = get_stat.sport_type_id;
END $$
LANGUAGE plpgsql;
