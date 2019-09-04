CREATE OR REPLACE FUNCTION get_years(sport_type_id INT) RETURNS SETOF years
AS $$ BEGIN
RETURN QUERY
SELECT y.sport_type_id, y.year, y.active
FROM years AS y
WHERE y.sport_type_id = get_years.sport_type_id
ORDER BY y.year ASC;
END $$
LANGUAGE plpgsql;
