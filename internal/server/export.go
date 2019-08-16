package server

import (
	"encoding/csv"
	"fmt"
	"io"
	"nate-mlb/internal/db"
	"nate-mlb/internal/request"
	"strconv"
)

func exportToCsv(es request.EtlStats, w io.Writer) error {
	records := make([][]string, 3)
	records[0] = []string{"nate-mlb", "2019"}
	records[2] = []string{"type", "friend", "value", "player", "score"}
	for i, sc := range es.ScoreCategories {
		for _, fs := range sc.FriendScores {
			records = append(records, nil)
			for k, ps := range fs.PlayerScores {
				record := make([]string, 5)
				if i == 0 {
					record[0] = db.PlayerType(sc.PlayerTypeID).Name()
				}
				if k == 0 {
					record[1] = fs.FriendName
					record[2] = strconv.Itoa(fs.Score)
				}
				record[3] = ps.PlayerName
				record[4] = strconv.Itoa(ps.Score)
				records = append(records, record)
			}
		}
		if i < len(es.ScoreCategories)-1 {
			records = append(records, nil)
		}
	}

	csvWriter := csv.NewWriter(w)
	err := csvWriter.WriteAll(records)
	if err != nil {
		return fmt.Errorf("problem writing to csv: %v", err)
	}
	return nil
}
