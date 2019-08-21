package db

import (
	"database/sql"
	"fmt"
)

// GetEtlStatsJSON gets the stats for the current year
func GetEtlStatsJSON(st SportType) (string, error) {
	var etlJSON sql.NullString
	row := db.QueryRow(
		"SELECT etl_json FROM stats WHERE sport_type_id = $1 AND active",
		st,
	)
	err := row.Scan(&etlJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("no active year to get previous stats for (sportType %v)", st)
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
func SetEtlStats(st SportType, etlStatsJSON string) error {
	result, err := db.Exec("UPDATE stats SET etl_json = $1 WHERE sport_type_id = $2 AND active", etlStatsJSON, st)
	if err != nil {
		return fmt.Errorf("problem saving stats current year: %v", err)
	}
	return expectSingleRowAffected(result)
}

// ClearEtlStats clears the stats for the current year
func ClearEtlStats(st SportType) error {
	_, err := db.Exec(
		"UPDATE stats SET etl_json = NULL WHERE sport_type_id = $1 AND active",
		st,
	)
	if err != nil {
		return fmt.Errorf("problem clearing saved stats: %v", err)
	}
	return nil
}
