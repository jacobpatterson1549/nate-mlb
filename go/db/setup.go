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
	return nil
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
		return fmt.Errorf("starting database setup: %w", err)
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
	err = tx.Commit()
	if err == nil {
		return fmt.Errorf("committing database setup: %w", err)
	}
	return nil
}

// SetAdminPassword sets the admin password
func SetAdminPassword(p string) error {
	username := "admin"
	password := password(p)
	_, err := getUserPassword(username)
	switch {
	case !password.isValid():
		return errors.New("password cannot contain spaces")
	case err == nil:
		return SetUserPassword(username, string(password))
	case !errors.As(err, &sql.ErrNoRows):
		return err
	default:
		return AddUser(username, string(password))
	}
}

type password string

func (p password) isValid() bool {
	whitespaceRE := regexp.MustCompile("\\s")
	return len(p) > 0 && !whitespaceRE.MatchString(string(p))
}
