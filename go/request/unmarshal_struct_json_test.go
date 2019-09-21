package request

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
)

func TestUnmarshalStructJson(t *testing.T) {
	unmarshalStructJSONTests := []struct {
		interfaceJSON string
		got           interface{}
		want          interface{}
	}{
		{
			interfaceJSON: `{"ref":"f370c06649a740542504b7ecb18031908f394fe8","updated_at":"2019-08-29T17:29:35Z"}`,
			got:           new(GithubRepoDeployment),
			want: &GithubRepoDeployment{
				Version: "f370c06649a740542504b7ecb18031908f394fe8",
				Time:    time.Date(2019, time.August, 29, 17, 29, 35, 0, time.UTC),
			},
		},
		{
			interfaceJSON: `{"stats":[{"group":{"displayName":"hitting"},"splits":[{"stat":{"homeRuns":43}}]}]}`,
			got:           new(MlbPlayerStats),
			want: &MlbPlayerStats{
				Stats: []MlbPlayerStat{
					{
						Group: MlbPlayerStatGroup{
							DisplayName: "hitting",
						},
						Splits: []MlbPlayerStatSplit{
							{
								Stat: MlbStat{
									HomeRuns: 43,
								},
							},
						},
					},
				},
			},
		},
		{
			interfaceJSON: `{"people":[{"id":502110,"fullName":"J.D. Martinez"},{"id":429665,"fullName":"Edwin Encarnacion"}]}`,
			got:           new(MlbPlayerNames),
			want: &MlbPlayerNames{
				People: []MlbPlayerName{
					{ID: 502110, FullName: "J.D. Martinez"},
					{ID: 429665, FullName: "Edwin Encarnacion"},
				},
			},
		},
		{
			interfaceJSON: `{"search_player_all":{"queryResults":{"totalSize":"314"}}}`,
			got:           new(MlbPlayerSearch),
			want: &MlbPlayerSearch{
				MlbPlayerSearchAll{
					QueryResults: MlbPlayerSearchQueryResults{
						TotalSize: 314,
					},
				},
			},
		},
		{
			interfaceJSON: `{"records":[{"teamRecords":[{"team":{"id":136,"name":"Seattle Mariners"},"wins":116,"losses":46}]}]}`,
			got:           new(MlbTeams),
			want: &MlbTeams{
				Records: []MlbTeamRecords{
					{
						TeamRecords: []MlbTeamRecord{
							{
								Team: MlbTeamRecordName{
									Name: "Seattle Mariners",
									ID:   136,
								},
								Wins:   116,
								Losses: 46,
							},
						},
					},
				},
			},
		},
		{
			interfaceJSON: `{"players":[{"id":"2532975","name":"Russell Wilson","position":"QB","teamAbbr":"SEA","stats":{"6":"35"}}]}`,
			got:           new(NflPlayerList),
			want: &NflPlayerList{
				Players: []NflPlayer{
					{
						ID:       2532975,
						Name:     "Russell Wilson",
						Position: "QB",
						Team:     "SEA",
						Stats: NflPlayerStats{
							PassingTD: 35,
						},
					},
				},
			},
		},
		{
			interfaceJSON: `{"nflTeams":{"20":{"fullName":"Minnesota Vikings","Record":"8-7-1"}}}`,
			got:           new(NflTeamsSchedule),
			want: &NflTeamsSchedule{
				Teams: map[db.SourceID]NflTeam{
					20: {
						Name:   "Minnesota Vikings",
						Record: "8-7-1",
					},
				},
			},
		},
	}
	for i, test := range unmarshalStructJSONTests {
		err1 := json.Unmarshal([]byte(test.interfaceJSON), &test.got)
		gotJSON, err2 := json.Marshal(test.got)
		wantJSON, err3 := json.Marshal(test.want)
		switch {
		case err1 != nil,
			err2 != nil,
			err3 != nil:
			t.Errorf("Test %v:\n%v/%v/%v", i, err1, err2, err3)
		case string(wantJSON) != string(gotJSON):
			t.Errorf("Test %v:\nwanted   %+v\nbut got  %+v\nfor json %v", i, test.want, test.got, test.interfaceJSON)
		}
	}
}
