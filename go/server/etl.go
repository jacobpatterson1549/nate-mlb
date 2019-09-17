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
	etlTime         time.Time
	scoreCategories []request.ScoreCategory
	etlRefreshTime  time.Time
	sportTypeName   string
	sportType       db.SportType
	year            int
}

// getEtlStats retrieves, calculates, and caches the player stats
func getEtlStats(st db.SportType) (EtlStats, error) {
	currentTime := db.GetUtcTime()
	es := EtlStats{
		etlRefreshTime: previousMidnight(currentTime),
	}
	stat, err := db.GetStat(st)
	if err != nil || stat == nil {
		return es, err
	}
	es.sportTypeName = st.Name()
	es.sportType = st
	es.year = stat.Year
	if stat.EtlTimestamp == nil || stat.EtlJSON == nil || stat.EtlTimestamp.Before(es.etlRefreshTime) {
		scoreCategories, err := getScoreCategories(st, es.year)
		if err != nil {
			return es, err
		}
		etlJSON, err := json.Marshal(scoreCategories)
		if err != nil {
			return es, fmt.Errorf("converting stats to json for sportType %v, year %v: %w", es.sportType, es.year, err)
		}
		stat.EtlJSON = &etlJSON
		stat.EtlTimestamp = &currentTime
		err = db.SetStat(*stat)
		if err != nil {
			return es, err
		}
	}
	err = es.setStat(*stat)
	return es, err
}

func getScoreCategories(st db.SportType, year int) ([]request.ScoreCategory, error) {
	friends, err := db.GetFriends(st)
	if err != nil {
		return nil, err
	}
	players, err := db.GetPlayers(st)
	if err != nil {
		return nil, err
	}
	playerTypes := db.GetPlayerTypes(st)
	fpi := request.NewFriendPlayerInfo(friends, players, playerTypes, year)
	scoreCategoriesCh := make(chan request.ScoreCategory, len(playerTypes))
	quit := make(chan error)
	for _, playerType := range playerTypes {
		go getScoreCategory(playerType, fpi, scoreCategoriesCh, quit)
	}
	scoreCategories := make([]request.ScoreCategory, len(playerTypes))
	finishedScoreCategories := 0
	for {
		select {
		case err = <-quit:
			return nil, err
		case scoreCategory := <-scoreCategoriesCh:
			scoreCategories[finishedScoreCategories] = scoreCategory
			finishedScoreCategories++
		}
		if finishedScoreCategories == len(playerTypes) {
			sort.Slice(scoreCategories, func(i, j int) bool {
				return scoreCategories[i].PlayerType.DisplayOrder() < scoreCategories[j].PlayerType.DisplayOrder()
			})
			return scoreCategories, nil
		}
	}
}

func getScoreCategory(playerType db.PlayerType, fpi request.FriendPlayerInfo, scoreCategories chan<- request.ScoreCategory, quit chan<- error) {
	scoreCategorizer, ok := request.ScoreCategorizers[playerType]
	if !ok {
		quit <- fmt.Errorf("no scoreCategorizer for player type %v", playerType)
		return
	}
	// providing playerType here is somewhat redundant, but this allows some scoreCategorizers to handle multiple PlayerTypes
	scoreCategory, err := scoreCategorizer.RequestScoreCategory(fpi, playerType)
	if err != nil {
		quit <- err
		return
	}
	scoreCategories <- scoreCategory
}

// previousMidnight returns the previous midnight relative to Honolulu time (UTC-10)
func previousMidnight(t time.Time) time.Time {
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, time.UTC)
	if midnight.After(t) {
		midnight = midnight.Add(-24 * time.Hour)
	}
	return midnight.In(t.Location())
}

// setStat sets the etlTime and scoreCategories (etlJson) from the Stat
func (es *EtlStats) setStat(stat db.Stat) error {
	if stat.EtlJSON == nil {
		return fmt.Errorf("stat has no etlJSON: %v", stat)
	}
	var scoreCategories []request.ScoreCategory
	err := json.Unmarshal([]byte(*stat.EtlJSON), &scoreCategories)
	if err != nil {
		return fmt.Errorf("unmarshalling ScoreCategories from Stat etlJSON: %w", err)
	}
	es.etlTime = *stat.EtlTimestamp
	es.scoreCategories = scoreCategories
	return nil
}
