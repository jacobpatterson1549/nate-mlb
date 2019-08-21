package server

import (
	"encoding/json"
	"fmt"
	"nate-mlb/internal/db"
	"nate-mlb/internal/request"
	"sync"
	"time"
)

// EtlStats contains ScoreCategories that were stored at a specific time
type EtlStats struct {
	EtlTime         time.Time
	ScoreCategories []request.ScoreCategory
	// do not serialize these fields:
	etlRefreshTime time.Time
	sportType      db.SportType
	year           int
}

// getEtlStats retrieves, calculates, and caches the player stats
func getEtlStats(st db.SportType) (EtlStats, error) {
	var es EtlStats

	var year int
	stat, err := db.GetEtlStatsJSON(st)
	if err != nil {
		return es, err
	}
	fetchStats := true
	currentTime := db.GetUtcTime()
	es.etlRefreshTime = previousMidnight(currentTime)
	es.sportType = st
	es.year = stat.Year
	if len(stat.EtlStatsJSON) > 0 {
		err = json.Unmarshal([]byte(stat.EtlStatsJSON), &es)
		if err != nil {
			return es, fmt.Errorf("problem converting stats from json for year %v: %v", year, err)
		}
		fetchStats = es.EtlTime.Before(es.etlRefreshTime)
	}
	if fetchStats {
		scoreCategories, err := es.getScoreCategories(st)
		if err != nil {
			return es, err
		}
		es.EtlTime = currentTime
		es.ScoreCategories = scoreCategories
		es.sportType = st
		etlJSON, err := json.Marshal(es)
		if err != nil {
			return es, fmt.Errorf("problem converting stats to json for year %v: %v", year, err)
		}
		err = db.SetEtlStats(st, string(etlJSON))
	}
	return es, err
}

func (es EtlStats) getScoreCategories(st db.SportType) ([]request.ScoreCategory, error) {
	friends, err := db.GetFriends(st)
	if err != nil {
		return nil, err
	}
	playerTypes, err := db.LoadPlayerTypes(st)
	if err != nil {
		return nil, err
	}
	players, err := db.GetPlayers(st)
	if err != nil {
		return nil, err
	}
	activeYear, err := db.GetActiveYear(st)
	if err != nil {
		return nil, err
	}
	fpi := request.FriendPlayerInfo{
		Friends: friends,
		Players: players,
		Year:    activeYear,
	}

	numCategories := len(playerTypes)
	scoreCategories := make([]request.ScoreCategory, numCategories)
	var wg sync.WaitGroup
	wg.Add(numCategories)
	var lastError error
	for i, playerType := range playerTypes {
		go es.getScoreCategory(scoreCategories, i, playerType, fpi, &wg, &lastError)
	}
	wg.Wait()
	return scoreCategories, lastError
}

func (es EtlStats) getScoreCategory(scoreCategories []request.ScoreCategory, index int, playerType db.PlayerType, fpi request.FriendPlayerInfo, wg *sync.WaitGroup, lastError *error) {
	var scoreCategory request.ScoreCategory
	var err error
	scoreCategorizer, ok := request.ScoreCategorizers[playerType]
	if !ok {
		err = fmt.Errorf("problem: no scoreCategorizer for player type %v", playerType)
	} else {
		scoreCategory, err = scoreCategorizer.RequestScoreCategory(fpi, playerType)
	}

	if err != nil {
		*lastError = err // ingore earlier errors - the last one set is shown
	} else {
		scoreCategories[index] = scoreCategory
	}
	wg.Done()
}

// previousMidnight returns the previous midnight relative to Honolulu time (UTC-10)
func previousMidnight(t time.Time) time.Time {
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, t.Location())
	if midnight.After(t) {
		midnight = midnight.Add(-24 * time.Hour)
	}
	return midnight
}
