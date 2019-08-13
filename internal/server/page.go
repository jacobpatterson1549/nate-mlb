package server

import (
	"encoding/json"
	"log"
	"strings"
	"time"
)

// Page is a page that gets rendered by the main template
type Page struct {
	Title            string
	Tabs             []Tab
	TimesMessageJSON string
	templateNames    []string
	PageLoadTime     time.Time
}

// Tab is a tab which gets rendered by the main template
type Tab interface {
	GetName() string
	GetID() string
}

// AdminTab provides tabs with admin tasks.
type AdminTab struct {
	Name   string
	Action string
	Data   []interface{} // each template knows what data to expect
}

// AboutTab provides a constant tab with about information
type AboutTab struct{}

// TimesMessage contains times to insert between messages
type TimesMessage struct {
	Messages []string
	Times    []time.Time
}

// GetName implements the Tab interface for AdminTab
func (at AdminTab) GetName() string {
	return at.Name
}

// GetID implements the Tab interface for AdminTab
func (at AdminTab) GetID() string {
	return strings.ReplaceAll(at.GetName(), " ", "-")
}

// GetName implements the Tab interface for AdminTab
func (at AboutTab) GetName() string {
	return "About"
}

// GetID implements the Tab interface for AdminTab
func (at AboutTab) GetID() string {
	return "About"
}

func (tm *TimesMessage) toJSON() string {
	if len(tm.Times) > len(tm.Messages) {
		log.Printf("Must have at least as many Messages as Times.  Found %v", tm)
		return "[Invalid TimesMessage]"
	}
	b, err := json.Marshal(tm)
	if err != nil {
		log.Printf("problem converting TemplatesMessage (%v) to json: %v", tm, err)
		b = []byte{}
	}
	return string(b)
}
