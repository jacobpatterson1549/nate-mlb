package db

import (
	"errors"
	"fmt"
)

// Year contains a year that has been set for stats and whether it is active
type Year struct {
	Value  int
	Active bool
}

// GetYears gets the specified years
func GetYears(st SportType) ([]Year, error) {
	var years []Year

	rows, err := db.Query("SELECT year, active FROM get_years($1)", st)
	if err != nil {
		return years, fmt.Errorf("problem reading years: %v", err)
	}
	defer rows.Close()

	activeYearFound := false
	var active bool
	i := 0
	for rows.Next() {
		years = append(years, Year{})
		err = rows.Scan(&years[i].Value, &active)
		if err != nil {
			return years, fmt.Errorf("problem reading year: %v", err)
		}
		if active {
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

	queries := make(chan sqlFunction, len(insertYears)+len(previousYearsMap)+2)
	quit := make(chan error)
	go exececuteInTransaction(queries, quit)
	// do this first to ensure one row is affected, in the case that the active row is deleted
	queries <- newSQLFunction("clr_year_active", st)
	for deleteYear := range previousYearsMap {
		queries <- newSQLFunction("del_year", st, deleteYear)
	}
	for _, insertYear := range insertYears {
		queries <- newSQLFunction("add_year", st, insertYear)
	}
	queries <- newSQLFunction("set_year_active", st, activeYear)
	close(queries)
	return <-quit
}
