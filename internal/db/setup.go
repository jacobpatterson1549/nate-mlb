package db

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func setup() error {
	// to remove tables, indexes and data, run this line and not the setup files below
	// db.Exec("DROP TABLE IF EXISTS users, sport_types, stats, friends, player_types, players")

	// order of setup files matters - some tables reference others
	setupFileNames := []string{"users", "sport_types", "stats", "friends", "player_types", "players"}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	rollbackTransaction := func(err error) error {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			err = fmt.Errorf("%v, ROLLBACK ERROR: %v", err, rollbackErr)
		}
		return err
	}
	for _, setupFileName := range setupFileNames {
		b, err := ioutil.ReadFile(fmt.Sprintf("sql/setup/%s.sql", setupFileName))
		if err != nil {
			return rollbackTransaction(err)
		}
		queries := strings.Split(string(b), ";")
		for _, query := range queries {
			_, err := tx.Exec(query)
			if err != nil {
				err = fmt.Errorf("problem executing %s setup file: %v", setupFileName, err)
				return rollbackTransaction(err)
			}
		}
	}
	// TODO: add stored functions
	return tx.Commit()
}
