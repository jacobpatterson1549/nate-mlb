package db

import "fmt"

// PlayerType is an enumeration of types of players
type PlayerType interface {
	ID() int
	Name() string
	Description() string
	ScoreType() string
	DisplayOrder() int
}

type playerType struct {
	id           int
	sportType    SportType
	name         string
	description  string
	scoreType    string
	displayOrder int
}

// The expected PlayerTypes
var (
	PlayerTypeMlbTeam    PlayerType
	PlayerTypeHitter     PlayerType
	PlayerTypePitcher    PlayerType
	PlayerTypeNflTeam    PlayerType
	PlayerTypeNflQB      PlayerType
	PlayerTypeNflMisc    PlayerType
	sportTypePlayerTypes = make(map[SportType][]PlayerType)
)

// GetPlayerTypes returns the PlayerTypes for a given SportType
func GetPlayerTypes(st SportType) []PlayerType {
	return sportTypePlayerTypes[st]
}

// GetPlayerType returns the PlayerType with the specified id for a given SportType
func GetPlayerType(st SportType, id int) PlayerType {
	for _, pt := range sportTypePlayerTypes[st] {
		if pt.ID() == id {
			return pt
		}
	}
	return nil
}

// ID gets the id for a PlayerType
func (pt playerType) ID() int {
	return pt.id
}

// SportType gets the SportType for a PlayerType
func (pt playerType) SportType() SportType {
	return pt.sportType
}

// Name gets the name for a PlayerType
func (pt playerType) Name() string {
	return pt.name
}

// Description gets the description for a PlayerType
func (pt playerType) Description() string {
	return pt.description
}

// ScoreType gets the score type for a PlayerType
func (pt playerType) ScoreType() string {
	return pt.scoreType
}

// DisplayOrder gets the display order for a PlayerType
func (pt playerType) DisplayOrder() int {
	return pt.displayOrder
}

// LoadPlayerTypes loads the PlayerTypes from the database
func LoadPlayerTypes() error {
	rows, err := db.Query("SELECT id, sport_type_id, name, description, score_type FROM get_player_types()")
	if err != nil {
		return fmt.Errorf("reading playerTypes: %w", err)
	}
	defer rows.Close()

	var (
		id          int
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
		pt := playerType{
			id:           id,
			sportType:    sportType,
			name:         name,
			description:  description,
			displayOrder: displayOrder,
			// TODO: investigate if SportType should have displayOrder.  Also, should this be determined by the row order in the database?
		}
		switch id {
		case 1:
			PlayerTypeMlbTeam = pt
		case 2:
			PlayerTypeHitter = pt
		case 3:
			PlayerTypePitcher = pt
		case 4:
			PlayerTypeNflTeam = pt
		case 5:
			PlayerTypeNflQB = pt
		case 6:
			PlayerTypeNflMisc = pt
		default:
			return fmt.Errorf("unknown PlayerType id: %v", id)
		}
		sportTypePlayerTypes[sportType] = append(sportTypePlayerTypes[sportType], pt)
		displayOrder++
	}
	if len(sportTypePlayerTypes) != 2 ||
		len(sportTypePlayerTypes[SportTypeNfl]) != 3 ||
		len(sportTypePlayerTypes[SportTypeMlb]) != 3 {
		return fmt.Errorf("did not load expected amount of PlayerTypes.  Loaded: %d SportTypes", len(sportTypePlayerTypes))
	}
	return nil
}
