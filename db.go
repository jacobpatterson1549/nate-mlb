package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func getFriendPlayerInfo() FriendPlayerInfo {
	driverName := "postgres"
	datasourceName := os.Getenv("DATABASE_URL")
	db, err := sql.Open(driverName, datasourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return FriendPlayerInfo{
		friends:     getFriends(db),
		playerTypes: getPlayerTypes(db),
		players:     getPlayers(db),
	}
}

// TODO: use shared logic to request friends, playerTypes, players (but with helper mapper functions)
func getFriends(db *sql.DB) []Friend {
	rows, err := db.Query("SELECT id, name FROM friends ORDER BY display_order ASC")
	if err != nil {
		log.Fatalf("Error reading friends: %q", err)
	}
	defer rows.Close()

	friends := []Friend{}
	i := 0
	for rows.Next() {
		friends = append(friends, Friend{})
		err = rows.Scan(&friends[i].id, &friends[i].name)
		if err != nil {
			log.Fatalf("Problem reading data: %q", err)
		}
		i++
	}
	return friends
}

func getPlayerTypes(db *sql.DB) []PlayerType {
	rows, err := db.Query("SELECT id, name FROM player_types")
	if err != nil {
		log.Fatalf("Error reading playerTypes: %q", err)
	}
	defer rows.Close()

	playerTypes := []PlayerType{}
	i := 0
	for rows.Next() {
		playerTypes = append(playerTypes, PlayerType{})
		err = rows.Scan(&playerTypes[i].id, &playerTypes[i].name)
		if err != nil {
			log.Fatalf("Problem reading data: %q", err)
		}
		i++
	}
	return playerTypes
}

func getPlayers(db *sql.DB) []Player {
	rows, err := db.Query("SELECT player_type_id, player_id, friend_id FROM players ORDER BY player_type_id, friend_id, display_order")
	if err != nil {
		log.Fatalf("Error reading playerTypes: %q", err)
	}
	defer rows.Close()

	players := []Player{}
	i := 0
	for rows.Next() {
		players = append(players, Player{})
		err = rows.Scan(&players[i].playerTypeID, &players[i].playerID, &players[i].friendID)
		if err != nil {
			log.Fatalf("Problem reading data: %q", err)
		}
		i++
	}
	return players
}

// FriendPlayerInfo contain all the pool items for each Friend.
type FriendPlayerInfo struct {
	friends     []Friend
	playerTypes []PlayerType
	players     []Player
}

// Friend contains the name of the person in the pool.
type Friend struct {
	id   int
	name string
}

// PlayerType contain a name of a pool item.
type PlayerType struct {
	id   int
	name string
}

// Player maps a player (of a a specific PlayerType) to a Friend.
type Player struct {
	playerTypeID int
	playerID     int
	friendID     int
}
