CREATE OR REPLACE FUNCTION set_stat(etl_timestamp TIMESTAMP, etl_json JSONB, sport_type_id INT, year int) RETURNS BOOLEAN
AS $$ BEGIN
UPDATE stats AS s
SET etl_timestamp = set_stat.etl_timestamp, etl_json = set_stat.etl_json
WHERE s.sport_type_id = set_stat.sport_type_id
AND s.active
AND s.year = set_stat.year;
RETURN FOUND;
END $$
LANGUAGE plpgsql;
