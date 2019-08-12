package db

import "fmt"

// PlayerType is an enumeration of types of players
type PlayerType int

// The expected PlayerTypes
const (
	PlayerTypeTeam    PlayerType = 1
	PlayerTypeHitter  PlayerType = 2
	PlayerTypePitcher PlayerType = 3
)

// Name gets the name for a PlayerType
func (pt *PlayerType) Name() string {
	return playerTypeNames[*pt]
}

// Description gets the name for a PlayerType
func (pt *PlayerType) Description() string {
	return playerTypeDescriptions[*pt]
}

var playerTypeNames = make(map[PlayerType]string)
var playerTypeDescriptions = make(map[PlayerType]string)

// LoadPlayerTypes loads the PlayerTypes from the database
func LoadPlayerTypes() ([]PlayerType, error) {
	rows, err := db.Query("SELECT id, name, description FROM player_types ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("problem reading playerTypes: %v", err)
	}
	defer rows.Close()

	var (
		playerType  PlayerType
		name        string
		description string
	)
	i := 0
	for rows.Next() {
		err = rows.Scan(&playerType, &name, &description)
		if err != nil {
			return nil, fmt.Errorf("problem reading data: %v", err)
		}
		playerTypeNames[playerType] = name
		playerTypeDescriptions[playerType] = description
		i++
	}
	return []PlayerType{PlayerTypeTeam, PlayerTypeHitter, PlayerTypePitcher}, nil
}
