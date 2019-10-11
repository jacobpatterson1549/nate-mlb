package server

import (
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

func TestTabGetID(t *testing.T) {
	getNameTests := []struct {
		tab  Tab
		want string
	}{
		{
			tab: AdminTab{
				Name: "& Smart Functions",
			},
			want: "--smart-functions",
		},
		{
			tab: StatsTab{
				ScoreCategory: request.ScoreCategory{
					Name: "American Football",
				},
			},
			want: "american-football",
		},
	}
	for i, test := range getNameTests {
		got := test.tab.GetID(test.tab.GetName()) // TODO: GetID() should call getName internally
		if test.want != got {
			t.Errorf("Test %v: want %v, got %v", i, test.want, got)
		}
	}
}

func TestTabGetName(t *testing.T) {
	getNameTests := []struct {
		tab  Tab
		want string
	}{
		{
			tab: AdminTab{
				Name: "Smart Functions",
			},
			want: "Smart Functions",
		},
		{
			tab: StatsTab{
				ScoreCategory: request.ScoreCategory{
					Name: "Lacross",
				},
			},
			want: "Lacross",
		},
	}
	for i, test := range getNameTests {
		got := test.tab.GetName()
		if test.want != got {
			t.Errorf("Test %v: want %v, got %v", i, test.want, got)
		}
	}
}
