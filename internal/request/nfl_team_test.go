package request

import "testing"

type nflTeamWinsTest struct {
	nflTeam   NflTeam
	want      int
	wantError bool
}

var nflTeamWinsTests = []nflTeamWinsTest{
	{
		nflTeam: NflTeam{Record: "7-9-0"},
		want:    7,
	},
	{
		nflTeam: NflTeam{Record: "6-9-1"},
		want:    6,
	},
	{
		nflTeam: NflTeam{Record: "16"},
		want:    16,
	},
	{
		nflTeam:   NflTeam{Record: ""},
		wantError: true,
	},
	{
		nflTeam:   NflTeam{Record: "eight-8-0"},
		wantError: true,
	},
	{
		nflTeam:   NflTeam{Record: "-4-12-0"},
		wantError: true,
	},
}

func TestNflTeamWins(t *testing.T) {
	for i, test := range nflTeamWinsTests {
		got, err := test.nflTeam.wins()
		if test.wantError {
			if err == nil {
				t.Errorf("Test %v: wanted error", i)
			}
		} else if err != nil {
			t.Errorf("Test %v: %v", i, err)
		} else if test.want != got {
			t.Errorf("Test %v: wanted %v, got %v", i, test.want, got)
		}
	}
}
