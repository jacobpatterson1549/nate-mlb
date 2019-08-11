package db

import (
	"database/sql"
	"fmt"
)

// Player maps a player (of a a specific PlayerType) to a Friend.
type Player struct {
	ID           int
	DisplayOrder int
	PlayerTypeID int
	PlayerID     int
	FriendID     int
}

func getPlayers() ([]Player, error) {
	rows, err := db.Query("SELECT p.id, p.display_order, p.player_type_id, p.player_id, p.friend_id FROM players AS p JOIN stats AS s ON p.year = s.year WHERE s.active ORDER BY p.player_type_id, p.friend_id, p.display_order")
	if err != nil {
		return nil, fmt.Errorf("problem reading players: %v", err)
	}
	defer rows.Close()

	players := []Player{}
	i := 0
	for rows.Next() {
		players = append(players, Player{})
		err = rows.Scan(&players[i].ID, &players[i].DisplayOrder, &players[i].PlayerTypeID, &players[i].PlayerID, &players[i].FriendID)
		if err != nil {
			return nil, fmt.Errorf("problem reading data: %v", err)
		}
		i++
	}
	return players, nil
}

// SetPlayers saves the specied players in for the active year. TODO: rename to SavePlayers
func SetPlayers(futurePlayers []Player) error {
	players, err := getPlayers()
	if err != nil {
		return err
	}
	previousPlayers := make(map[int]Player)
	for _, player := range players {
		previousPlayers[player.ID] = player
	}

	insertPlayers := []Player{}
	updatePlayers := []Player{}
	for _, player := range futurePlayers {
		previousPlayer, ok := previousPlayers[player.ID]
		if !ok {
			insertPlayers = append(insertPlayers, player)
		} else if player.DisplayOrder != previousPlayer.DisplayOrder { // can only update display order
			updatePlayers = append(updatePlayers, player)
		}
		delete(previousPlayers, player.ID)
	}
	deletePlayers := previousPlayers

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("problem starting transaction: %v", err)
	}
	var result sql.Result
	for _, player := range insertPlayers {
		if err == nil {
			result, err = tx.Exec(
				"INSERT INTO players (display_order, player_type_id, player_id, friend_id, year) SELECT $1, $2, $3, $4, year FROM stats AS s WHERE s.active",
				player.DisplayOrder,
				player.PlayerTypeID,
				player.PlayerID,
				player.FriendID)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for _, player := range updatePlayers {
		if err == nil {
			result, err = tx.Exec(
				"UPDATE players SET display_order = $1 WHERE id = $2",
				player.DisplayOrder,
				player.ID)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	for playerID := range deletePlayers {
		if err == nil {
			result, err = tx.Exec(
				"DELETE FROM players WHERE id = $1",
				playerID)
			if err == nil {
				err = expectSingleRowAffected(result)
			}
		}
	}
	if err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			err = fmt.Errorf("problem: %v, ROLLBACK ERROR: %v", err.Error(), err2.Error())
		}
	} else {
		err = tx.Commit()
	}
	if err != nil {
		return fmt.Errorf("problem saving players: %v", err)
	}
	return nil
}
