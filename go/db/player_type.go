package db

import (
	"fmt"
	"sort"
)

type (
	// PlayerType is an enumeration of types of players
	PlayerType int

	playerType struct {
		sportType    SportType
		name         string
		description  string
		scoreType    string
		displayOrder int
	}
)

// The expected PlayerTypes
const (
	PlayerTypeMlbTeam PlayerType = 1
	PlayerTypeHitter  PlayerType = 2
	PlayerTypePitcher PlayerType = 3
	PlayerTypeNflTeam PlayerType = 4
	PlayerTypeNflQB   PlayerType = 5
	PlayerTypeNflMisc PlayerType = 6
)

var playerTypes = make(map[PlayerType]playerType)

// PlayerTypes returns the PlayerTypes for a given SportType
func PlayerTypes(st SportType) []PlayerType {
	playerTypesList := make([]PlayerType, 0, len(sportTypes))
	for pt := range playerTypes {
		if pt.SportType() == st {
			playerTypesList = append(playerTypesList, pt)
		}
	}
	sort.Slice(playerTypesList, func(i, j int) bool {
		return playerTypesList[i].DisplayOrder() < playerTypesList[j].DisplayOrder()
	})
	return playerTypesList
}

// SportType gets the SportType for a PlayerType
func (pt PlayerType) SportType() SportType {
	return playerTypes[pt].sportType
}

// Name gets the name for a PlayerType
func (pt PlayerType) Name() string {
	return playerTypes[pt].name
}

// Description gets the description for a PlayerType
func (pt PlayerType) Description() string {
	return playerTypes[pt].description
}

// ScoreType gets the score type for a PlayerType
func (pt PlayerType) ScoreType() string {
	return playerTypes[pt].scoreType
}

// DisplayOrder gets the display order for a PlayerType
func (pt PlayerType) DisplayOrder() int {
	return playerTypes[pt].displayOrder
}

// LoadPlayerTypes loads the PlayerTypes from the database
func LoadPlayerTypes() error {
	rows, err := db.Query("SELECT id, sport_type_id, name, description, score_type FROM get_player_types()")
	if err != nil {
		return fmt.Errorf("reading playerTypes: %w", err)
	}
	defer rows.Close()

	var (
		id          PlayerType
		sportType   SportType
		name        string
		description string
		scoreType   string
	)
	displayOrder := 0
	for rows.Next() {
		err = rows.Scan(&id, &sportType, &name, &description, &scoreType)
		if err != nil {
			return fmt.Errorf("reading player type: %w", err)
		}
		playerType := playerType{
			sportType:    sportType,
			name:         name,
			description:  description,
			scoreType:    scoreType,
			displayOrder: displayOrder,
		}
		switch id {
		case
			PlayerTypeMlbTeam, PlayerTypeHitter, PlayerTypePitcher,
			PlayerTypeNflTeam, PlayerTypeNflQB, PlayerTypeNflMisc:
			playerTypes[id] = playerType
		default:
			return fmt.Errorf("unknown PlayerType id: %v", id)
		}
		displayOrder++
	}
	if len(playerTypes) != 6 {
		return fmt.Errorf("did not load expected amount of PlayerTypes.  Loaded: %d PlayerTypes", len(playerTypes))
	}
	return nil
}
