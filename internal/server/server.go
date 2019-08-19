package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"nate-mlb/internal/db"
	"nate-mlb/internal/request"
	"net/http"
	"net/url"
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
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	firstPathSegment := getFirstPathSegment(r.URL.Path)
	st := db.GetSportType(firstPathSegment)
	switch {
	case r.Method == "GET" && r.RequestURI == "/":
		err = writeHomePage(w)
	case r.Method == "GET" && r.RequestURI == "/about":
		err = writeAboutPage(w)
	case r.Method == "GET" && r.RequestURI == "/"+firstPathSegment:
		err = writeStatsPage(st, w)
	case r.Method == "GET" && r.RequestURI == "/"+firstPathSegment+"/export":
		err = exportStats(st, w)
	case (r.Method == "GET" || r.Method == "POST") && r.URL.Path == "/"+firstPathSegment+"/admin":
		err = handleAdminPage(st, firstPathSegment, w, r)
	case r.Method == "GET" && r.URL.Path == "/"+firstPathSegment+"/admin/search":
		err = handlePlayerSearch(st, w, r)
	case r.Method == "GET" && r.URL.Path == "/admin/password":
		err = handleHashPassword(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	if err != nil {
		log.Printf("server error: %q", err)
		http.Error(w, err.Error(), http.StatusInternalServerError) // will warn "http: superfluous response.WriteHeader call" if template write fails
	}
}

func getFirstPathSegment(urlPath string) string {
	parts := strings.Split(urlPath, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func writeHomePage(w http.ResponseWriter) error {
	homeTab := AdminTab{Name: "Home"}
	homePage := newPage("Nate's Stats", []Tab{homeTab}, TimesMessage{}, "html/tmpl/home.html")
	return renderTemplate(w, homePage)
}

func writeStatsPage(st db.SportType, w http.ResponseWriter) error {
	es, err := getEtlStats(st)
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
	statsPage := newPage("Nate's MLB pool", tabs, timesMessage, "html/tmpl/stats.html")
	return renderTemplate(w, statsPage)
}

func writeAdminPage(st db.SportType, w http.ResponseWriter, message string) error {
	es, err := getEtlStats(st)
	if err != nil {
		return err
	}
	years, err := db.GetYears(st)
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
	templateNames[0] = "html/tmpl/admin.html"
	templateNames[1] = fmt.Sprintf("html/tmpl/admin-form-inputs/player-search.html")
	for i, adminTab := range adminTabs {
		tabs[i] = adminTab
		templateNames[i+2] = fmt.Sprintf("html/tmpl/admin-form-inputs/%s.html", adminTab.Action)
	}
	timesMessage := TimesMessage{Messages: []string{message}}
	adminPage := newPage("Nate's MLB pool [ADMIN MODE]", tabs, timesMessage, templateNames...)
	return renderTemplate(w, adminPage)
}

func writeAboutPage(w http.ResponseWriter) error {
	lastDeploy, err := request.PreviousDeployment()
	if err != nil {
		return err
	}

	timesMessage := TimesMessage{
		Messages: []string{"Server last deployed on ", fmt.Sprintf(" (version %s).", lastDeploy.Version)},
		Times:    []time.Time{lastDeploy.Time},
	}
	aboutTab := AdminTab{Name: "About"}
	aboutPage := newPage("About Nate's MLB", []Tab{aboutTab}, timesMessage, "html/tmpl/about.html")
	return renderTemplate(w, aboutPage)
}

func exportStats(st db.SportType, w http.ResponseWriter) error {
	es, err := getEtlStats(st)
	if err != nil {
		return err
	}

	activeYear, err := db.GetActiveYear(st)
	if err != nil {
		return err
	}
	asOfDate := es.EtlTime.Format("2006-01-02")
	fileName := fmt.Sprintf("nate-mlb_%d_%s.csv", activeYear, asOfDate)
	contentDisposition := fmt.Sprintf(`attachment; filename="%s"`, fileName)
	w.Header().Set("Content-Disposition", contentDisposition)

	return exportToCsv(st, es, w)
}

func renderTemplate(w http.ResponseWriter, p Page) error {
	templateNames := make([]string, len(p.templateNames)+1)
	templateNames[0] = "html/tmpl/main.html"
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

func handleAdminPage(st db.SportType, firstPathSegment string, w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		message := r.Form.Get("message")
		if len(message) == 0 {
			message = "Enter password before submitting."
		}
		return writeAdminPage(st, w, message)
	}

	var message string
	err := handleAdminRequest(st, r)
	if err != nil {
		message = err.Error()
	} else {
		message = "Change made."
	}
	// prevent the post from being made again on refresh
	message = url.QueryEscape(message)
	http.Redirect(w, r, fmt.Sprintf("/%s/admin?message=%s", firstPathSegment, message), http.StatusSeeOther)
	return nil
}

func handlePlayerSearch(st db.SportType, w http.ResponseWriter, r *http.Request) error {
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
	playerType := db.PlayerType(playerTypeIDI)
	activePlayersOnly := r.Form.Get("apo")
	activePlayersOnlyB := activePlayersOnly == "true"

	searcher, ok := request.Searchers[playerType]
	if !ok {
		return fmt.Errorf("problem finding searcher for playerType %v", playerType)
	}
	playerSearchResults, err := searcher.PlayerSearchResults(st, searchQuery, activePlayersOnlyB)
	if err != nil {
		return err
	}
	err = json.NewEncoder(w).Encode(playerSearchResults)
	if err != nil {
		return fmt.Errorf("problem converting PlayerSearchResults (%v) to json: %v", playerSearchResults, err)
	}
	return nil
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
