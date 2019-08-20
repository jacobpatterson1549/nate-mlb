package main

import (
	"encoding/json"
	"fmt"
	"log"
	"nate-mlb/internal/db"
	"nate-mlb/internal/server"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func main() {
	data := `{"lastUpdated":"2019-08-20 03:01:34","players":[{"id":"100029","esbid":null,"gsisPlayerId":false,"firstName":"San Francisco","lastName":"49ers","teamAbbr":"SF","opponentTeamAbbr":"@MIN","position":"DEF","percentOwned":false,"percentOwnedChange":false,"percentStarted":false,"percentStartedChange":false,"depthChartOrder":null,"numAdds":78,"numDrops":49}]}`
	var playerList NflPlayerList
	errJ := json.Unmarshal([]byte(data), &playerList)
	if errJ != nil {
		panic(errJ)
	}
	fmt.Println("playerList:", playerList)

	driverName := "postgres"
	dataSourceName, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatal("PORT environment variable not set")
	}

	err := db.InitDB(driverName, dataSourceName)
	if err != nil {
		log.Fatal("Could not set database ", err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("PORT (%v) invalid as number: %v", port, err)
	}

	err = server.Run(portNumber)
	if err != nil {
		log.Fatal(err)
	}
}

// NflPlayerList contains information about all the players for a particular year
type NflPlayerList struct {
	Date    string          `json:"lastUpdated"`
	Players []NflPlayerInfo `json:"players"`
}

// NflPlayerInfo is information about a player for a year
type NflPlayerInfo struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Team      string `json:"teamAbbr"`
}
