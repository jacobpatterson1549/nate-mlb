package db

import (
	"fmt"
)

// Player maps a player (of a a specific PlayerType) to a Friend.
type Player struct {
	ID           ID
	DisplayOrder int
	PlayerType   PlayerType
	SourceID     SourceID
	FriendID     ID
}

// SourceID is the id used to retrieve information about the player from external sources
type SourceID int

// GetPlayers gets the players for the active year
func GetPlayers(st SportType) ([]Player, error) {
	rows, err := db.Query("SELECT id, player_type_id, source_id, friend_id, display_order FROM get_players($1)", st)
	if err != nil {
		return nil, fmt.Errorf("problem reading players: %v", err)
	}
	defer rows.Close()

	var players []Player
	i := 0
	for rows.Next() {
		players = append(players, Player{})
		err = rows.Scan(&players[i].ID, &players[i].PlayerType, &players[i].SourceID, &players[i].FriendID, &players[i].DisplayOrder)
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
	previousPlayers := make(map[ID]Player, len(players))
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

	queries := make(chan sqlFunction, len(insertPlayers)+len(updatePlayers)+len(previousPlayers))
	quit := make(chan error)
	go exececuteInTransaction(queries, quit)
	for deleteID := range previousPlayers {
		queries <- newSQLFunction("del_player", deleteID)
	}
	for _, insertPlayer := range insertPlayers {
		queries <- newSQLFunction("add_player", insertPlayer.DisplayOrder, insertPlayer.PlayerType, insertPlayer.SourceID, insertPlayer.FriendID, st)
	}
	for _, updateplayer := range updatePlayers {
		queries <- newSQLFunction("set_player", updateplayer.DisplayOrder, updateplayer.ID)
	}
	close(queries)
	return <-quit
}
