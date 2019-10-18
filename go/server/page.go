package server

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

type (
	// Page is a page that gets rendered by the main template
	Page struct {
		ApplicationName string
		Title           string
		Tabs            []Tab
		htmlFolderName  string
		ShowTabs        bool
		Sports          []SportEntry
		TimesMessage    TimesMessage
		PageLoadTime    time.Time
	}

	// Tab is a tab which gets rendered by the main template
	Tab interface {
		GetName() string
		GetID() string
	}

	// StatsTab provides stats information
	StatsTab struct {
		ScoreCategory request.ScoreCategory
		ExportURL     string
	}

	// AdminTab provides tabs with admin tasks.
	AdminTab struct {
		Name   string
		Action string
		Data   []interface{} // each template knows what data to expect
	}

	// SportEntry contains the url and name of a SportType
	SportEntry struct {
		URL       string
		Name      string
		sportType db.SportType
	}

	// TimesMessage contains times to insert between messages
	TimesMessage struct {
		Messages []string
		Times    []time.Time
	}
)

var validJavascriptIDCharsRE = regexp.MustCompile("[^-_:.A-Za-z0-9]")

func newSportEntries(sportTypes db.SportTypeMap) []SportEntry {
	sportEntries := make([]SportEntry, 0, len(sportTypes))
	for st, stInfo := range sportTypes {
		sportEntry := SportEntry{
			URL:       stInfo.URL,
			Name:      stInfo.Name,
			sportType: st,
		}
		sportEntries = append(sportEntries, sportEntry)
	}
	displayOrder := func(i int) int { return sportTypes[sportEntries[i].sportType].DisplayOrder }
	sort.Slice(sportEntries, func(i, j int) bool {
		return displayOrder(i) < displayOrder(j)
	})
	return sportEntries
}

func newPage(applicationName string, sportEntries []SportEntry, tg timeGetter, title string, tabs []Tab, showTabs bool, timesMessage TimesMessage, htmlFolderName string) Page {
	return Page{
		ApplicationName: applicationName,
		Title:           title,
		Tabs:            tabs,
		htmlFolderName:  htmlFolderName,
		Sports:          sportEntries,
		ShowTabs:        showTabs,
		TimesMessage:    timesMessage,
		PageLoadTime:    tg.GetUtcTime(),
	}
}

func (p Page) htmlFolderNameGlob() string {
	return fmt.Sprintf("html/%s/*.html", p.htmlFolderName)
}

// GetID returns the js-safe id for the specified name
// https://www.w3.org/TR/html4/types.html#type-id
func jsID(name string) string {
	switch {
	case len(name) == 0:
		return "y"
	case validJavascriptIDCharsRE.MatchString(name[:1]):
		name = "z" + name
	}
	return strings.ToLower(validJavascriptIDCharsRE.ReplaceAllString(name, "-"))
}

// GetName implements the Tab interface for AdminTab
func (at AdminTab) GetName() string {
	return at.Name
}

// GetID implements the Tab interface for AdminTab
func (at AdminTab) GetID() string {
	return jsID(at.GetName())
}

// GetName implements the Tab interface for StatsTab
func (st StatsTab) GetName() string {
	return st.ScoreCategory.Name
}

// GetID implements the Tab interface for StatsTab
func (st StatsTab) GetID() string {
	return jsID(st.GetName())
}
