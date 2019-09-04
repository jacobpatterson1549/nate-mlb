package db

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func setup() error {
	// to remove tables, indexes and data, run this line and not the code below
	//db.Exec("DROP TABLE IF EXISTS users, sport_types, stats, friends, player_types, players")

	setupFuncs := []func() error{
		setupTablesAndFunctions,
		LoadSportTypes,
		LoadPlayerTypes,
	}
	for _, setupFunc := range setupFuncs {
		if err := setupFunc(); err != nil {
			return err
		}
	}
	return nil
}

func setupTablesAndFunctions() error {
	var queries []string
	// add setup queries first
	// order of setup files matters - some queries reference others
	setupFileNames := []string{"users", "sport_types", "stats", "years", "friends", "player_types", "players"}
	for _, setupFileName := range setupFileNames {
		b, err := ioutil.ReadFile(fmt.Sprintf("sql/setup/%s.sql", setupFileName))
		if err != nil {
			return err
		}
		setupQueries := strings.Split(string(b), ";")
		queries = append(queries, setupQueries...)
	}
	// add the function queries
	functionDirTypes := []string{"get", "set"}
	for _, functionDirType := range functionDirTypes {
		functionDir := fmt.Sprintf("sql/functions/%s", functionDirType)
		functionFileInfos, err := ioutil.ReadDir(functionDir)
		if err != nil {
			return fmt.Errorf("problem reading functions directory: %v", err)
		}
		for _, functionFileInfo := range functionFileInfos {
			b, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", functionDir, functionFileInfo.Name()))
			if err != nil {
				return err
			}
			queries = append(queries, string(b))
		}
	}
	// run all the queries in a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, query := range queries {
		_, err := tx.Exec(query)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("%v, ROLLBACK ERROR: %v", err, rollbackErr)
			}
			return err
		}
	}
	return tx.Commit()
}
