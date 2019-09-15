package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

// Run configures and starts the server
func Run(portNumber int) error {
	fileInfo, err := ioutil.ReadDir("static")
	if err != nil {
		return fmt.Errorf("reading static dir: %w", err)
	}
	for _, file := range fileInfo {
		path := "/" + file.Name()
		http.HandleFunc(path, handleStatic)
	}
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.HandleFunc("/", handleRoot)
	addr := fmt.Sprintf(":%d", portNumber)
	fmt.Printf("starting server - locally running at http://127.0.0.1%s\n", addr)
	err = http.ListenAndServe(addr, nil)
	if err != http.ErrServerClosed {
		return fmt.Errorf("server stopped unexpectedly: %w", err)
	}
	return nil
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	path := "static" + r.URL.Path
	http.ServeFile(w, r, path)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	err := handlePage(w, r)
	if err != nil {
		log.Printf("server error: %q", err)
		http.Error(w, err.Error(), http.StatusInternalServerError) // will warn "http: superfluous response.WriteHeader call" if template write fails
	}
}

func handlePage(w http.ResponseWriter, r *http.Request) error {
	firstPathSegment := getFirstPathSegment(r.URL.Path)
	st := db.SportTypeFromURL(firstPathSegment)
	if st == 0 {
		switch r.URL.Path {
		case "/", "/about":
			break
		default:
			return fmt.Errorf("unknown SportType: %v", firstPathSegment)
		}
	}
	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/":
			return writeHomePage(w)
		case "/about":
			return writeAboutPage(w)
		case "/" + firstPathSegment:
			return writeStatsPage(st, w)
		case "/" + firstPathSegment + "/export":
			return exportStats(st, w)
		case "/" + firstPathSegment + "/admin":
			return writeAdminPage(st, w)
		case "/" + firstPathSegment + "/admin/search":
			return handleAdminSearch(st, w, r)
		}
	case "POST":
		if r.URL.Path == "/"+firstPathSegment+"/admin" {
			handleAdminPost(st, w, r)
			return nil
		}
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	return nil
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
	homePage := newPage("Nate's Stats", []Tab{homeTab}, false, TimesMessage{}, "home")
	return renderTemplate(w, homePage)
}

func writeStatsPage(st db.SportType, w http.ResponseWriter) error {
	es, err := getEtlStats(st)
	if err != nil {
		return err
	}
	tabs := make([]Tab, len(es.scoreCategories))
	for i, sc := range es.scoreCategories {
		tabs[i] = StatsTab{
			ScoreCategory: sc,
			ExportURL:     fmt.Sprintf("/%s/export", es.sportType.URL()),
		}
	}
	if len(tabs) == 0 {
		tabs = []Tab{
			StatsTab{
				ScoreCategory: request.ScoreCategory{Name: "No Stats"},
			},
		}
	}
	timesMessage := TimesMessage{
		Messages: []string{"Stats reset daily after first page load is loaded after", "and last reset at"},
		Times:    []time.Time{es.etlRefreshTime, es.etlTime},
	}
	title := fmt.Sprintf("Nate's %s pool - %d", st.Name(), es.year)
	statsPage := newPage(title, tabs, true, timesMessage, "stats")
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
	scoreCategoriesData := make([]interface{}, len(es.scoreCategories))
	for i, sc := range es.scoreCategories {
		scoreCategoriesData[i] = sc
	}
	var friendsData []interface{}
	if len(es.scoreCategories) > 0 {
		friendsData = make([]interface{}, len(es.scoreCategories[0].FriendScores))
		for i, fs := range es.scoreCategories[0].FriendScores {
			friendsData[i] = fs
		}
	}
	yearsData := make([]interface{}, len(years))
	for i, year := range years {
		yearsData[i] = year
	}
	tabs := []Tab{
		AdminTab{Name: "Players", Action: "players", Data: scoreCategoriesData},
		AdminTab{Name: "Friends", Action: "friends", Data: scoreCategoriesData},
		AdminTab{Name: "Years", Action: "years", Data: yearsData},
		AdminTab{Name: "Clear Cache", Action: "cache"},
		AdminTab{Name: "Reset Password", Action: "password"},
	}
	timesMessage := TimesMessage{}
	title := fmt.Sprintf("Nate's %s pool [ADMIN MODE]", st.Name())
	adminPage := newPage(title, tabs, true, timesMessage, "admin")
	return renderTemplate(w, adminPage)
}

func writeAboutPage(w http.ResponseWriter) error {
	lastDeploy, err := request.PreviousDeployment()
	if err != nil {
		return err
	}
	timesMessage := TimesMessage{
		Messages: []string{"Server last deployed on", fmt.Sprintf("version %s", lastDeploy.Version)},
		Times:    []time.Time{lastDeploy.Time},
	}
	aboutTab := AdminTab{Name: "About"}
	aboutPage := newPage("About Nate's Stats", []Tab{aboutTab}, false, timesMessage, "about")
	return renderTemplate(w, aboutPage)
}

func exportStats(st db.SportType, w http.ResponseWriter) error {
	es, err := getEtlStats(st)
	if err != nil {
		return err
	}
	asOfDate := es.etlTime.Format("2006-01-02")
	fileName := fmt.Sprintf("nate-mlb_%s-%d_%s.csv", es.sportTypeName, es.year, asOfDate)
	contentDisposition := fmt.Sprintf(`attachment; filename="%s"`, fileName)
	w.Header().Set("Content-Disposition", contentDisposition)
	return exportToCsv(st, es, w)
}

func renderTemplate(w http.ResponseWriter, p Page) error {
	t := template.New("main.html")
	_, err := t.ParseGlob("html/main/*.html")
	if err != nil {
		return fmt.Errorf("loading template main files: %w", err)
	}
	_, err = t.ParseGlob(p.tabFilePatternGlob())
	if err != nil {
		return fmt.Errorf("loading template page files: %w", err)
	}
	err = t.Execute(w, p)
	if err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}
	return nil
}

func handleAdminPost(st db.SportType, w http.ResponseWriter, r *http.Request) {
	err := handleAdminPostRequest(st, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusSeeOther)
}

func handleAdminSearch(st db.SportType, w http.ResponseWriter, r *http.Request) error {
	es, err := getEtlStats(st)
	if err != nil {
		return err
	}
	playerSearchResults, err := handleAdminSearchRequest(st, es.year, r)
	if err != nil {
		return err
	}
	err = json.NewEncoder(w).Encode(playerSearchResults)
	if err != nil {
		return fmt.Errorf("converting PlayerSearchResults (%v) to json: %w", playerSearchResults, err)
	}
	return nil
}
