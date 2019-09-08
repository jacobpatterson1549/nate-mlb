CREATE OR REPLACE FUNCTION get_years(sport_type_id INT, OUT year INT, OUT active BOOLEAN) RETURNS SETOF RECORD
AS $$
SELECT s.year, COALESCE(s.active, FALSE)
FROM stats AS s
WHERE s.sport_type_id = get_years.sport_type_id
ORDER BY s.year ASC;
$$
LANGUAGE SQL;
