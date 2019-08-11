package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"nate-mlb/internal/db"
	"nate-mlb/internal/stats"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Run configures and starts the server
func Run(portNumber int) error {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handle)

	addr := fmt.Sprintf(":%d", portNumber)
	err := http.ListenAndServe(addr, nil)
	if err != http.ErrServerClosed {
		return fmt.Errorf("server stopped unexpectedly: %v", err)
	}
	return nil
}

func handle(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err == nil {
		switch {
		case r.Method == "GET" && r.RequestURI == "/":
			err = writeStatsPage(w)
		case r.Method == "GET" && r.RequestURI == "/about":
			err = writeAboutPage(w)
		case r.Method == "GET" && r.URL.Path == "/admin/password":
			err = handleHashPassword(w, r)
		case (r.Method == "GET" || r.Method == "POST") && r.URL.Path == "/admin":
			err = handleAdminPage(w, r)
		case r.Method == "GET" && r.URL.Path == "/admin/search":
			err = handlePlayerSearch(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}
	if err != nil {
		log.Printf("server error: %q", err)
		http.Error(w, err.Error(), http.StatusInternalServerError) // will warn "http: superfluous response.WriteHeader call" if template write fails
	}
}

func handleHashPassword(w http.ResponseWriter, r *http.Request) error {
	password := r.Form.Get("v")
	if len(password) == 0 {
		return errors.New("missing query param: v")
	}
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(hashedPassword))
	return err
}

func handleAdminPage(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		message := r.Form.Get("message")
		if len(message) == 0 {
			message = "Enter password before submitting."
		}
		return writeAdminPage(w, message)
	}

	var message string
	if err := handleAdminRequest(r); err != nil {
		message = err.Error()
	} else {
		message = "Change made."
	}
	// prevent the post from being made again on refresh
	http.Redirect(w, r, fmt.Sprintf("/admin?message=%s", message), http.StatusSeeOther)
	return nil
}

func handlePlayerSearch(w http.ResponseWriter, r *http.Request) error {
	searchQuery := r.Form.Get("q")
	if len(searchQuery) == 0 {
		return errors.New("missing search query param: q")
	}
	playerTypeID := r.Form.Get("pt")
	if len(playerTypeID) == 0 {
		return errors.New("missing player type query param: pt")
	}
	playerTypeIDI, err := strconv.Atoi(playerTypeID)
	if err != nil {
		return fmt.Errorf("problem converting playerTypeID (%v) to number: %v", playerTypeID, err)
	}
	activePlayersOnly := r.Form.Get("apo")
	activePlayersOnlyB := activePlayersOnly == "true"

	playerSearchResults, err := searchPlayers(playerTypeIDI, searchQuery, activePlayersOnlyB)
	if err != nil {
		return err
	}
	err = json.NewEncoder(w).Encode(playerSearchResults)
	if err != nil {
		return fmt.Errorf("problem converting PlayerSearchResults (%v) to json: %v", playerSearchResults, err)
	}
	return nil
}

func writeStatsPage(w http.ResponseWriter) error {
	es, err := stats.GetEtlStats()
	if err != nil {
		return err
	}

	tabs := make([]Tab, len(es.ScoreCategories))
	for i, sc := range es.ScoreCategories {
		tabs[i] = sc
	}
	timesMessage := TimesMessage{
		Messages: []string{"Stats reset daily after first page load is loaded after ", ".  Last reset at ", "."},
		Times:    []time.Time{es.EtlRefreshTime, es.EtlTime},
	}
	viewPage := Page{
		Title:            "Nate's MLB pool",
		Tabs:             tabs,
		TimesMessageJSON: timesMessage.toJSON(),
		templateNames:    []string{"templates/stats.html"},
		PageLoadTime:     db.GetUtcTime(),
	}

	return renderTemplate(w, viewPage)
}

func writeAboutPage(w http.ResponseWriter) error {
	lastDeploy, err := getLastDeploy()
	if err != nil {
		return err
	}

	timesMessage := TimesMessage{
		Messages: []string{"Server last deployed on ", fmt.Sprintf(" (version %s).", lastDeploy.Version)},
		Times:    []time.Time{lastDeploy.Time},
	}
	adminPage := Page{
		Title:            "About Nate's MLB",
		Tabs:             []Tab{AboutTab{}},
		TimesMessageJSON: timesMessage.toJSON(),
		templateNames:    []string{"templates/about.html"},
		PageLoadTime:     db.GetUtcTime(),
	}

	return renderTemplate(w, adminPage)
}

func writeAdminPage(w http.ResponseWriter, message string) error {
	es, err := stats.GetEtlStats()
	if err != nil {
		return err
	}
	years, err := db.GetYears()
	if err != nil {
		return err
	}

	scoreCategoriesData := make([]interface{}, len(es.ScoreCategories))
	for i, sc := range es.ScoreCategories {
		scoreCategoriesData[i] = sc
	}
	var friendsData []interface{}
	if len(es.ScoreCategories) > 0 {
		friendsData = make([]interface{}, len(es.ScoreCategories[0].FriendScores))
		for i, fs := range es.ScoreCategories[0].FriendScores {
			friendsData[i] = fs
		}
	}
	yearsData := make([]interface{}, len(years))
	for i, year := range years {
		yearsData[i] = year
	}

	adminTabs := []AdminTab{
		AdminTab{Name: "Players", Action: "players", Data: scoreCategoriesData},
		AdminTab{Name: "Friends", Action: "friends", Data: friendsData},
		AdminTab{Name: "Years", Action: "years", Data: yearsData},
		AdminTab{Name: "Clear Cache", Action: "cache"},
		AdminTab{Name: "Reset Password", Action: "password"},
	}
	tabs := make([]Tab, len(adminTabs))
	templateNames := make([]string, len(adminTabs)+2)
	templateNames[0] = "templates/admin.html"
	templateNames[1] = fmt.Sprintf("templates/admin-form-inputs/player-search.html")
	for i, adminTab := range adminTabs {
		tabs[i] = adminTab
		templateNames[i+2] = fmt.Sprintf("templates/admin-form-inputs/%s.html", adminTab.Action)
	}
	timesMessage := TimesMessage{Messages: []string{message}}
	adminPage := Page{
		Title:            "Nate's MLB pool [ADMIN MODE]",
		Tabs:             tabs,
		TimesMessageJSON: timesMessage.toJSON(),
		templateNames:    templateNames,
		PageLoadTime:     db.GetUtcTime(),
	}
	return renderTemplate(w, adminPage)
}

func renderTemplate(w http.ResponseWriter, p Page) error {
	templateNames := make([]string, len(p.templateNames)+1)
	templateNames[0] = "templates/main.html"
	for i, templateName := range p.templateNames {
		templateNames[i+1] = templateName
	}

	t, err := template.ParseFiles(templateNames...)
	if err != nil {
		return err
	}
	err = t.Execute(w, p)
	if err != nil {
		return fmt.Errorf("problem rendering templates (%v): %v", templateNames, err)
	}
	return nil
}

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

// GetName implements the Tab interface for AdminTab
func (at AdminTab) GetName() string {
	return at.Name
}

// GetID implements the Tab interface for AdminTab
func (at AdminTab) GetID() string {
	return strings.ReplaceAll(at.GetName(), " ", "-")
}

// AboutTab provides a constant tab with about information
type AboutTab struct{}

// GetName implements the Tab interface for AdminTab
func (at AboutTab) GetName() string {
	return "About"
}

// GetID implements the Tab interface for AdminTab
func (at AboutTab) GetID() string {
	return "About"
}

// TimesMessage contains times to insert between messages
type TimesMessage struct {
	Messages []string
	Times    []time.Time
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