package db

import (
	"os"
	"time"
)

// Datastore interface can be used to access and persist data
type (
	Datastore interface {
		// Stat
		GetStat(st SportType) (*Stat, error)
		SetStat(stat Stat) error
		ClearStat(st SportType) error

		// Year
		GetYears(st SportType) ([]Year, error)
		SaveYears(st SportType, futureYears []Year) error

		// Friend
		GetFriends(st SportType) ([]Friend, error)
		SaveFriends(st SportType, futureFriends []Friend) error

		// Player
		GetPlayers(st SportType) ([]Player, error)
		SavePlayers(st SportType, futurePlayers []Player) error

		// User
		SetUserPassword(username string, p Password) error
		AddUser(username string, p Password) error
		IsCorrectUserPassword(username string, p Password) (bool, error)
		SetAdminPassword(p Password) error // TODO: move to other package - maybe admin.go?

		// Setup
		LimitPlayerTypes(playerTypesCsv string) error
		SetupTablesAndFunctions() error
		getSetupTableQueries(readFileFunc func(filename string) ([]byte, error)) ([]string, error)
		getSetupFunctionQueries(readFileFunc func(filename string) ([]byte, error), readDirFunc func(dirname string) ([]os.FileInfo, error)) ([]string, error)

		// DB
		Ping() error
		GetUtcTime() time.Time

		// fields
		SportTypes() map[SportType]SportTypeInfo
		PlayerTypes() map[PlayerType]PlayerTypeInfo
	}

	sqlDatastore struct {
		db          database
		ph          passwordHasher
		sportTypes  map[SportType]SportTypeInfo
		playerTypes map[PlayerType]PlayerTypeInfo
	}
)

// NewDatastore creates a new sqlDatastore
func NewDatastore(dataSourceName string, sportTypes map[SportType]SportTypeInfo, playerTypes map[PlayerType]PlayerTypeInfo) (Datastore, error) {
	db, err := newSQLDatabase(dataSourceName)
	if err != nil {
		return nil, err
	}
	ds := sqlDatastore{
		db:          db,
		ph:          bcryptPasswordHasher{},
		sportTypes:  sportTypes,
		playerTypes: playerTypes,
	}
	return ds, nil
}

// SportTypes implements the Datastore interface
func (ds sqlDatastore) SportTypes() map[SportType]SportTypeInfo {
	return ds.sportTypes
}

// SportTypes implements the Datastore interface
func (ds sqlDatastore) PlayerTypes() map[PlayerType]PlayerTypeInfo {
	return ds.playerTypes
}
