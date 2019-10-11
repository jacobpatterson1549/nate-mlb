package server

import (
	"strings"
	"testing"

	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

func TestHtmlFolderNameGlob(t *testing.T) {
	p := Page{htmlFolderName: "special_folder_name"}
	got := p.htmlFolderNameGlob()
	if !strings.Contains(got, p.htmlFolderName) {
		t.Errorf("%v does not contain %v", got, p.htmlFolderName)
	}
	wantEnding := "/*.html"
	if got[len(got)-len(wantEnding):] != wantEnding {
		t.Errorf("%v does not end with %v", got, wantEnding)
	}
}

func TestTabGetID(t *testing.T) {
	getNameTests := []struct {
		tab  Tab
		want string
	}{
		{
			tab: AdminTab{
				Name: "",
			},
			want: "y",
		},
		{
			tab: AdminTab{
				Name: "& Smart Functions",
			},
			want: "z--smart-functions",
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
		got := test.tab.GetID()
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
