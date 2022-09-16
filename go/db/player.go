package db

import (
	"fmt"
)

type (
	// Player maps a player (of a a specific PlayerType) to a Friend.
	Player struct {
		ID           ID
		PlayerType   PlayerType
		SourceID     SourceID
		FriendID     ID
		DisplayOrder int
	}

	// SourceID is the id used to retrieve information about the player from external sources
	SourceID int
)

// GetPlayers gets the players for the active year for a SportType
func (ds Datastore) GetPlayers(st SportType) ([]Player, error) {
	sqlFunction := newReadSQLFunction("get_players", []string{"id", "player_type_id", "source_id", "friend_id", "display_order"}, st)
	rows, err := ds.db.Query(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return nil, fmt.Errorf("reading players: %w", err)
	}
	defer rows.Close()

	var players []Player
	i := 0
	for rows.Next() {
		players = append(players, Player{})
		err = rows.Scan(&players[i].ID, &players[i].PlayerType, &players[i].SourceID, &players[i].FriendID, &players[i].DisplayOrder)
		if err != nil {
			return nil, fmt.Errorf("reading player: %w", err)
		}
		i++
	}
	return players, nil
}

// SavePlayers saves the specified players for the active year for a SportType
func (ds Datastore) SavePlayers(st SportType, futurePlayers []Player) error {
	players, err := ds.GetPlayers(st)
	if err != nil {
		return err
	}
	previousPlayers := make(map[ID]Player, len(players))
	for _, player := range players {
		previousPlayers[player.ID] = player
	}

	insertPlayers := make([]Player, 0, len(futurePlayers))
	updatePlayers := make([]Player, 0, len(futurePlayers))
	for _, player := range futurePlayers {
		previousPlayer, ok := previousPlayers[player.ID]
		switch {
		case !ok:
			ptInfo := ds.playerTypes[player.PlayerType]
			if ptInfo.SportType != st {
				return fmt.Errorf("cannot add Player with PlayerType of %v when saving Players of SportType %v: it has a SportType of %v", player.PlayerType, st, ptInfo.SportType)
			}
			insertPlayers = append(insertPlayers, player)
		case player.DisplayOrder != previousPlayer.DisplayOrder: // can only update display order
			updatePlayers = append(updatePlayers, player)
		}
		delete(previousPlayers, player.ID)
	}

	queries := make([]writeSQLFunction, 0, len(insertPlayers)+len(updatePlayers)+len(previousPlayers))
	for deleteID := range previousPlayers {
		queries = append(queries, newWriteSQLFunction("del_player", deleteID))
	}
	for _, insertPlayer := range insertPlayers {
		queries = append(queries, newWriteSQLFunction("add_player", insertPlayer.DisplayOrder, insertPlayer.PlayerType, insertPlayer.SourceID, insertPlayer.FriendID))
	}
	for _, updatePlayer := range updatePlayers {
		queries = append(queries, newWriteSQLFunction("set_player", updatePlayer.DisplayOrder, updatePlayer.ID))
	}
	return ds.executeInTransaction(queries)
}
