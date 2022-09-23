// Package db contains persistence functions to store and query data for the server.
package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"strings"
	"time"
)

type (
	// ID is used to identify an item in the database or a relation to another noun's id
	ID string

	readSQLFunction struct {
		name string
		cols []string
		args []interface{}
	}

	writeSQLFunction struct {
		name string
		args []interface{}
	}

	datastoreConfig struct {
		dataSourceName string
		ph             passwordHasher
		log            *log.Logger
		fs             fs.ReadFileFS
	}

	// Datastore interface can be used to access and persist data
	Datastore struct {
		db          db
		ph          passwordHasher
		sportTypes  SportTypeMap
		playerTypes PlayerTypeMap
		log         *log.Logger
	}

	db interface {
		begin() (dbTX, error) // returning the dbTX interface is smelly
		GetSportTypes() (SportTypeMap, error)
		GetPlayerTypes() (PlayerTypeMap, error)
		GetYears(st SportType) ([]Year, error)
		GetStat(st SportType) (*Stat, error)
		SetStat(stat Stat) error
		ClrStat(st SportType) error
		GetFriends(st SportType) ([]Friend, error)
		GetPlayers(st SportType) ([]Player, error)
		GetUserPassword(username string) (string, error)
		SetUserPassword(username, hashedPassword string) error
		AddUser(username, hashedPassword string) error
		// IsNotExist is used by the datastore to determine if a query failed because data does not exist.
		IsNotExist(err error) bool
	}

	dbTX interface {
		execute() error
		AddYear(st SportType, year int)
		DelYear(st SportType, year int)
		SetYearActive(st SportType, year int)
		ClrYearActive(st SportType)
		AddFriend(st SportType, displayOrder int, name string)
		SetFriend(st SportType, id ID, displayOrder int, name string)
		DelFriend(st SportType, id ID)
		AddPlayer(st SportType, displayOrder int, pt PlayerType, sourceID SourceID, friendID ID)
		SetPlayer(st SportType, id ID, displayOrder int)
		DelPlayer(st SportType, id ID)
	}
)

// NewDatastore creates a new sqlDatastore
func NewDatastore(dataSourceName string, log *log.Logger, fs fs.ReadFileFS) (*Datastore, error) {
	cfg := datastoreConfig{
		dataSourceName: dataSourceName,
		ph:             bcryptPasswordHasher{},
		log:            log,
		fs:             fs,
	}
	db, err := cfg.newDatabase()
	if err != nil {
		return nil, err
	}
	return cfg.newDatastore(db)
}

// newDatabase creates a database from the dataSourceName in the config
func (cfg datastoreConfig) newDatabase() (db, error) {
	url, err := url.Parse(cfg.dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("parsing data source: %w", err)
	}
	var db db
	switch url.Scheme {
	case "postgres":
		db, err = newSQLDatabase(url.Scheme, cfg.dataSourceName)
	case "firestore":
		projectID := url.Host
		db, err = newFirestoreDB(projectID)
	}
	if err != nil {
		return nil, err
	}
	return db, nil
}

// tewDataStore creates a datastore using the database
func (cfg datastoreConfig) newDatastore(db db) (*Datastore, error) {

	ds := Datastore{
		db:  db,
		ph:  cfg.ph,
		log: cfg.log,
	}

	if d, ok := db.(*sqlDB); ok {
		if err := d.SetupTablesAndFunctions(cfg.fs); err != nil {
			return nil, err
		}
	}

	sportTypes, err := ds.GetSportTypes()
	if err != nil {
		return nil, err
	}
	ds.sportTypes = sportTypes

	playerTypes, err := ds.GetPlayerTypes()
	if err != nil {
		return nil, err
	}
	ds.playerTypes = playerTypes

	return &ds, nil
}

// GetUtcTime retrieves the current UTC time
func (Datastore) GetUtcTime() time.Time {
	return time.Now().UTC()
}

// SportTypes implements the SportTypeGetter interface for Datastore
func (ds Datastore) SportTypes() SportTypeMap {
	return ds.sportTypes
}

// PlayerTypes implements the PlayerTypeGetter interface for Datastore
func (ds Datastore) PlayerTypes() PlayerTypeMap {
	return ds.playerTypes
}

func (t *sqlTX) execute() error {
	var result sql.Result
	var err error
	for _, sqlFunction := range t.queries {
		result, err = t.tx.Exec(sqlFunction.sql(), sqlFunction.args...)
		if err == nil {
			err = expectSingleRowAffected(result)
		}
		if err != nil {
			err = fmt.Errorf("%s: %w", sqlFunction.name, err)
			break
		}
	}
	switch {
	case err != nil:
		err = fmt.Errorf("saving: %w", err)
		rollbackErr := t.tx.Rollback()
		if rollbackErr != nil {
			err = fmt.Errorf("%v and ROLLBACK ERROR: %w", err, rollbackErr)
		}
	case len(t.queries) > 0:
		err = t.tx.Commit()
		if err != nil {
			err = fmt.Errorf("committing transaction to save: %w", err)
		}
	}
	return err
}

func expectSingleRowAffected(r sql.Result) error {
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("expected to update 1 row, but updated %d", rows)
	}
	return nil
}

func newReadSQLFunction(name string, cols []string, args ...interface{}) readSQLFunction {
	return readSQLFunction{
		name: name,
		cols: cols,
		args: args,
	}
}

func newWriteSQLFunction(name string, args ...interface{}) writeSQLFunction {
	return writeSQLFunction{
		name: name,
		args: args,
	}
}

func (f readSQLFunction) sql() string {
	argIndexes := make([]string, len(f.args))
	for i := range argIndexes {
		argIndexes[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf("SELECT %s FROM %s(%s)", strings.Join(f.cols, ", "), f.name, strings.Join(argIndexes, ", "))
}

func (f writeSQLFunction) sql() string {
	argIndexes := make([]string, len(f.args))
	for i := range argIndexes {
		argIndexes[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf("SELECT %s(%s)", f.name, strings.Join(argIndexes, ", "))
}

func (id *ID) Scan(src interface{}) error {
	switch t := src.(type) {
	case int, int64:
		*id = ID(fmt.Sprint(src))
	case string:
		*id = ID(src.(string))
	default:
		return fmt.Errorf("unsupported Scan, storing %v in %T", t, id)
	}
	return nil
}
