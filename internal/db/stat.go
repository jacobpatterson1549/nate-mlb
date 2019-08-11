package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// GetEtlStatsJSON gets the stats for the current year
func GetEtlStatsJSON() (string, error) {
	var etlJSON sql.NullString
	row := db.QueryRow("SELECT etl_json FROM stats WHERE active")
	err := row.Scan(&etlJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("no active year")
		} else {
			err = fmt.Errorf("problem getting stats: %v", err)
		}
		return "", err
	}

	if !etlJSON.Valid {
		return "", nil
	}
	return etlJSON.String, nil
}

// SetEtlStats sets the stats for the current year
func SetEtlStats(etlStatsJSON string) error {
	result, err := db.Exec("UPDATE stats SET etl_json = $1 WHERE active", etlStatsJSON)
	if err != nil {
		return fmt.Errorf("problem saving stats current year: %v", err)
	}
	return expectSingleRowAffected(result)
}

// ClearEtlStats clears the stats for the current year
func ClearEtlStats() error {
	_, err := db.Exec("UPDATE stats SET etl_json = NULL WHERE active")
	if err != nil {
		return fmt.Errorf("problem clearing saved stats: %v", err)
	}
	return nil
}
