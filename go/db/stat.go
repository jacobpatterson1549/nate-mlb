package db

import (
	"fmt"
	"time"
)

type (
	// Stat is a wrapper for EtlJSON
	// It is for a particular year and SportType.  It has an etl timestamp.
	Stat struct {
		SportType    SportType
		Year         int
		EtlTimestamp *time.Time
		EtlJSON      *[]byte
	}
)

// GetStat gets the Stat for the active year, nil if there is not active stat
func (ds Datastore) GetStat(st SportType) (*Stat, error) {
	return ds.db.GetStat(st)
}

func (d sqlDB) GetStat(st SportType) (*Stat, error) {
	stat := Stat{SportType: st}
	sqlFunction := newReadSQLFunction("get_stat", []string{"year", "etl_timestamp", "etl_json"}, st)
	row := d.db.QueryRow(sqlFunction.sql(), sqlFunction.args...)
	err := row.Scan(&stat.Year, &stat.EtlTimestamp, &stat.EtlJSON)
	if err != nil {
		if d.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting stats: %w", err)
	}
	return &stat, nil
}

// SetStat sets the etl timestamp and json for the year (which must be active)
func (ds Datastore) SetStat(stat Stat) error {
	return ds.db.SetStat(stat)
}

func (d *sqlDB) SetStat(stat Stat) error {
	sqlFunction := newWriteSQLFunction("set_stat", stat.EtlTimestamp, stat.EtlJSON, stat.SportType, stat.Year)
	result, err := d.db.Exec(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return fmt.Errorf("saving stats: %w", err)
	}
	return expectSingleRowAffected(result)
}

// ClearStat clears the stats for the active year
func (ds Datastore) ClearStat(st SportType) error {
	return ds.db.ClrStat(st)
}

func (d *sqlDB) ClrStat(st SportType) error {
	sqlFunction := newWriteSQLFunction("clr_stat", st)
	_, err := d.db.Exec(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return fmt.Errorf("clearing saved stats: %w", err)
	}
	return nil
}
