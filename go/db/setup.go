package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func setup() error {
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
	return setAdminPassword()
}

func getSetupTableQueries() ([]string, error) {
	var queries []string
	// order of setup files matters - some queries reference others
	setupFileNames := []string{"users", "sport_types", "stats", "friends", "player_types", "players"}
	for _, setupFileName := range setupFileNames {
		b, err := ioutil.ReadFile(fmt.Sprintf("sql/setup/%s.sql", setupFileName))
		if err != nil {
			return nil, err
		}
		setupQueries := strings.Split(string(b), ";")
		queries = append(queries, setupQueries...)
	}
	return queries, nil
}

func getSetupFunctionQueries() ([]string, error) {
	var queries []string
	functionDirTypes := []string{"add", "clr", "del", "get", "set"}
	for _, functionDirType := range functionDirTypes {
		functionDir := fmt.Sprintf("sql/functions/%s", functionDirType)
		functionFileInfos, err := ioutil.ReadDir(functionDir)
		if err != nil {
			return nil, err
		}
		for _, functionFileInfo := range functionFileInfos {
			b, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", functionDir, functionFileInfo.Name()))
			if err != nil {
				return nil, err
			}
			queries = append(queries, string(b))
		}
	}
	return queries, nil
}

func setupTablesAndFunctions() error {
	setupTableQueries, err := getSetupTableQueries()
	if err != nil {
		return err
	}
	setupFunctionQueries, err := getSetupFunctionQueries()
	if err != nil {
		return err
	}
	queries := append(setupTableQueries, setupFunctionQueries...)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, sql := range queries {
		_, err := tx.Exec(sql)
		if err != nil {
			err = fmt.Errorf("setting: %w\nquery: %v", err, strings.TrimSpace(sql))
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("%w, ROLLBACK ERROR: %w", err, rollbackErr)
			}
			return err
		}
	}
	return tx.Commit()
}

func setAdminPassword() error {
	username := "admin"
	_, err := getUserPassword(username)
	switch {
	case err == nil:
		return nil
	case !errors.As(err, &sql.ErrNoRows):
		return err
	}
	fmt.Printf("INITIAL SETUP: Enter password for '%s' user: ", username)
	var password password
	for {
		_, err := fmt.Scanln(&password)
		if err != nil {
			fmt.Printf("\ninvalid password: %v.  Enter again: ", err)
			continue // TODO: debug this (continue does not seem to work)
		}
		break
	}
	return AddUser(username, string(password))
}

type password string

func (p password) isValid() bool {
	whitespaceRE := regexp.MustCompile("\\s")
	return !whitespaceRE.MatchString(string(p)) && len(p) != 0
}
