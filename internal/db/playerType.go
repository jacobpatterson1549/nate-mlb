package db

import "fmt"

// IDs of constant db enums
const (
	PlayerTypeTeam     = 1 // TODO: Make PlayerId and enum type
	PlayerTypeHitting  = 2
	PlayerTypePitching = 3
)

// PlayerType contain a name of a pool item.
type PlayerType struct {
	ID          int
	Name        string
	Description string
}

func getPlayerTypes() ([]PlayerType, error) {
	rows, err := db.Query("SELECT id, name, description FROM player_types ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("problem reading playerTypes: %v", err)
	}
	defer rows.Close()

	playerTypes := []PlayerType{}
	i := 0
	for rows.Next() {
		playerTypes = append(playerTypes, PlayerType{})
		err = rows.Scan(&playerTypes[i].ID, &playerTypes[i].Name, &playerTypes[i].Description)
		if err != nil {
			return nil, fmt.Errorf("problem reading data: %v", err)
		}
		i++
	}
	return playerTypes, nil
}
