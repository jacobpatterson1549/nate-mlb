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
func GetActiveYear() (int, error) {
	var activeYear int

	row := db.QueryRow("SELECT year FROM stats WHERE active")
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
func GetYears() ([]Year, error) {
	years := []Year{}

	rows, err := db.Query("SELECT year, active FROM stats ORDER BY year ASC")
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
			return years, fmt.Errorf("problem reading data: %v", err)
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

// SetYears saves the specified years and sets the active year // TODO: rename to SaveYears
func SetYears(activeYear int, years []int) error { // TODO: swap param order
	currentYears, err := GetYears()
	if err != nil {
		return err
	}
	currentYearsMap := make(map[int]bool)
	for _, year := range currentYears {
		currentYearsMap[year.Value] = true
	}

	insertYears := []int{}
	activeYearPresent := false
	for _, year := range years {
		if year == activeYear {
			activeYearPresent = true
		}
		if _, ok := currentYearsMap[year]; !ok {
			insertYears = append(insertYears, year)
		}
		delete(currentYearsMap, year)
	}
	if len(years) > 0 && !activeYearPresent {
		return fmt.Errorf("active year %v not present in years: %v", activeYear, years)
	}
	deleteYears := currentYearsMap

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("problem starting transaction: %v", err)
	}
	var result sql.Result
	for year := range deleteYears {
		if err == nil {
			result, err = tx.Exec(
				"DELETE FROM stats WHERE year = $1",
				year)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for _, year := range insertYears {
		if err == nil {
			result, err = tx.Exec(
				"INSERT INTO stats (year) VALUES ($1)",
				year)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	// remove active year
	if err == nil && len(years) > 0 {
		result, err = tx.Exec("UPDATE stats SET active = NULL WHERE active")
		// TODO: no need to expecet 1 row because no years may be present, but still need to rollback on error
	}
	// set active year
	if err == nil && len(years) > 0 {
		// TODO: make "func affectOneRow(tx *sql.Tx, sql string) error" function to make rollback
		result, err = tx.Exec(
			"UPDATE stats SET active = TRUE WHERE year = $1",
			activeYear)
		if err != nil {
			err = expectSingleRowAffected(result)
		}
	}
	if err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			err = fmt.Errorf("problem: %v, ROLLBACK ERROR: %v", err.Error(), err2.Error())
		}
	} else {
		err = tx.Commit()
	}
	if err != nil {
		return fmt.Errorf("problem saving years: %v", err)
	}
	return nil
}
