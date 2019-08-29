package request

import (
	"fmt"
	"nate-mlb/internal/db"
	"strconv"
	"strings"
)

// mlbPlayerRequestor contains invormation about requests for hitter/pitcher names/stats
type mlbPlayerRequestor struct {
	playerType db.PlayerType
}

// MlbPlayerNames is used to unmarshal a request for player names
type MlbPlayerNames struct {
	People []struct {
		ID       int    `json:"id"`
		FullName string `json:"fullName"`
	} `json:"people"`
}

// MlbPlayerStats is used to unmarshal a player homeRuns/wins request
type MlbPlayerStats struct {
	MlbPlayerTypeStats []MlbPlayerTypeStat `json:"stats"`
}

// MlbPlayerTypeStat contains stats for a type of position for a player
type MlbPlayerTypeStat struct {
	Group struct {
		DisplayName string `json:"displayName"`
	} `json:"group"`
	Splits []struct {
		Stat MlbStat `json:"stat"`
	} `json:"splits"`
}

// MlbStat contains a stat for a particular team the player has been on, or is the sum of stats if it is the last one
type MlbStat struct {
	HomeRuns int `json:"homeRuns"`
	Wins     int `json:"wins"`
}

// RequestScoreCategory implements the ScoreCategorizer interface
func (r mlbPlayerRequestor) RequestScoreCategory(fpi FriendPlayerInfo, pt db.PlayerType) (ScoreCategory, error) {
	playerScores, lastError := r.requestPlayerScores(fpi.Players, fpi.Year)
	var scoreCategory ScoreCategory
	if lastError != nil {
		return scoreCategory, lastError
	}
	err := scoreCategory.populate(fpi.Friends, fpi.Players, r.playerType, playerScores, true)
	return scoreCategory, err
}

func (r mlbPlayerRequestor) requestPlayerScores(players []db.Player, year int) (map[int]*PlayerScore, error) {
	playerScores := make(map[int]*PlayerScore)
	for _, player := range players {
		if player.PlayerType == r.playerType {
			playerScores[player.PlayerID] = &PlayerScore{
				PlayerID: player.PlayerID,
			}
		}
	}

	playerNames := make(chan playerName, len(playerScores))
	playerStats := make(chan playerStat, len(playerScores))
	quit := make(chan error)
	go r.requestPlayerNames(playerScores, playerNames, quit)
	go r.requestPlayerStats(year, playerScores, playerStats, quit)

	i := 0
	for {
		select {
		case err := <-quit:
			return playerScores, err
		case playerName := <-playerNames:
			if playerScore, ok := playerScores[playerName.id]; ok {
				playerScore.PlayerName = playerName.name
				i++
			}
		case playerStat := <-playerStats:
			if playerScore, ok := playerScores[playerStat.id]; ok {
				playerScore.Score = playerStat.stat
				i++
			}
		}
		if i == len(playerScores)*2 {
			return playerScores, nil
		}
	}
}

func (r *mlbPlayerRequestor) requestPlayerNames(playerIDs map[int]*PlayerScore, playerNames chan<- playerName, quit chan<- error) {
	playerIDStrings := make([]string, len(playerIDs))
	i := 0
	for playerID := range playerIDs {
		playerIDStrings[i] = strconv.Itoa(playerID)
		i++
	}
	playerNamesURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people?personIds=%s&fields=people,id,fullName", strings.Join(playerIDStrings, ",")), ",", "%2C")
	var mlbPlayerNames MlbPlayerNames
	err := request.structPointerFromURL(playerNamesURL, &mlbPlayerNames)
	if err != nil {
		quit <- err
		return
	}

	i = 0
	for _, person := range mlbPlayerNames.People {
		if _, ok := playerIDs[person.ID]; ok {
			playerNames <- playerName{
				id:   person.ID,
				name: person.FullName,
			}
			i++
		}
	}
	if i < len(playerIDs) {
		quit <- fmt.Errorf("Expected recieve %d player names, but only got %d", len(playerIDs), i)
	}
}

func (r *mlbPlayerRequestor) requestPlayerStats(year int, playerIDs map[int]*PlayerScore, playerStats chan<- playerStat, quit chan<- error) {
	for playerID := range playerIDs {
		go r.getPlayerStat(playerID, year, playerStats, quit)
	}
}

func (r *mlbPlayerRequestor) getPlayerStat(playerID int, year int, playerStats chan<- playerStat, quit chan<- error) {
	stat, err := r.requestPlayerStat(playerID, year)
	if err != nil {
		quit <- err
		return
	}

	playerStats <- playerStat{
		id:   playerID,
		stat: stat,
	}
}

func (r *mlbPlayerRequestor) requestPlayerStat(playerID int, year int) (int, error) { // TODO: make return (playerStat, error)
	mlbPlayerStatsURL := strings.ReplaceAll(fmt.Sprintf("http://statsapi.mlb.com/api/v1/people/%d/stats?&season=%d&stats=season&fields=stats,group,displayName,splits,stat,homeRuns,wins", playerID, year), ",", "%2C")
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
	for _, playerTypeStat := range mps.MlbPlayerTypeStats {
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
