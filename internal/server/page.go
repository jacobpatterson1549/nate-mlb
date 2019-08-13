package server

import (
	"strings"
	"time"
)

// Page is a page that gets rendered by the main template
type Page struct {
	Title         string
	Tabs          []Tab
	TimesMessage  TimesMessage
	templateNames []string
	PageLoadTime  time.Time
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
