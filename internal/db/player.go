package db

import (
	"fmt"
)

// Player maps a player (of a a specific PlayerType) to a Friend.
type Player struct {
	ID           int
	DisplayOrder int
	PlayerType   PlayerType
	SourceID     SourceID
	FriendID     int
}

// SourceID is the id used to retrieve information about the player from external sources
type SourceID int

// GetPlayers gets the players for the active year
// TODO: return map[PlayerType]Player -> this rwill reduce filtering later on -> maybe use map[PlayerType]map[int]Player (int=friendId)
func GetPlayers(st SportType) ([]Player, error) {
	rows, err := db.Query(
		`SELECT p.id, p.display_order, p.player_type_id, p.source_id, p.friend_id
		FROM stats AS s
		JOIN friends AS f ON s.year = f.year AND s.sport_type_id = f.sport_type_id
		JOIN players AS p ON f.id = p.friend_id
		WHERE s.sport_type_id = $1
		AND s.active`,
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
		err = rows.Scan(&players[i].ID, &players[i].DisplayOrder, &players[i].PlayerType, &players[i].SourceID, &players[i].FriendID)
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

	queries := make(chan query, len(insertPlayers)+len(updatePlayers)+len(previousPlayers))
	quit := make(chan error)
	go exececuteInTransaction(queries, quit)
	for deleteID := range previousPlayers {
		queries <- newQuery(
			`DELETE FROM players
			WHERE id = $1`,
			deleteID,
		)
	}
	for _, insertPlayer := range insertPlayers {
		queries <- newQuery(
			`INSERT INTO players
			(display_order, player_type_id, source_id, friend_id)
			SELECT $1, $2, $3, $4
			FROM stats
			WHERE sport_type_id = $5
			AND active`,
			insertPlayer.DisplayOrder,
			insertPlayer.PlayerType,
			insertPlayer.SourceID,
			insertPlayer.FriendID,
			st,
		)
	}
	for _, updateplayer := range updatePlayers {
		queries <- newQuery(
			`UPDATE players
			SET display_order = $1
			WHERE id = $2`,
			updateplayer.DisplayOrder,
			updateplayer.ID,
		)
	}
	close(queries)
	return <-quit
}
