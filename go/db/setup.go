package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type password string

var getSetupFileContents = ioutil.ReadFile
var getSetupFunctionDirContents = ioutil.ReadDir

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
	setupFileTitles := []string{"users", "sport_types", "stats", "friends", "player_types", "players"}
	for _, setupFileTitle := range setupFileTitles {
		tableFileName := fmt.Sprintf("sql/setup/%s.pgsql", setupFileTitle)
		b, err := getSetupFileContents(tableFileName)
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
		functionDirName := fmt.Sprintf("sql/functions/%s", functionDirType)
		functionFileInfos, err := getSetupFunctionDirContents(functionDirName)
		if err != nil {
			return nil, err
		}
		for _, functionFileInfo := range functionFileInfos {
			functionFileName := fmt.Sprintf("%s/%s", functionDirName, functionFileInfo.Name())
			b, err := getSetupFileContents(functionFileName)
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
	if err != nil {
		return fmt.Errorf("committing database setup: %w", err)
	}
	return nil
}

// SetAdminPassword sets the admin password
// If the admin user does not exist, it is created.
func SetAdminPassword(p string) error {
	username := "admin"
	password := password(p)
	if !password.isValid() {
		return errors.New("password cannot contain spaces")
	}
	_, err := getUserPassword(username)
	switch {
	case err == nil:
		return SetUserPassword(username, string(password))
	case !errors.As(err, &sql.ErrNoRows):
		return err
	default:
		return AddUser(username, string(password))
	}
}

func (p password) isValid() bool {
	whitespaceRE := regexp.MustCompile("\\s")
	return len(p) > 0 && !whitespaceRE.MatchString(string(p))
}
