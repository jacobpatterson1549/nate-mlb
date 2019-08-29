package db

import "fmt"

// PlayerType is an enumeration of types of players
type PlayerType int

// The expected PlayerTypes
const (
	PlayerTypeMlbTeam PlayerType = 1
	PlayerTypeHitter  PlayerType = 2
	PlayerTypePitcher PlayerType = 3
	PlayerTypeNflTeam PlayerType = 4
	PlayerTypeNflQB   PlayerType = 5
	PlayerTypeNflMisc PlayerType = 6
)

var (
	playerTypeSportTypes   = make(map[PlayerType]SportType)
	playerTypeNames        = make(map[PlayerType]string)
	playerTypeDescriptions = make(map[PlayerType]string)
)

// SportType gets the SportType for a PlayerType
func (pt PlayerType) SportType() SportType {
	return playerTypeSportTypes[pt]
}

// Name gets the name for a PlayerType
func (pt PlayerType) Name() string {
	return playerTypeNames[pt]
}

// Description gets the name for a PlayerType
func (pt PlayerType) Description() string {
	return playerTypeDescriptions[pt]
}

// LoadPlayerTypes loads the PlayerTypes from the database
func LoadPlayerTypes(st SportType) ([]PlayerType, error) {
	rows, err := db.Query(
		`SELECT id, sport_type_id, name, description
		FROM player_types
		WHERE sport_type_id = $1
		ORDER BY id ASC`,
		st,
	)
	if err != nil {
		return nil, fmt.Errorf("problem reading playerTypes: %v", err)
	}
	defer rows.Close()

	var (
		playerTypes []PlayerType
		playerType  PlayerType
		sportType   SportType
		name        string
		description string
	)
	for rows.Next() {
		err = rows.Scan(&playerType, &sportType, &name, &description)
		if err != nil {
			return nil, fmt.Errorf("problem reading player type: %v", err)
		}
		switch playerType {
		case PlayerTypeMlbTeam, PlayerTypeHitter, PlayerTypePitcher,
			PlayerTypeNflTeam, PlayerTypeNflQB, PlayerTypeNflMisc:
			playerTypeSportTypes[playerType] = sportType
			playerTypeNames[playerType] = name
			playerTypeDescriptions[playerType] = description
			playerTypes = append(playerTypes, playerType)
		default:
			return nil, fmt.Errorf("Unknown PlayerType: id=%d", playerType)
		}
	}
	switch {
	case st == SportTypeMlb && len(playerTypes) == 3,
		st == SportTypeNfl && len(playerTypes) == 3:
		return playerTypes, nil
	default:
		return nil, fmt.Errorf("Did not load expected amount of PlayerTypes.  Loaded: %d", len(playerTypes))
	}
}
