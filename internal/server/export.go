package server

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/jacobpatterson1549/nate-mlb/internal/db"
	"github.com/jacobpatterson1549/nate-mlb/internal/request"
)

func exportToCsv(st db.SportType, es EtlStats, w io.Writer) error {
	records := createCsvRecords(es)
	csvWriter := csv.NewWriter(w)
	err := csvWriter.WriteAll(records)
	if err != nil {
		return fmt.Errorf("problem writing to csv: %v", err)
	}
	return nil
}

func createCsvRecords(es EtlStats) [][]string {
	records := make([][]string, 3)
	title := fmt.Sprintf("%d %s scores", es.year, es.sportTypeName)
	records[0] = []string{"nate-mlb", title}
	records[2] = []string{"type", "friend", "value", "player", "score"}
	for i, sc := range es.ScoreCategories {
		if i != 0 {
			records = append(records, nil)
		}
		for j, fs := range sc.FriendScores {
			records = append(records, nil)
			for k, ps := range fs.PlayerScores {
				record := createCsvRecord(sc, fs, ps, j, k)
				records = append(records, record)
			}
		}
	}
	return records
}

func createCsvRecord(sc request.ScoreCategory, fs request.FriendScore, ps request.PlayerScore, fsIndex, psIndex int) []string {
	record := make([]string, 5)
	if psIndex == 0 {
		if fsIndex == 0 {
			record[0] = sc.Name
		}
		record[1] = fs.Name
		record[2] = strconv.Itoa(fs.Score)
	}
	record[3] = ps.Name
	record[4] = strconv.Itoa(ps.Score)
	return record
}
