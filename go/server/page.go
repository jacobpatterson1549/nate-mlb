package server

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

type (
	// Page is a page that gets rendered by the main template
	Page struct {
		Title        string
		Tabs         []Tab
		tabName      string
		ShowTabs     bool
		Sports       []SportEntry
		TimesMessage TimesMessage
		PageLoadTime time.Time
	}

	// Tab is a tab which gets rendered by the main template
	Tab interface {
		GetName() string
		GetID(string) string
	}

	jsTab struct{}

	// StatsTab provides stats information
	StatsTab struct {
		jsTab
		ScoreCategory request.ScoreCategory
		ExportURL     string
	}

	// AdminTab provides tabs with admin tasks.
	AdminTab struct {
		jsTab
		Name   string
		Action string
		Data   []interface{} // each template knows what data to expect
	}

	// SportEntry contains the url and name of a SportType
	SportEntry struct {
		URL  string
		Name string
	}

	// TimesMessage contains times to insert between messages
	TimesMessage struct {
		Messages []string
		Times    []time.Time
	}
)

func newPage(title string, tabs []Tab, showTabs bool, timesMessage TimesMessage, tabName string) Page {
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
		Title:        title,
		Tabs:         tabs,
		tabName:      tabName,
		Sports:       sports,
		ShowTabs:     showTabs,
		TimesMessage: timesMessage,
		PageLoadTime: db.GetUtcTime(),
	}
}

func (p Page) tabFilePatternGlob() string {
	return fmt.Sprintf("html/%s/*.html", p.tabName)
}

// GetID returns the js-safe id for the specified name
// https://www.w3.org/TR/html4/types.html#type-id
func (jsTab) GetID(name string) string {
	invalidCharacterRegex := regexp.MustCompile("[^-_:.A-Za-z0-9]")
	return strings.ToLower(invalidCharacterRegex.ReplaceAllString(name, "-"))
}

// GetName implements the Tab interface for AdminTab
func (at AdminTab) GetName() string {
	return at.Name
}

// GetName implements the Tab interface for StatsTab
func (st StatsTab) GetName() string {
	return st.ScoreCategory.Name
}
