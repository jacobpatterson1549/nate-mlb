CREATE OR REPLACE FUNCTION clr_year_active(sport_type_id INT) RETURNS BOOLEAN
AS $$ BEGIN
UPDATE stats AS s
SET active = NULL
WHERE s.active
AND s.sport_type_id = clr_year_active.sport_type_id;
RETURN FOUND;
END $$
LANGUAGE plpgsql;