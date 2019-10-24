package server

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

func TestNewSportEntries(t *testing.T) {
	sportTypes := db.SportTypeMap{
		db.SportType(1): db.SportTypeInfo{
			Name:         "apples",
			URL:          "/a",
			DisplayOrder: 2,
		},
		db.SportType(2): db.SportTypeInfo{
			Name:         "berries",
			URL:          "/b",
			DisplayOrder: 1,
		},
	}
	want := []SportEntry{
		{
			URL:       "/b",
			Name:      "berries",
			sportType: db.SportType(2),
		},
		{
			URL:       "/a",
			Name:      "apples",
			sportType: db.SportType(1),
		},
	}
	got := newSportEntries(sportTypes)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("not equal:\nwant: %v\ngot:  %v", want, got)
	}
}

func TestNewPage(t *testing.T) {
	applicationName := "aWesome App"
	sportEntries := []SportEntry{
		{
			URL:       "/q",
			Name:      "query",
			sportType: db.SportType(8),
		},
	}
	time1 := time.Date(2019, time.October, 17, 3, 19, 42, 197, time.UTC)
	time2 := time.Date(2019, time.October, 17, 3, 19, 42, 200, time.UTC)
	time3 := time.Date(2019, time.June, 6, 12, 0, 0, 0, time.UTC)
	ds := mockServerDatastore{
		nil,
		nil,
		mockEtlDatastore{
			GetUtcTimeFunc: func() time.Time {
				return time1
			},
		},
	}
	title := "Shall be lifted -- nevermore!"
	tabs := []Tab{} // none
	showTabs := true
	timesMessage := TimesMessage{
		Messages: []string{"a", "scary", "omen"},
		Times:    []time.Time{time2, time3},
	}
	htmlFolderName := "edgar_p"
	cfg := Config{
		serverName:   applicationName,
		sportEntries: sportEntries,
		ds:           ds,
	}
	want := Page{
		ApplicationName: applicationName,
		Title:           title,
		Tabs:            tabs,
		htmlFolderName:  htmlFolderName,
		ShowTabs:        showTabs,
		Sports:          sportEntries,
		TimesMessage:    timesMessage,
		PageLoadTime:    time1,
	}
	got := newPage(cfg, title, tabs, showTabs, timesMessage, htmlFolderName)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("not equal:\nwant: %v\ngot:  %v", want, got)
	}
}

type mockTimeGetter struct {
	GetUtcTimeFunc func() time.Time
}

func (tg mockTimeGetter) GetUtcTime() time.Time {
	return tg.GetUtcTimeFunc()
}

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
