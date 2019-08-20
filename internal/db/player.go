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
func GetPlayers(st SportType) ([]Player, error) {
	rows, err := db.Query(
		"SELECT p.id, p.display_order, p.player_type_id, p.player_id, p.friend_id FROM players AS p JOIN stats AS s ON p.year = s.year JOIN friends AS f ON p.friend_id = f.id WHERE s.sport_type_id = $1 AND s.active ORDER BY p.player_type_id, f.display_order, p.display_order",
		st,
	)
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
			return nil, fmt.Errorf("problem reading player: %v", err)
		}
		i++
	}
	return players, nil
}

// SavePlayers saves the specified players for the active year
func SavePlayers(st SportType, futurePlayers []Player) error {
	players, err := GetPlayers(st)
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
			sql:  "INSERT INTO players (display_order, player_type_id, player_id, friend_id, sport_type_id, year) SELECT $1, $2, $3, $4, $5, year FROM stats AS s WHERE s.sport_type_id = $5 AND s.active",
			args: make([]interface{}, 5),
		}
		queries[i].args[0] = insertPlayer.DisplayOrder
		queries[i].args[1] = insertPlayer.PlayerType
		queries[i].args[2] = insertPlayer.PlayerID
		queries[i].args[3] = insertPlayer.FriendID
		queries[i].args[4] = st
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
