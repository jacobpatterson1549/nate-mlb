package server

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

// EtlStats contains ScoreCategories that were stored at a specific time
type EtlStats struct {
	EtlTime         time.Time
	ScoreCategories []request.ScoreCategory
	// do not serialize these fields:
	etlRefreshTime time.Time
	sportTypeName  string
	sportType      db.SportType
	year           int
}

// getEtlStats retrieves, calculates, and caches the player stats
func getEtlStats(st db.SportType) (EtlStats, error) {
	var es EtlStats

	stat, err := db.GetStat(st)
	if err != nil {
		return es, err
	}
	currentTime := db.GetUtcTime()
	es.etlRefreshTime = previousMidnight(currentTime)
	es.sportTypeName = st.Name()
	es.sportType = st
	es.year = stat.Year
	fetchStats := true
	if stat.EtlTimestamp != nil && stat.EtlJSON != nil {
		fetchStats = stat.EtlTimestamp.Before(es.etlRefreshTime)
	}
	if fetchStats {
		es.ScoreCategories, err = es.getScoreCategories(st)
		if err != nil {
			return es, err
		}
		es.EtlTime = currentTime
		stat, err = es.toStat()
		if err != nil {
			return es, err
		}
		err = db.SetStat(stat)
	} else {
		es.EtlTime = *stat.EtlTimestamp
		err = es.setScoreCategories(stat)
	}
	return es, err
}

func (es EtlStats) getScoreCategories(st db.SportType) ([]request.ScoreCategory, error) {
	friends, err := db.GetFriends(st)
	if err != nil {
		return nil, err
	}
	players, err := db.GetPlayers(st)
	if err != nil {
		return nil, err
	}
	playerTypes := db.GetPlayerTypes(st)
	fpi := request.NewFriendPlayerInfo(friends, players, playerTypes, es.year)

	numScoreCategories := len(playerTypes)
	scoreCategoriesCh := make(chan request.ScoreCategory, len(playerTypes))
	quit := make(chan error)
	for _, playerType := range playerTypes {
		go es.getScoreCategory(playerType, fpi, scoreCategoriesCh, quit)
	}
	scoreCategories := []request.ScoreCategory{}
	finishedScoreCategories := 0
	for {
		select {
		case err = <-quit:
			return nil, err
		case scoreCategory := <-scoreCategoriesCh:
			scoreCategories = append(scoreCategories, scoreCategory)
			finishedScoreCategories++
		}
		if finishedScoreCategories == numScoreCategories {
			sort.Slice(scoreCategories, func(i, j int) bool {
				return scoreCategories[i].PlayerType.DisplayOrder() < scoreCategories[j].PlayerType.DisplayOrder()
			})
			return scoreCategories, nil
		}
	}
}

func (es EtlStats) getScoreCategory(playerType db.PlayerType, fpi request.FriendPlayerInfo, scoreCategories chan<- request.ScoreCategory, quit chan<- error) {
	if scoreCategorizer, ok := request.ScoreCategorizers[playerType]; ok {
		scoreCategory, err := scoreCategorizer.RequestScoreCategory(fpi, playerType)
		if err != nil {
			quit <- err
			return
		}
		scoreCategories <- scoreCategory
	} else {
		quit <- fmt.Errorf("problem: no scoreCategorizer for player type %v", playerType)
	}
}

// previousMidnight returns the previous midnight relative to Honolulu time (UTC-10)
func previousMidnight(t time.Time) time.Time {
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, time.UTC)
	if midnight.After(t) {
		midnight = midnight.Add(-24 * time.Hour)
	}
	return midnight.In(t.Location())
}

func (es EtlStats) toStat() (db.Stat, error) {
	var stat db.Stat
	etlJSON, err := json.Marshal(es.ScoreCategories)
	if err != nil {
		return stat, fmt.Errorf("problem converting stats to json for sportType %v, year %v: %v", es.sportType, es.year, err)
	}
	etlJSONText := string(etlJSON)

	stat.SportType = es.sportType
	stat.Year = es.year
	stat.EtlTimestamp = &es.EtlTime
	stat.EtlJSON = &etlJSONText
	return stat, nil
}

func (es *EtlStats) setScoreCategories(stat db.Stat) error {
	if stat.EtlJSON == nil {
		return fmt.Errorf("Stat has no etlJSON: %v", stat)
	}
	var scoreCategories []request.ScoreCategory
	err := json.Unmarshal([]byte(*stat.EtlJSON), &scoreCategories)
	if err != nil {
		return fmt.Errorf("problem getting ScoreCategories from Stat etlJSON: %v", err)
	}
	es.ScoreCategories = scoreCategories
	return nil
}