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

	row := db.QueryRow(
		`SELECT year FROM stats
		WHERE sport_type_id = $1
		AND active`,
		st)
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

	rows, err := db.Query(
		`SELECT year, active
		FROM stats
		WHERE sport_type_id = $1
		ORDER BY year ASC`,
		st)
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
		return years, errors.New("no active year when retrieving year list")
	}
	return years, nil
}

// SaveYears saves the specified years and sets the active year
func SaveYears(st SportType, futureYears []Year) error {
	previousYears, err := GetYears(st)
	if err != nil {
		return err
	}
	previousYearsMap := make(map[int]bool, len(previousYears))
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

	queries := make(chan query, len(insertYears)+len(previousYearsMap)+2)
	quit := make(chan error)
	go exececuteInTransaction(queries, quit)
	// remove active year
	// do this first to ensure one row is affected, in the case that the active row is deleted
	queries <- newQuery(
		`UPDATE stats
		SET active = NULL
		WHERE sport_type_id = $1
		AND active`,
		st,
	)
	for deleteYear := range previousYearsMap {
		queries <- newQuery(
			`DELETE FROM stats
			WHERE sport_type_id = $1
			AND year = $2`,
			st,
			deleteYear,
		)
	}
	for _, insertYear := range insertYears {
		queries <- newQuery(
			`INSERT INTO stats
			(sport_type_id, year)
			VALUES ($1, $2)`,
			st,
			insertYear,
		)
	}

	// set active year
	queries <- query{
		sql: `UPDATE stats
			SET active = TRUE
			WHERE sport_type_id = $1
			AND year = $2`,
		args: []interface{}{
			st,
			activeYear,
		},
	}
	close(queries)
	return <-quit
}
