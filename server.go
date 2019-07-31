package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
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
	var message string
	var err error
	switch {
	case r.Method == "GET" && r.RequestURI == "/admin":
		message = ""
	case r.Method == "POST":
		var body []byte
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		switch r.RequestURI {
		case "/admin/password":
			err = adminSetPassword(body)
		case "/admin/years":
			err = adminSetYears(body)
		case "/admin/friends":
			err = adminSetFriends(body)
		case "/admin/players":
			err = adminSetPlayers(body)
		default:
			pageNotFound(w)
			return nil
		}
	default:
		pageNotFound(w)
		return nil
	}
	if err != nil {
		message = err.Error()
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

	adminPage := Page{
		Title:        "Nate's MLB pool [ADMIN MODE]",
		Tabs:         []Tab{},
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
	GetName() string
}

// GetName implements the Tab interface for ScoreCategory
func (sc ScoreCategory) GetName() string {
	return sc.Name
}
