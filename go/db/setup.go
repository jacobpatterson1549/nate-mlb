package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func getSetupTableQueries(readFileFunc func(filename string) ([]byte, error)) ([]string, error) {
	var queries []string
	// order of setup files matters - some queries reference others
	setupFileNames := []string{"users", "sport_types", "stats", "friends", "player_types", "players"}
	for _, setupFileName := range setupFileNames {
		b, err := readFileFunc(fmt.Sprintf("sql/setup/%s.pgsql", setupFileName))
		if err != nil {
			return nil, err
		}
		setupQueries := strings.Split(string(b), ";")
		queries = append(queries, setupQueries...)
	}
	return queries, nil
}

func getSetupFunctionQueries(readFileFunc func(filename string) ([]byte, error), readDirFunc func(dirname string) ([]os.FileInfo, error)) ([]string, error) {
	var queries []string
	functionDirTypes := []string{"add", "clr", "del", "get", "set"}
	for _, functionDirType := range functionDirTypes {
		functionDir := fmt.Sprintf("sql/functions/%s", functionDirType)
		functionFileInfos, err := readDirFunc(functionDir)
		if err != nil {
			return nil, err
		}
		for _, functionFileInfo := range functionFileInfos {
			b, err := readFileFunc(fmt.Sprintf("%s/%s", functionDir, functionFileInfo.Name()))
			if err != nil {
				return nil, err
			}
			queries = append(queries, string(b))
		}
	}
	return queries, nil
}

// SetupTablesAndFunctions runs setup scripts to ensure tables are initialized, populated, and re-adds all functions to access/change saved data
func SetupTablesAndFunctions() error {
	return setupTablesAndFunctions(ioutil.ReadFile, ioutil.ReadDir)
}

func setupTablesAndFunctions(readFileFunc func(filename string) ([]byte, error), readDirFunc func(dirname string) ([]os.FileInfo, error)) error {
	setupTableQueries, err := getSetupTableQueries(readFileFunc)
	if err != nil {
		return err
	}
	setupFunctionQueries, err := getSetupFunctionQueries(readFileFunc, readDirFunc)
	if err != nil {
		return err
	}
	queries := append(setupTableQueries, setupFunctionQueries...)
	tx, err := db.Begin()
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
func LimitPlayerTypes(playerTypesCsv string) error {
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
		if _, ok := playerTypes[pt]; !ok {
			return fmt.Errorf("unknown PlayerType: %v", pt)
		}
		selectedPlayerTypesMap[pt] = true
		selectedSportTypesMap[pt.SportType()] = true
	}
	// limit PlayerTypes and SportTypes
	for pt := range playerTypes {
		if _, ok := selectedPlayerTypesMap[pt]; !ok {
			delete(playerTypes, pt)
		}
	}
	for st := range sportTypes {
		if _, ok := selectedSportTypesMap[st]; !ok {
			delete(sportTypes, st)
			continue
		}
	}
	return nil
}
