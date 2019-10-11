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

// GetYears gets years for a SportType
func GetYears(st SportType) ([]Year, error) {
	sqlFunction := newReadSQLFunction("get_years", []string{"year", "active"}, st)
	rows, err := db.Query(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return nil, fmt.Errorf("reading years: %w", err)
	}
	defer rows.Close()

	var years []Year
	activeYearFound := false
	var active bool
	i := 0
	for rows.Next() {
		years = append(years, Year{})
		err = rows.Scan(&years[i].Value, &active)
		if err != nil {
			return years, fmt.Errorf("reading year: %w", err)
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
	return years, nil
}

// SaveYears saves the specified years and sets the active year for a SportType
func SaveYears(st SportType, futureYears []Year) error {
	return saveYears(st, futureYears, GetYears, executeInTransaction)
}

func saveYears(st SportType, futureYears []Year, getYearsFunc func(st SportType) ([]Year, error), executeInTransactionFunc func(queries []writeSQLFunction) error) error {
	previousYears, err := getYearsFunc(st)
	if err != nil {
		return err
	}
	previousYearsMap := make(map[int]bool, len(previousYears))
	for _, year := range previousYears {
		previousYearsMap[year.Value] = true
	}

	insertYears := make([]int, 0, len(futureYears))
	var activeYear int
	activeYearPresent := false
	for _, year := range futureYears {
		if year.Active {
			if activeYearPresent {
				return fmt.Errorf("multiple active years present in %v", futureYears)
			}
			activeYear = year.Value
			activeYearPresent = true
		}
		if _, ok := previousYearsMap[year.Value]; !ok {
			insertYears = append(insertYears, year.Value)
		}
		delete(previousYearsMap, year.Value)
	}

	queries := make([]writeSQLFunction, 0, len(insertYears)+len(previousYearsMap)+2)
	// do this first to ensure one row is affected, in the case that the active row is deleted
	queries = append(queries, newWriteSQLFunction("clr_year_active", st))
	for deleteYear := range previousYearsMap {
		queries = append(queries, newWriteSQLFunction("del_year", st, deleteYear))
	}
	for _, insertYear := range insertYears {
		queries = append(queries, newWriteSQLFunction("add_year", st, insertYear))
	}
	if activeYearPresent {
		queries = append(queries, newWriteSQLFunction("set_year_active", st, activeYear))
	}
	return executeInTransactionFunc(queries)
}
