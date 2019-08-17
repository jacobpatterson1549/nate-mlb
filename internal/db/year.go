package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// Year contains a year that has been set for stats and whether it is active
type Year struct {
	Value  int
	Active bool
}

// GetActiveYear gets the active year for stat retrieval
func GetActiveYear(st SportType) (int, error) {
	var activeYear int

	row := db.QueryRow("SELECT year FROM stats WHERE active AND sport_type_id = $1", st)
	err := row.Scan(&activeYear)
	if err == sql.ErrNoRows {
		return activeYear, errors.New("no active year")
	}
	if err != nil {
		return activeYear, fmt.Errorf("problem getting active year: %v", err)
	}
	return activeYear, nil
}

// GetYears gets the specified years
func GetYears(st SportType) ([]Year, error) {
	var years []Year

	rows, err := db.Query("SELECT year, active FROM stats WHERE sport_id = $1 ORDER BY year ASC", st)
	if err != nil {
		return years, fmt.Errorf("problem reading years: %v", err)
	}
	defer rows.Close()

	activeYearFound := false
	var active sql.NullBool
	i := 0
	for rows.Next() {
		years = append(years, Year{})
		err = rows.Scan(&years[i].Value, &active)
		if err != nil {
			return years, fmt.Errorf("problem reading year: %v", err)
		}
		if active.Valid && active.Bool {
			if activeYearFound {
				return years, errors.New("multiple active years in db")
			}
			activeYearFound = true
			years[i].Active = true
		}
		i++
	}
	if !activeYearFound && len(years) > 0 {
		return years, errors.New("no active year in db")
	}
	return years, nil
}

// SaveYears saves the specified years and sets the active year
func SaveYears(st SportType, futureYears []Year) error {
	previousYears, err := GetYears(st)
	if err != nil {
		return err
	}
	previousYearsMap := make(map[int]bool)
	for _, year := range previousYears {
		previousYearsMap[year.Value] = true
	}

	var insertYears []int
	var activeYear int
	activeYearPresent := false
	for _, year := range futureYears {
		if year.Active {
			if activeYearPresent {
				return fmt.Errorf("multiple active years present in %v", err)
			}
			activeYear = year.Value
			activeYearPresent = true
		}
		if _, ok := previousYearsMap[year.Value]; !ok {
			insertYears = append(insertYears, year.Value)
		}
		delete(previousYearsMap, year.Value)
	}
	if len(futureYears) > 0 && !activeYearPresent {
		return fmt.Errorf("active year not present in years: %v", futureYears)
	}

	queries := make([]query, len(insertYears)+len(previousYearsMap)+2)
	// remove active year
	// do this first to ensure one row is affected, in the case that the active row is deleted
	queries[0] = query{
		sql:  "UPDATE stats SET active = NULL WHERE sport_type_id = $1 AND active",
		args: make([]interface{}, 1),
	}
	queries[0].args[0] = st
	i := 1
	for _, insertYear := range insertYears {
		queries[i] = query{
			sql:  "INSERT INTO stats (sport_type_id, year) VALUES ($1, $2)",
			args: make([]interface{}, 2),
		}
		queries[i].args[0] = st
		queries[i].args[1] = insertYear
		i++
	}
	for deleteYear := range previousYearsMap {
		queries[i] = query{
			sql:  "DELETE FROM stats WHERE sport_type_id = $1 AND year = $1",
			args: make([]interface{}, 2),
		}
		queries[i].args[0] = st
		queries[i].args[1] = deleteYear
		i++
	}
	// set active year
	queries[i] = query{
		sql:  "UPDATE stats SET active = TRUE WHERE sport_type_id = $1 AND year = $2",
		args: make([]interface{}, 2),
	}
	queries[i].args[0] = st
	queries[i].args[1] = activeYear
	return exececuteInTransaction(&queries)
}
