CREATE OR REPLACE FUNCTION add_year(sport_type_id INT, year int) RETURNS BOOLEAN
AS $$
WITH inserted AS (
INSERT INTO stats (sport_type_id, year)
VALUES (add_year.sport_type_id, add_year.year)
RETURNING id)
SELECT COUNT(*) > 0 FROM inserted
$$
LANGUAGE SQL;
