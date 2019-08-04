package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func startServer(portNumber int) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handle)

	addr := fmt.Sprintf(":%d", portNumber)
	http.ListenAndServe(addr, nil)
}

func handle(w http.ResponseWriter, r *http.Request) {
	var err error
	switch {
	case r.Method == "GET" && r.RequestURI == "/":
		err = writeView(w)
	case r.Method == "GET" && r.RequestURI == "/about":
		err = writeAbout(w)
	case r.Method == "GET" && r.URL.Path == "/admin/password":
		err = handleHashPassword(w, r)
	case (r.Method == "GET" || r.Method == "POST") && r.URL.Path == "/admin":
		err = handleAdminPage(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleHashPassword(w http.ResponseWriter, r *http.Request) error {
	if password, ok := r.URL.Query()["v"]; ok {
		hashedPassword, err := hashPassword(password[0])
		if err != nil {
			return err
		}
		w.Write([]byte(hashedPassword))
		return nil
	}
	return errors.New("missing query param: v")
}

func handleAdminPage(w http.ResponseWriter, r *http.Request) error {
	var message string
	if r.Method == "GET" {
		if queryMessages, ok := r.URL.Query()["message"]; ok {
			message = queryMessages[0]
		} else {
			message = "Enter password before submitting."
		}
		return writeAdminTabs(w, message)
	}
	if err := handleAdminRequest(r); err != nil {
		message = err.Error()
	} else {
		message = fmt.Sprintf("Change made at: %s.", formatTime(time.Now()))
	}
	// prevent the post from being made again on refresh
	http.Redirect(w, r, fmt.Sprintf("/admin?message=%s", message), http.StatusSeeOther)
	return nil
}

func writeView(w http.ResponseWriter) error {
	es, err := getETLStats()
	if err != nil {
		return err
	}

	tabs := make([]Tab, len(es.Stats))
	for i, sc := range es.Stats {
		tabs[i] = sc
	}

	viewPage := Page{
		Title:         "Nate's MLB pool",
		Tabs:          tabs,
		Message:       fmt.Sprintf("Stats reset on first load after midnight.  Last load: %s.", formatTime(es.EtlTime)),
		templateNames: []string{"view"},
	}

	return renderTemplate(w, viewPage)
}

func writeAbout(w http.ResponseWriter) error {
	adminPage := Page{
		Title:         "About Nate's MLB",
		Tabs:          []Tab{AboutTab{}},
		Message:       "", // TODO: updated info?
		templateNames: []string{"about"},
	}

	return renderTemplate(w, adminPage)
}

func writeAdminTabs(w http.ResponseWriter, message string) error {
	es, err := getETLStats()
	if err != nil {
		return err
	}

	tabs := []Tab{
		AdminTab{Name: "Friends", Action: "friends", ScoreCategories: es.Stats},
		AdminTab{Name: "Clear_Cache", Action: "cache"},
		AdminTab{Name: "Reset_Password", Action: "password"},
	}

	adminPage := Page{
		Title:         "Nate's MLB pool [ADMIN MODE]",
		Tabs:          tabs,
		Message:       message,
		templateNames: []string{"adminTabs", "friendsForm", "playersForm"},
	}

	return renderTemplate(w, adminPage)
}

func renderTemplate(w http.ResponseWriter, p Page) error {
	templateNames := make([]string, len(p.templateNames)+1)
	templateNames[0] = "templates/main.html"
	for i, templateName := range p.templateNames {
		templateNames[i+1] = fmt.Sprintf("templates/%s.html", templateName)
	}
	template, err := template.ParseFiles(templateNames...)
	if err != nil {
		return err
	}

	return template.Execute(w, p)
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

// Page is a page that gets rendered by the main template
type Page struct {
	Title         string
	Tabs          []Tab
	Message       string
	templateNames []string
}

// Tab is a tab which gets rendered by the main template
type Tab interface {
	// TODO: this is also used as the id.  It must not have spaces
	GetName() string
}

// AdminTab provides tabs with admin tasks.
type AdminTab struct {
	Name            string
	Action          string
	ScoreCategories []ScoreCategory
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
