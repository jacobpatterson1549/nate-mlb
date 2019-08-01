package main

import (
	"encoding/json"
	"time"
)

func getETLStats() (EtlStats, error) {
	// TODO: this thrashes the db for getting connections.  Shuld use connection pool

	es := EtlStats{}
	etlStatsJSON, err := getKeyStoreValue("etl")
	if err != nil {
		return es, err
	}
	err = json.Unmarshal([]byte(etlStatsJSON), &es)
	if err != nil {
		return es, err
	}

	currentTime := time.Now()
	if es.isStale(currentTime) {
		scoreCategories, err := getStats()
		if err != nil {
			return es, err
		}
		es.Stats = scoreCategories
		es.EtlTime = currentTime
		err = es.save()
		if err != nil {
			return es, err
		}
	}
	return es, nil
}

func (es *EtlStats) save() error {
	etlStatsJSON, err := json.Marshal(*es)
	if err != nil {
		return err
	}

	return setKeyStoreValue("etl", string(etlStatsJSON))
}

func (es *EtlStats) isStale(currentTime time.Time) bool {
	previousMidnight := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())

	return es.EtlTime.Before(previousMidnight)
}

// EtlStats contain some score categories that were stored at a specific time
type EtlStats struct {
	Stats   []ScoreCategory `json:"stats"`
	EtlTime time.Time       `json:"etlTime"`
}
