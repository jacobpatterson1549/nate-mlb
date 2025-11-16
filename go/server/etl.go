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
		SportTypes() db.SportTypeMap
		PlayerTypes() db.PlayerTypeMap
		GetUtcTime() time.Time
	}
	scoreCategoryInfo struct {
		pt      db.PlayerType
		pti     db.PlayerTypeInfo
		year    int
		friends []db.Friend
		players []db.Player
	}
)

// getEtlStats retrieves, calculates, and caches the player stats
func getEtlStats(st db.SportType, ds etlDatastore, scoreCategorizers map[db.PlayerType]request.ScoreCategorizer) (*EtlStats, error) {
	currentTime := ds.GetUtcTime()
	es := EtlStats{
		etlRefreshTime: previousMidnight(currentTime),
	}
	stat, err := ds.GetStat(st)
	if err != nil {
		return nil, err
	}
	if stat == nil {
		return &es, nil
	}
	es.sportTypeName = ds.SportTypes()[st].Name
	es.sportType = st
	es.year = stat.Year
	if err := updateStat(stat, st, ds, scoreCategorizers, es.etlRefreshTime, currentTime); err != nil {
		return nil, err
	}
	if err := es.setStat(*stat); err != nil {
		return nil, err
	}
	return &es, nil
}

func updateStat(stat *db.Stat, st db.SportType, ds etlDatastore, scoreCategorizers map[db.PlayerType]request.ScoreCategorizer, etlRefreshTime, currentTime time.Time) error {
	if stat.EtlTimestamp == nil || len(stat.EtlJSON) == 0 || stat.EtlTimestamp.Before(etlRefreshTime) {
		scoreCategories, err := getScoreCategories(st, ds, stat.Year, scoreCategorizers)
		if err != nil {
			return err
		}
		etlJSON, err := json.Marshal(scoreCategories)
		if err != nil {
			return fmt.Errorf("converting stats to json for sportType %v, year %v: %w", st, stat.Year, err)
		}
		stat.EtlJSON = string(etlJSON)
		stat.EtlTimestamp = &currentTime
		err = ds.SetStat(*stat)
		if err != nil {
			return err
		}
	}
	return nil
}

func getScoreCategories(st db.SportType, ds etlDatastore, year int, scoreCategorizers map[db.PlayerType]request.ScoreCategorizer) ([]request.ScoreCategory, error) {
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
		sci := scoreCategoryInfo{
			pt:      pt,
			pti:     playerTypes[pt],
			year:    year,
			friends: friends,
			players: playersByType[pt],
		}
		go getScoreCategory(sci, scoreCategorizers[pt], scoreCategoriesCh, quit)
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

func getScoreCategory(sci scoreCategoryInfo, scoreCategorizer request.ScoreCategorizer, scoreCategories chan<- request.ScoreCategory, quit chan<- error) {
	if scoreCategorizer == nil {
		quit <- fmt.Errorf("no ScoreCategorizer for PlayerType %v", sci.pt)
		return
	}
	// providing playerType here is somewhat redundant, but this allows some scoreCategorizers to handle multiple PlayerTypes
	scoreCategory, err := scoreCategorizer.RequestScoreCategory(sci.pt, sci.pti, sci.year, sci.friends, sci.players)
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
	if len(stat.EtlJSON) == 0 {
		return fmt.Errorf("stat has no etlJSON: %v", stat)
	}
	var scoreCategories []request.ScoreCategory
	err := json.Unmarshal([]byte(stat.EtlJSON), &scoreCategories)
	if err != nil {
		return fmt.Errorf("decoding ScoreCategories from Stat etlJSON: %w", err)
	}
	es.etlTime = *stat.EtlTimestamp
	es.scoreCategories = scoreCategories
	return nil
}
