package server

import (
	"nate-mlb/internal/db"
	"nate-mlb/internal/request"
	"regexp"
	"strings"
	"time"
)

// Page is a page that gets rendered by the main template
type Page struct {
	Title         string
	Tabs          []Tab
	ShowTabs      bool
	Sports        []SportEntry
	TimesMessage  TimesMessage
	templateNames []string
	PageLoadTime  time.Time
}

// Tab is a tab which gets rendered by the main template
type Tab interface {
	GetName() string
	GetID(string) string
}

type jsTab struct{}

// StatsTab provides stats information
type StatsTab struct {
	jsTab
	ScoreCategory request.ScoreCategory
	ExportURL     string
}

// AdminTab provides tabs with admin tasks.
type AdminTab struct {
	jsTab
	Name   string
	Action string
	Data   []interface{} // each template knows what data to expect
}

// SportEntry contains the url and name of a SportType
type SportEntry struct {
	URL  string
	Name string
}

// TimesMessage contains times to insert between messages
type TimesMessage struct {
	Messages []string
	Times    []time.Time
}

func newPage(title string, tabs []Tab, showTabs bool, timesMessage TimesMessage, templateNames ...string) Page {
	getSportEntry := func(st db.SportType) SportEntry {
		return SportEntry{
			URL:  strings.ToLower(st.Name()),
			Name: st.Name(),
		}
	}
	sports := []SportEntry{
		getSportEntry(db.SportTypeMlb),
		getSportEntry(db.SportTypeNfl),
	}
	return Page{
		Title:         title,
		Tabs:          tabs,
		Sports:        sports,
		ShowTabs:      showTabs,
		TimesMessage:  timesMessage,
		templateNames: templateNames,
		PageLoadTime:  db.GetUtcTime(),
	}
}

// GetID returns the js-safe id for the specified name
// https://www.w3.org/TR/html4/types.html#type-id
func (jsTab) GetID(name string) string {
	invalidCharacterRegex := regexp.MustCompile("[^-_:.A-Za-z0-9]")
	return invalidCharacterRegex.ReplaceAllString(name, "-")
}

// GetName implements the Tab interface for AdminTab
func (at AdminTab) GetName() string {
	return at.Name
}

// GetName implements the Tab interface for StatsTab
func (st StatsTab) GetName() string {
	return st.ScoreCategory.Name
}
