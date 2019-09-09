package request

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

// mlbPlayerRequestor contains information about requests for hitter/pitcher names/stats
type mlbPlayerRequestor struct {
	playerType db.PlayerType
}

// MlbPlayerNames is used to unmarshal a request for player names
type MlbPlayerNames struct {
	People []MlbPlayerName `json:"people"`
}

// MlbPlayerName contains a player's source ID and full name
type MlbPlayerName struct {
	ID       db.SourceID `json:"id"`
	FullName string      `json:"fullName"`
}

// MlbPlayerStats is used to unmarshal a player homeRuns/wins request
type MlbPlayerStats struct {
	Stats []MlbPlayerStat `json:"stats"`
}

// MlbPlayerStat contains stats for a type of position for a player
type MlbPlayerStat struct {
	Group  MlbPlayerStatGroup   `json:"group"`
	Splits []MlbPlayerStatSplit `json:"splits"`
}

// MlbPlayerStatGroup contains the type of stat a MlbPlayerStat is for
type MlbPlayerStatGroup struct {
	DisplayName string `json:"displayName"`
}

// MlbPlayerStatSplit contains stats for a single team or is a total of others
type MlbPlayerStatSplit struct {
	Stat MlbStat `json:"stat"`
}

// MlbStat contains a stat for a particular team the player has been on, or is the sum of stats if it is the last one
type MlbStat struct {
	HomeRuns int `json:"homeRuns"`
	Wins     int `json:"wins"`
}

// RequestScoreCategory implements the ScoreCategorizer interface
func (r mlbPlayerRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	sourceIDs := make(map[db.SourceID]bool, len(fpi.Players[pt]))
	for _, player := range fpi.Players[pt] {
		sourceIDs[player.SourceID] = true
	}

	playerNames := make(map[db.SourceID]string, len(sourceIDs))
	playerStats := make(map[db.SourceID]int, len(sourceIDs))
	playerNamesCh := make(chan playerName, len(sourceIDs))
	playerStatsCh := make(chan playerStat, len(sourceIDs))
	quit := make(chan error)

	var scoreCategory ScoreCategory
	if len(sourceIDs) > 0 {
		go r.requestPlayerNames(sourceIDs, playerNamesCh, quit)
		go r.requestPlayerStats(fpi.Year, sourceIDs, playerStatsCh, quit)
		i := 0
		for {
			select {
			case err := <-quit:
				return scoreCategory, err
			case playerName := <-playerNamesCh:
				playerNames[playerName.sourceID] = playerName.name
			case playerStat := <-playerStatsCh:
				playerStats[playerStat.sourceID] = playerStat.stat
			}
			i++
			if i == len(sourceIDs)*2 {
				break
			}
		}
	}
	playerNameScores, err := playerNameScoresFromFieldMaps(fpi.Players[pt], playerNames, playerStats)
	if err != nil {
		return scoreCategory, err
	}
	return newScoreCategory(fpi, pt, playerNameScores, true), nil
}

func (r mlbPlayerRequestor) requestPlayerNames(sourceIDs map[db.SourceID]bool, playerNames chan<- playerName, quit chan<- error) {
	sourceIDStrings := make([]string, len(sourceIDs))
	i := 0
	for sourceID := range sourceIDs {
		sourceIDStrings[i] = strconv.Itoa(int(sourceID))
		i++
	}
	playerNamesURL := strings.ReplaceAll(
		fmt.Sprintf(
			"http://statsapi.mlb.com/api/v1/people?personIds=%s&fields=people,id,fullName",
			strings.Join(sourceIDStrings, ",")),
		",",
		"%2C")
	var mlbPlayerNames MlbPlayerNames
	err := request.structPointerFromURL(playerNamesURL, &mlbPlayerNames)
	if err != nil {
		quit <- err
		return
	}

	i = 0
	for _, person := range mlbPlayerNames.People {
		if _, ok := sourceIDs[person.ID]; ok {
			playerNames <- playerName{
				sourceID: person.ID,
				name:     person.FullName,
			}
			i++
		}
	}
	if i != len(sourceIDs) {
		// this might not be read if names for the same sourceID occur multiple times and this request takes longer than the stat requests
		quit <- fmt.Errorf("problem: expected to receive %d mlb player names, but only got %d", len(sourceIDs), i)
	}
}

func (r mlbPlayerRequestor) requestPlayerStats(year int, sourceIDs map[db.SourceID]bool, playerStats chan<- playerStat, quit chan<- error) {
	for sourceID := range sourceIDs {
		go r.getPlayerStat(sourceID, year, playerStats, quit)
	}
}

func (r mlbPlayerRequestor) getPlayerStat(sourceID db.SourceID, year int, playerStats chan<- playerStat, quit chan<- error) {
	stat, err := r.requestPlayerStat(sourceID, year)
	if err != nil {
		quit <- err
		return
	}
	playerStats <- playerStat{
		sourceID: sourceID,
		stat:     stat,
	}
}

func (r mlbPlayerRequestor) requestPlayerStat(sourceID db.SourceID, year int) (int, error) {
	mlbPlayerStatsURL := strings.ReplaceAll(
		fmt.Sprintf(
			"http://statsapi.mlb.com/api/v1/people/%d/stats?&season=%d&stats=season&fields=stats,group,displayName,splits,stat,homeRuns,wins",
			sourceID,
			year),
		",",
		"%2C")
	var mlbPlayerStats MlbPlayerStats
	err := request.structPointerFromURL(mlbPlayerStatsURL, &mlbPlayerStats)
	if err != nil {
		return -1, err
	}
	return mlbPlayerStats.getStat(r.playerType)
}

func (mps MlbPlayerStats) getStat(playerType db.PlayerType) (int, error) {
	switch playerType {
	case db.PlayerTypeHitter:
		return mps.lastStat("hitting", MlbStat.getHomeRuns), nil
	case db.PlayerTypePitcher:
		return mps.lastStat("pitching", MlbStat.getWins), nil
	default:
		return -1, fmt.Errorf("Cannot get stat of playerType %v for player", playerType)
	}
}

func (mps MlbPlayerStats) lastStat(groupDisplayName string, stat func(MlbStat) int) int {
	for _, playerTypeStat := range mps.Stats {
		if groupDisplayName == playerTypeStat.Group.DisplayName {
			splits := playerTypeStat.Splits
			if len(splits) > 0 {
				lastStat := splits[len(splits)-1].Stat
				return stat(lastStat)
			}
		}
	}
	return 0
}

func (ms MlbStat) getHomeRuns() int {
	return ms.HomeRuns
}

func (ms MlbStat) getWins() int {
	return ms.Wins
}
