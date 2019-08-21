package db

import (
	"database/sql"
	"fmt"
)

// Stat is a wrapper for EtlStatsJSON and the year the stats are for
type Stat struct {
	EtlStatsJSON string
	Year         int
}

// GetEtlStatsJSON gets the stats for the current year
func GetEtlStatsJSON(st SportType) (Stat, error) {
	var etlJSON sql.NullString
	var stat Stat
	row := db.QueryRow(
		"SELECT etl_json, year FROM stats WHERE sport_type_id = $1 AND active",
		st,
	)
	err := row.Scan(&etlJSON, &stat.Year)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("no active year to get previous stats for (sportType %v)", st)
		} else {
			err = fmt.Errorf("problem getting stats: %v", err)
		}
		return stat, err
	}

	if etlJSON.Valid {
		stat.EtlStatsJSON = etlJSON.String
	}
	return stat, nil
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
