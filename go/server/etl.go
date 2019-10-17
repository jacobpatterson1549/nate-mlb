package server

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

type (
	// EtlStats contains ScoreCategories that were stored at a specific time
	EtlStats struct {
		etlTime         time.Time
		scoreCategories []request.ScoreCategory
		etlRefreshTime  time.Time
		sportTypeName   string
		sportType       db.SportType
		year            int
	}
	etlDatastore interface {
		GetStat(st db.SportType) (*db.Stat, error)
		GetFriends(st db.SportType) ([]db.Friend, error)
		GetPlayers(st db.SportType) ([]db.Player, error)
		SetStat(stat db.Stat) error
		db.SportTypeGetter
		db.PlayerTypeGetter
		timeGetter
	}
)

// getEtlStats retrieves, calculates, and caches the player stats
func getEtlStats(st db.SportType, ds etlDatastore) (EtlStats, error) {
	// TODO: use interface for ds/sportTypes/playerTypes
	currentTime := ds.GetUtcTime()
	es := EtlStats{
		etlRefreshTime: previousMidnight(currentTime),
	}
	stat, err := ds.GetStat(st)
	if err != nil || stat == nil {
		return es, err
	}
	es.sportTypeName = ds.SportTypes()[st].Name
	es.sportType = st
	es.year = stat.Year
	if stat.EtlTimestamp == nil || stat.EtlJSON == nil || stat.EtlTimestamp.Before(es.etlRefreshTime) {
		scoreCategories, err := getScoreCategories(st, ds, es.year)
		if err != nil {
			return es, err
		}
		etlJSON, err := json.Marshal(scoreCategories)
		if err != nil {
			return es, fmt.Errorf("converting stats to json for sportType %v, year %v: %w", es.sportType, es.year, err)
		}
		stat.EtlJSON = &etlJSON
		stat.EtlTimestamp = &currentTime
		err = ds.SetStat(*stat)
		if err != nil {
			return es, err
		}
	}
	err = es.setStat(*stat)
	return es, err
}

func getScoreCategories(st db.SportType, ds etlDatastore, year int) ([]request.ScoreCategory, error) {
	friends, err := ds.GetFriends(st)
	if err != nil {
		return nil, err
	}
	players, err := ds.GetPlayers(st)
	if err != nil {
		return nil, err
	}
	playerTypes := ds.PlayerTypes()
	stPlayerTypes := getPlayerTypes(st, playerTypes)
	playersByType := make(map[db.PlayerType][]db.Player)
	for _, player := range players {
		playersByType[player.PlayerType] = append(playersByType[player.PlayerType], player)
	}
	scoreCategoriesCh := make(chan request.ScoreCategory, len(stPlayerTypes))
	quit := make(chan error)
	for _, pt := range stPlayerTypes {
		go getScoreCategory(pt, playerTypes[pt], year, friends, playersByType[pt], scoreCategoriesCh, quit)
	}
	scoreCategories := make([]request.ScoreCategory, len(stPlayerTypes))
	finishedScoreCategories := 0
	for {
		select {
		case err = <-quit:
			return nil, err
		case scoreCategory := <-scoreCategoriesCh:
			scoreCategories[finishedScoreCategories] = scoreCategory
			finishedScoreCategories++
		}
		if finishedScoreCategories == len(stPlayerTypes) {
			displayOrder := func(i int) int { return playerTypes[scoreCategories[i].PlayerType].DisplayOrder }
			sort.Slice(scoreCategories, func(i, j int) bool {
				return displayOrder(i) < displayOrder(j)
			})
			return scoreCategories, nil
		}
	}
}

func getPlayerTypes(st db.SportType, playerTypes db.PlayerTypeMap) []db.PlayerType {
	playerTypesList := make([]db.PlayerType, 0, len(playerTypes))
	for pt, ptInfo := range playerTypes {
		if ptInfo.SportType == st {
			playerTypesList = append(playerTypesList, pt)
		}
	}
	displayOrder := func(i int) int { return playerTypes[playerTypesList[i]].DisplayOrder }
	sort.Slice(playerTypesList, func(i, j int) bool {
		return displayOrder(i) < displayOrder(j)
	})
	return playerTypesList
}

func getScoreCategory(pt db.PlayerType, pti db.PlayerTypeInfo, year int, friends []db.Friend, players []db.Player, scoreCategories chan<- request.ScoreCategory, quit chan<- error) {
	// providing playerType here is somewhat redundant, but this allows some scoreCategorizers to handle multiple PlayerTypes
	scoreCategory, err := request.Score(pt, pti, year, friends, players)
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
