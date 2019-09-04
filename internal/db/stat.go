package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Stat is a wrapper for EtlJSON
// It is for a particular year and SportType.  It has an etl timestamp.
type Stat struct {
	SportType    SportType
	Year         int
	EtlTimestamp *time.Time
	EtlJSON      *string
}

// GetStat gets the Stat for the active year
func GetStat(st SportType) (Stat, error) {
	var etlJSON sql.NullString
	var stat Stat
	row := db.QueryRow(
		`SELECT sport_type_id, year, etl_timestamp, etl_json FROM get_stat($1)`,
		st,
	)
	err := row.Scan(&stat.SportType, &stat.Year, &stat.EtlTimestamp, &etlJSON)
	if err != nil {
		return stat, fmt.Errorf("problem getting stats: %v", err)
	}

	if etlJSON.Valid {
		stat.EtlJSON = &etlJSON.String
	}
	return stat, nil
}

// SetStat sets the etl timestamp and json for the year (which must be active)
func SetStat(stat Stat) error {
	result, err := db.Exec(
		`SELECT set_stat($1, $2, $3, $4)`,
		stat.EtlTimestamp, stat.EtlJSON, stat.SportType, stat.Year)
	if err != nil {
		return fmt.Errorf("problem saving stats current year: %v", err)
	}
	return expectSingleRowAffected(result)
}

// ClearStat clears the stats for the active year
func ClearStat(st SportType) error {
	_, err := db.Exec(
		`SELECT clr_stat($1)`,
		st,
	)
	if err != nil {
		return fmt.Errorf("problem clearing saved stats: %v", err)
	}
	return nil
}
