package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

func startServer(portNumber int) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handle)

	addr := fmt.Sprintf(":%d", portNumber)
	http.ListenAndServe(addr, nil)
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
		log.Println(err)
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
	w.Write([]byte(hashedPassword))
	return nil
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
		return err
	}

	playerSearchResult, err := searchPlayers(playerTypeIDI, searchQuery)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(playerSearchResult)
}

func writeStatsPage(w http.ResponseWriter) error {
	es, err := getEtlStats()
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
		PageLoadTime:     getUtcTime(),
	}

	return renderTemplate(w, viewPage)
}

func writeAboutPage(w http.ResponseWriter) error {
	adminPage := Page{
		Title: "About Nate's MLB",
		Tabs:  []Tab{AboutTab{}},
		// TimesMessage:       "", // TODO: updated info (last deploy time)?
		templateNames: []string{"templates/about.html"},
		PageLoadTime:  getUtcTime(),
	}

	return renderTemplate(w, adminPage)
}

func writeAdminPage(w http.ResponseWriter, message string) error {
	es, err := getEtlStats()
	if err != nil {
		return err
	}
	years, err := getYears()
	if err != nil {
		return err
	}

	scoreCategoriesData := make([]interface{}, len(es.ScoreCategories))
	for i, sc := range es.ScoreCategories {
		scoreCategoriesData[i] = sc
	}
	friendsData := scoreCategoriesData
	yearsData := make([]interface{}, len(years))
	for i, year := range years {
		yearsData[i] = year
	}

	adminTabs := []AdminTab{
		AdminTab{Name: "Players", Action: "players", Data: scoreCategoriesData},
		AdminTab{Name: "Friends", Action: "friends", Data: friendsData}, // TODO: return just friends
		AdminTab{Name: "Years", Action: "years", Data: yearsData},
		// TODO: use .Action attribute on ui for id purposes so .Name can contain spaces
		AdminTab{Name: "Clear_Cache", Action: "cache"},
		AdminTab{Name: "Reset_Password", Action: "password"},
	}
	tabs := make([]Tab, len(adminTabs))
	templateNames := make([]string, len(adminTabs)+1)
	templateNames[0] = "templates/admin.html"
	for i, adminTab := range adminTabs {
		tabs[i] = adminTab
		templateNames[i+1] = fmt.Sprintf("templates/admin-form-inputs/%s.html", adminTab.Action)
	}
	timesMessage := TimesMessage{Messages: []string{message}}
	adminPage := Page{
		Title:            "Nate's MLB pool [ADMIN MODE]",
		Tabs:             tabs,
		TimesMessageJSON: timesMessage.toJSON(),
		templateNames:    templateNames,
		PageLoadTime:     getUtcTime(),
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
	return t.Execute(w, p)
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC1123Z)
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
	// TODO: this is also used as the id.  It must not have spaces
	GetName() string
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

// GetName implements the Tab interface for ScoreCategory
func (sc ScoreCategory) GetName() string {
	return sc.Name
}

// AboutTab provides a constant tab with about information
type AboutTab struct{}

// GetName implements the Tab interface for AdminTab
func (at AboutTab) GetName() string {
	return "About"
}

// TimesMessage contains times to insert between messages
type TimesMessage struct {
	Messages []string
	Times    []time.Time
}

func (tm *TimesMessage) toJSON() string {
	if len(tm.Times) > len(tm.Messages) {
		log.Printf("Must have at least as many Messages as Times.  Found %q.\n", tm)
		return "[Invalid TimesMessag]"
	}
	b, err := json.Marshal(tm)
	if err != nil {
		log.Println(err)
		b = []byte{}
	}
	return string(b)
}
