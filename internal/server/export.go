package server

import (
	"encoding/csv"
	"fmt"
	"io"
	"nate-mlb/internal/request"
)

func exportToCsv(es request.EtlStats, w io.Writer) error {
	records := make([][]string, 1)
	records[0] = []string{"type", "friend", "value", "player", "score"}

	csvWriter := csv.NewWriter(w)
	err := csvWriter.WriteAll(records)
	if err != nil {
		return fmt.Errorf("problem writing to csv: %v", err)
	}
	return nil
}
