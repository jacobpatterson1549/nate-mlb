package db

import (
	"fmt"
)

type (
	// PlayerType is an enumeration of types of players
	PlayerType int

	// PlayerTypeInfo contains supplementary information about a PlayerType
	PlayerTypeInfo struct {
		SportType    SportType
		Name         string
		Description  string
		ScoreType    string
		DisplayOrder int
	}

	// PlayerTypeMap contains information about multiple PlayerTypes and their PlayerTypeInfos
	PlayerTypeMap map[PlayerType]PlayerTypeInfo
)

// The expected PlayerTypes
const (
	PlayerTypeMlbTeam    PlayerType = 1
	PlayerTypeMlbHitter  PlayerType = 2
	PlayerTypeMlbPitcher PlayerType = 3
	PlayerTypeNflTeam    PlayerType = 4
	PlayerTypeNflQB      PlayerType = 5
	PlayerTypeNflMisc    PlayerType = 6
)

// GetPlayerTypes loads the PlayerTypes from the database
func (ds Datastore) GetPlayerTypes() (PlayerTypeMap, error) {
	playerTypes, err := ds.db.GetPlayerTypes()
	if err != nil {
		return nil, err
	}
	if len(playerTypes) != 6 {
		return nil, fmt.Errorf("did not load expected PlayerTypes: %v", playerTypes)
	}
	return playerTypes, nil
}

func (d sqlDB) GetPlayerTypes() (PlayerTypeMap, error) {
	sqlFunction := newReadSQLFunction("get_player_types", []string{"id", "sport_type_id", "name", "description", "score_type"})
	rs, err := d.db.Query(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return nil, fmt.Errorf("reading playerTypes: %w", err)
	}
	defer rs.Close()

	var (
		id          PlayerType
		sportType   SportType
		name        string
		description string
		scoreType   string
	)
	playerTypes := make(map[PlayerType]PlayerTypeInfo)
	displayOrder := 0
	for rs.Next() {
		err = rs.Scan(&id, &sportType, &name, &description, &scoreType)
		if err != nil {
			return nil, fmt.Errorf("reading player type: %w", err)
		}
		switch id {
		case
			PlayerTypeMlbTeam, PlayerTypeMlbHitter, PlayerTypeMlbPitcher,
			PlayerTypeNflTeam, PlayerTypeNflQB, PlayerTypeNflMisc:
			playerTypes[id] = PlayerTypeInfo{
				SportType:    sportType,
				Name:         name,
				Description:  description,
				ScoreType:    scoreType,
				DisplayOrder: displayOrder,
			}
		default:
			return nil, fmt.Errorf("unknown PlayerType id: %v", id)
		}
		displayOrder++
	}
	return playerTypes, nil
}
