package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"nate-mlb/internal/db"
	"nate-mlb/internal/request"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Run configures and starts the server
func Run(portNumber int) error {
	fileInfo, err := ioutil.ReadDir("static")
	if err != nil {
		return fmt.Errorf("problem reading static dir: %v", err)
	}
	for _, file := range fileInfo {
		path := "/" + file.Name()
		http.HandleFunc(path, handleStatic)
	}
	http.HandleFunc("/", handleRoot)

	addr := fmt.Sprintf(":%d", portNumber)
	err = http.ListenAndServe(addr, nil)
	if err != http.ErrServerClosed {
		return fmt.Errorf("server stopped unexpectedly: %v", err)
	}
	return nil
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	path := "static" + r.URL.Path
	http.ServeFile(w, r, path)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	firstPathSegment := getFirstPathSegment(r.URL.Path)
	st := db.SportTypeFromURL(firstPathSegment)
	switch {
	case r.Method == "GET" && r.RequestURI == "/":
		err = writeHomePage(w)
	case r.Method == "GET" && r.RequestURI == "/about":
		err = writeAboutPage(w)
	case r.Method == "GET" && r.RequestURI == "/"+firstPathSegment:
		err = writeStatsPage(st, w)
	case r.Method == "GET" && r.RequestURI == "/"+firstPathSegment+"/export":
		err = exportStats(st, w)
	case r.Method == "GET" && r.URL.Path == "/"+firstPathSegment+"/admin":
		err = writeAdminPage(st, w)
	case r.Method == "POST" && r.URL.Path == "/"+firstPathSegment+"/admin":
		handleAdminPost(st, firstPathSegment, w, r)
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
	homePage := newPage("Nate's Stats", []Tab{homeTab}, false, TimesMessage{}, "html/tmpl/home.html")
	return renderTemplate(w, homePage)
}

func writeStatsPage(st db.SportType, w http.ResponseWriter) error {
	es, err := getEtlStats(st)
	if err != nil {
		return err
	}

	tabs := make([]Tab, len(es.ScoreCategories))
	for i, sc := range es.ScoreCategories {
		tabs[i] = StatsTab{
			ScoreCategory: sc,
			ExportURL:     fmt.Sprintf("/%s/export", es.sportType.URL()),
		}
	}
	timesMessage := TimesMessage{
		Messages: []string{"Stats reset daily after first page load is loaded after ", ".  Last reset at ", "."},
		Times:    []time.Time{es.etlRefreshTime, es.EtlTime},
	}
	title := fmt.Sprintf("Nate's %s pool - %d", st.Name(), es.year)
	statsPage := newPage(title, tabs, true, timesMessage, "html/tmpl/stats.html")
	return renderTemplate(w, statsPage)
}

func writeAdminPage(st db.SportType, w http.ResponseWriter) error {
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
	timesMessage := TimesMessage{}
	title := fmt.Sprintf("Nate's %s pool [ADMIN MODE]", st.Name())
	adminPage := newPage(title, tabs, true, timesMessage, templateNames...)
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
	aboutPage := newPage("About Nate's Stats", []Tab{aboutTab}, false, timesMessage, "html/tmpl/about.html")
	return renderTemplate(w, aboutPage)
}

func exportStats(st db.SportType, w http.ResponseWriter) error {
	es, err := getEtlStats(st)
	if err != nil {
		return err
	}

	asOfDate := es.EtlTime.Format("2006-01-02")
	fileName := fmt.Sprintf("nate-mlb_%d_%s.csv", es.year, asOfDate)
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

func handleAdminPost(st db.SportType, firstPathSegment string, w http.ResponseWriter, r *http.Request) {
	var message string
	err := handleAdminRequest(st, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		message = err.Error()
	} else {
		message = "Change made."
	}
	w.Write([]byte(message))
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
	activePlayersOnlyB := activePlayersOnly == "on"

	_, err = db.LoadPlayerTypes(st)
	if err != nil {
		return err
	}
	searcher, ok := request.Searchers[playerType]
	if !ok {
		return fmt.Errorf("problem finding searcher for playerType %v", playerType)
	}
	playerSearchResults, err := searcher.PlayerSearchResults(playerType, searchQuery, activePlayersOnlyB)
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
