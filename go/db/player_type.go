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
	playerTypes             = make(map[SportType][]PlayerType)
	playerTypeSportTypes    = make(map[PlayerType]SportType)
	playerTypeNames         = make(map[PlayerType]string)
	playerTypeDescriptions  = make(map[PlayerType]string)
	playerTypeScoreTypes    = make(map[PlayerType]string)
	playerTypeDisplayOrders = make(map[PlayerType]int)
)

// GetPlayerTypes returns the PlayerTyeps for a given SportType
func GetPlayerTypes(st SportType) []PlayerType {
	return playerTypes[st]
}

// SportType gets the SportType for a PlayerType
func (pt PlayerType) SportType() SportType {
	return playerTypeSportTypes[pt]
}

// Name gets the name for a PlayerType
func (pt PlayerType) Name() string {
	return playerTypeNames[pt]
}

// Description gets the description for a PlayerType
func (pt PlayerType) Description() string {
	return playerTypeDescriptions[pt]
}

// ScoreType gets the score type for a PlayerType
func (pt PlayerType) ScoreType() string {
	return playerTypeScoreTypes[pt]
}

// DisplayOrder gets the display order for a PlayerType
func (pt PlayerType) DisplayOrder() int {
	return playerTypeDisplayOrders[pt]
}

// LoadPlayerTypes loads the PlayerTypes from the database
func LoadPlayerTypes() error {
	rows, err := db.Query("SELECT id, sport_type_id, name, description, score_type FROM get_player_types()")
	if err != nil {
		return fmt.Errorf("reading playerTypes: %w", err)
	}
	defer rows.Close()

	var (
		playerType  PlayerType
		sportType   SportType
		name        string
		description string
		scoreType   string
	)
	displayOrder := 0
	for rows.Next() {
		err = rows.Scan(&playerType, &sportType, &name, &description, &scoreType)
		if err != nil {
			return fmt.Errorf("reading player type: %w", err)
		}
		switch playerType {
		case PlayerTypeMlbTeam, PlayerTypeHitter, PlayerTypePitcher,
			PlayerTypeNflTeam, PlayerTypeNflQB, PlayerTypeNflMisc:
			playerTypes[sportType] = append(playerTypes[sportType], playerType)
			playerTypeSportTypes[playerType] = sportType
			playerTypeNames[playerType] = name
			playerTypeDescriptions[playerType] = description
			playerTypeScoreTypes[playerType] = scoreType
			playerTypeDisplayOrders[playerType] = displayOrder
		default:
			return fmt.Errorf("unknown PlayerType: id=%d", playerType)
		}
		displayOrder++
	}
	if len(playerTypes) != 2 ||
		len(playerTypes[SportTypeNfl]) != 3 ||
		len(playerTypes[SportTypeMlb]) != 3 {
		return fmt.Errorf("did not load expected amount of PlayerTypes.  Loaded: %d SportTypes", len(playerTypes))
	}
	return nil
}
