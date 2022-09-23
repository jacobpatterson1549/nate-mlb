package db

import (
	"fmt"
)

// Year contains a year that has been set for stats and whether it is active
type Year struct {
	Value  int
	Active bool
}

// GetYears gets years for a SportType
func (ds Datastore) GetYears(st SportType) ([]Year, error) {
	return ds.db.GetYears(st)
}

func (d sqlDB) GetYears(st SportType) ([]Year, error) {
	sqlFunction := newReadSQLFunction("get_years", []string{"year", "active"}, st)
	rows, err := d.db.Query(sqlFunction.sql(), sqlFunction.args...)
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
				return years, fmt.Errorf("multiple active years in db (second is %v)", years[i].Value)
			}
			activeYearFound = true
			years[i].Active = true
		}
		i++
	}
	return years, nil
}

// SaveYears saves the specified years and sets the active year for a SportType
func (ds Datastore) SaveYears(st SportType, futureYears []Year) error {
	previousYears, err := ds.GetYears(st)
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

	t, err := ds.db.begin()
	if err != nil {
		return err
	}
	// do this first to ensure one row is affected, in the case that the active row is deleted
	t.ClrYearActive(st)
	for deleteYear := range previousYearsMap {
		t.DelYear(st, deleteYear)
	}
	for _, insertYear := range insertYears {
		t.AddYear(st, insertYear)
	}
	if activeYearPresent {
		t.SetYearActive(st, activeYear)
	}
	return t.execute()
}

func (t *sqlTX) ClrYearActive(st SportType) {
	t.queries = append(t.queries, newWriteSQLFunction("clr_year_active", st))
}

func (t *sqlTX) DelYear(st SportType, deleteYear int) {
	t.queries = append(t.queries, newWriteSQLFunction("del_year", st, deleteYear))
}

func (t *sqlTX) AddYear(st SportType, insertYear int) {
	t.queries = append(t.queries, newWriteSQLFunction("add_year", st, insertYear))
}

func (t *sqlTX) SetYearActive(st SportType, activeYear int) {
	t.queries = append(t.queries, newWriteSQLFunction("set_year_active", st, activeYear))
}
