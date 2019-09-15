CREATE OR REPLACE FUNCTION del_year(sport_type_id INT, year int) RETURNS BOOLEAN
AS $$
WITH deleted AS (
DELETE FROM stats AS s
WHERE s.sport_type_id = del_year.sport_type_id
AND s.year = del_year.year
RETURNING s.id)
SELECT COUNT(*) > 0 FROM deleted
$$
LANGUAGE SQL;
