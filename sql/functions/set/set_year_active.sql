CREATE OR REPLACE FUNCTION set_year_active(sport_type_id INT, year INT) RETURNS BOOLEAN
AS $$
WITH updated AS (
UPDATE stats AS s
SET active = TRUE
WHERE NOT COALESCE(s.active, FALSE)
AND s.sport_type_id = set_year_active.sport_type_id
AND s.year = set_year_active.year
RETURNING s.id)
SELECT COUNT(*) > 0 FROM updated
$$
LANGUAGE SQL;
