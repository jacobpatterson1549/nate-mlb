CREATE OR REPLACE FUNCTION del_year(sport_type_id INT, year int) RETURNS BOOLEAN
AS $$ BEGIN
DELETE FROM stats AS s
WHERE s.sport_type_id = del_year.sport_type_id
AND s.year = del_year.year;
RETURN FOUND;
END $$
LANGUAGE plpgsql;
