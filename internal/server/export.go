package server

import (
	"encoding/csv"
	"io"
	"nate-mlb/internal/request"
)

func exportToCsv(es request.EtlStats, w io.Writer) error {
	records := make([][]string, 1)
	records[0] = []string{"type", "friend", "value", "player", "score"}

	csvWriter := csv.NewWriter(w)
	return csvWriter.WriteAll(records)
}
