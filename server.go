package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
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
	case strings.HasPrefix(r.RequestURI, "/admin"):
		err = handleAdminPage(w, r)
	default:
		pageNotFound(w)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleAdminPage(w http.ResponseWriter, r *http.Request) error {
	message := ""
	switch {
	case r.Method == "GET" && r.RequestURI == "/admin":
		message = "Enter password before submitting."
	case r.Method == "GET" && r.URL.Path == "/admin/password":
		if password, ok := r.URL.Query()["v"]; ok {
			hashedPassword, err := adminHashPassword(password[0])
			if err != nil {
				return err
			}
			w.Write([]byte(hashedPassword))
			return nil
		}
		return errors.New("missing query param: v")
	case r.Method == "POST":
		adminActions := map[string](func(*http.Request) error){
			"/admin/password": adminSetPassword,
			// "/admin/friends":  adminSetFriends,
			// "/admin/players":  adminSetPlayers,
			// "/admin/cache":    adminClearCache,
		}
		if adminAction, ok := adminActions[r.RequestURI]; ok {
			if err := adminAction(r); err != nil {
				message = err.Error()
			} else {
				message = "Change made at: " + time.Now().String()
			}
		}
	}
	if len(message) == 0 {
		pageNotFound(w)
		return nil
	}
	return writeAdminTabs(w, message)
}

func pageNotFound(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func writeView(w http.ResponseWriter) error {
	scoreCategories, err := getStats()
	if err != nil {
		return err
	}

	tabs := make([]Tab, len(scoreCategories))
	for i, sc := range scoreCategories {
		tabs[i] = sc
	}

	viewPage := Page{
		Title:        "Nate's MLB pool",
		Tabs:         tabs,
		Message:      "", // TODO: include etl date
		templateName: "view",
	}

	return renderTemplate(w, viewPage)
}

func writeAdminTabs(w http.ResponseWriter, message string) error {

	tabs := []Tab{
		AdminTab{Name: "Reset_Password", PostURL: "/admin/password"},
		AdminTab{Name: "Clear_Cache", PostURL: "/admin/cache"},
	}

	adminPage := Page{
		Title:        "Nate's MLB pool [ADMIN MODE]",
		Tabs:         tabs,
		Message:      message,
		templateName: "adminTabs",
	}

	return renderTemplate(w, adminPage)
}

func renderTemplate(w http.ResponseWriter, p Page) error {
	template, err := template.ParseFiles(
		"templates/main.html",
		fmt.Sprintf("templates/%s.html", p.templateName),
	)
	if err != nil {
		return err
	}

	return template.Execute(w, p)
}

// Page is a page that gets rendered by the main template
type Page struct {
	Title        string
	Tabs         []Tab
	Message      string
	templateName string
}

// Tab is a tab which gets rendered by the main template
type Tab interface {
	// TODO: this is also used as the id.  It must not have spaces
	GetName() string
}

// AdminTab provides the lowest level of tab data
type AdminTab struct {
	Name    string
	PostURL string
}

// GetName implements the Tab interface for AdminTab
func (at AdminTab) GetName() string {
	return at.Name
}

// GetName implements the Tab interface for ScoreCategory
func (sc ScoreCategory) GetName() string {
	return sc.Name
}
