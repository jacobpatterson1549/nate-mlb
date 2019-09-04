CREATE OR REPLACE VIEW years (sport_type_id, year, active) AS
    SELECT s.sport_type_id
        , s.year
        , COALESCE(s.active, FALSE)
    FROM stats AS s
    ORDER BY s.sport_type_id, s.year ASC;
