package db

import (
	"fmt"
)

// Player maps a player (of a a specific PlayerType) to a Friend.
type Player struct {
	ID           int
	DisplayOrder int
	PlayerType   PlayerType
	PlayerID     int
	FriendID     int
}

// GetPlayers gets the players for the active year
func GetPlayers() ([]Player, error) {
	rows, err := db.Query("SELECT p.id, p.display_order, p.player_type_id, p.player_id, p.friend_id FROM players AS p JOIN stats AS s ON p.year = s.year WHERE s.active ORDER BY p.player_type_id, p.friend_id, p.display_order")
	if err != nil {
		return nil, fmt.Errorf("problem reading players: %v", err)
	}
	defer rows.Close()

	var players []Player
	i := 0
	for rows.Next() {
		players = append(players, Player{})
		err = rows.Scan(&players[i].ID, &players[i].DisplayOrder, &players[i].PlayerType, &players[i].PlayerID, &players[i].FriendID)
		if err != nil {
			return nil, fmt.Errorf("problem reading data: %v", err)
		}
		i++
	}
	return players, nil
}

// SavePlayers saves the specified players for the active year
func SavePlayers(futurePlayers []Player) error {
	players, err := GetPlayers()
	if err != nil {
		return err
	}
	previousPlayers := make(map[int]Player)
	for _, player := range players {
		previousPlayers[player.ID] = player
	}

	var insertPlayers []Player
	var updatePlayers []Player
	for _, player := range futurePlayers {
		previousPlayer, ok := previousPlayers[player.ID]
		if !ok {
			insertPlayers = append(insertPlayers, player)
		} else if player.DisplayOrder != previousPlayer.DisplayOrder { // can only update display order
			updatePlayers = append(updatePlayers, player)
		}
		delete(previousPlayers, player.ID)
	}

	queries := make([]query, len(insertPlayers)+len(updatePlayers)+len(previousPlayers))
	i := 0
	for _, insertPlayer := range insertPlayers {
		queries[i] = query{
			sql:  "INSERT INTO players (display_order, player_type_id, player_id, friend_id, year) SELECT $1, $2, $3, $4, year FROM stats AS s WHERE s.active",
			args: make([]interface{}, 4),
		}
		queries[i].args[0] = insertPlayer.DisplayOrder
		queries[i].args[1] = insertPlayer.PlayerType
		queries[i].args[2] = insertPlayer.PlayerID
		queries[i].args[3] = insertPlayer.FriendID
		i++
	}
	for _, updateplayer := range updatePlayers {
		queries[i] = query{
			sql:  "UPDATE players SET display_order = $1 WHERE id = $2",
			args: make([]interface{}, 2),
		}
		queries[i].args[0] = updateplayer.DisplayOrder
		queries[i].args[1] = updateplayer.ID
		i++
	}
	for deleteID := range previousPlayers {
		queries[i] = query{
			sql:  "DELETE FROM players WHERE id = $1",
			args: make([]interface{}, 1),
		}
		queries[i].args[0] = deleteID
		i++
	}
	return exececuteInTransaction(&queries)
}
