package db

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"
)

func (ds Datastore) getSetupTableQueries() ([]string, error) {
	var queries []string
	// order of setup files matters - some queries reference others
	setupFileNames := []string{"users", "sport_types", "stats", "friends", "player_types", "players"}
	for _, setupFileName := range setupFileNames {
		b, err := ds.fs.ReadFile(fmt.Sprintf("sql/setup/%s.pgsql", setupFileName))
		if err != nil {
			return nil, err
		}
		setupQueries := strings.Split(string(b), ";")
		queries = concat(queries, setupQueries)
	}
	return queries, nil
}

func (ds Datastore) getSetupFunctionQueries() ([]string, error) {
	var queries []string
	walkDirFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := ds.fs.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading sql function: %w", err)
		}
		queries = append(queries, string(b))
		return nil
	}
	if err := fs.WalkDir(ds.fs, "sql/functions", walkDirFunc); err != nil {
		return nil, fmt.Errorf("reading sql functions: %w", err)
	}
	return queries, nil
}

// SetupTablesAndFunctions runs setup scripts to ensure tables are initialized, populated, and re-adds all functions to access/change saved data
func (ds Datastore) SetupTablesAndFunctions() error {
	setupTableQueries, err := ds.getSetupTableQueries()
	if err != nil {
		return err
	}
	setupFunctionQueries, err := ds.getSetupFunctionQueries()
	if err != nil {
		return err
	}
	queries := concat(setupTableQueries, setupFunctionQueries)
	tx, err := ds.db.Begin()
	if err != nil {
		return fmt.Errorf("starting database setup: %w", err)
	}
	for _, sql := range queries {
		_, err := tx.Exec(sql)
		if err != nil {
			err = fmt.Errorf("setting: %w\nquery: %v", err, strings.TrimSpace(sql))
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("%v, ROLLBACK ERROR: %w", err, rollbackErr)
			}
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing database setup: %w", err)
	}
	return nil
}

// LimitPlayerTypes reduces the player types to those in the specified csv.
// Also limits the sport types to those for the specified player types.
// Note that this function mutates the supplied maps.
func (ds *Datastore) LimitPlayerTypes(playerTypesCsv string) error {
	if len(playerTypesCsv) == 0 {
		return nil
	}
	playerTypeStrings := strings.Split(playerTypesCsv, ",")
	// determine which PlayerTypes and SportTypes to use
	selectedPlayerTypesMap := make(map[PlayerType]bool)
	selectedSportTypesMap := make(map[SportType]bool)
	for _, pts := range playerTypeStrings {
		pti, err := strconv.Atoi(pts)
		if err != nil {
			return fmt.Errorf("invalid PlayerType: %w", err)
		}
		pt := PlayerType(pti)
		ptInfo, ok := ds.playerTypes[pt]
		if !ok {
			return fmt.Errorf("unknown PlayerType: %v", pt)
		}
		selectedPlayerTypesMap[pt] = true
		selectedSportTypesMap[ptInfo.SportType] = true
	}
	// limit PlayerTypes and SportTypes
	for pt := range ds.playerTypes {
		if _, ok := selectedPlayerTypesMap[pt]; !ok {
			delete(ds.playerTypes, pt)
		}
	}
	for st := range ds.sportTypes {
		if _, ok := selectedSportTypesMap[st]; !ok {
			delete(ds.sportTypes, st)
			continue
		}
	}
	return nil
}

// concat joins the slices, returning a new slice.
func concat(a, b []string) []string {
	c := make([]string, len(a)+len(b))
	copy(c[copy(c, a):], b)
	return c
}
