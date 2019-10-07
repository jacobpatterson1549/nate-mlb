// Package server runs the http server after initializing the database.
package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jacobpatterson1549/nate-mlb/go/db"
	"github.com/jacobpatterson1549/nate-mlb/go/request"
)

type (
	sportTypeURLResolver func(string) db.SportType
	urlPathTransformer   func(string, sportTypeURLResolver) (db.SportType, string)
	httpMethod           string
	sportTypeHandler     func(st db.SportType, w http.ResponseWriter, r *http.Request) error
	sportTypeHandlers    map[httpMethod]map[string]sportTypeHandler
)

var serverName string
var serverSportTypeHandlers = sportTypeHandlers{
	"GET": {
		"/":                       handleHomePage,
		"/about":                  handleAboutPage,
		"/SportType":              handleStatsPage,
		"/SportType/export":       handleExport,
		"/SportType/admin":        handleAdminPage,
		"/SportType/admin/search": handleAdminSearch,
	},
	"POST": {
		"/SportType/admin": handleAdminPost,
	},
}

// Run configures and starts the server
func Run(port, applicationName string) error {
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("Invalid port number: %s", port)
	}
	serverName = applicationName
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
	addr := fmt.Sprintf(":%s", port)
	log.Println("starting server - locally running at http://127.0.0.1" + addr)
	err = http.ListenAndServe(addr, nil) // BLOCKS
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
	err := handlePage(w, r, db.SportTypeFromURL, transformURLPath, serverSportTypeHandlers)
	if err != nil {
		log.Printf("server error: %q", err)
		http.Error(w, err.Error(), http.StatusInternalServerError) // will warn "http: superfluous response.WriteHeader call" if template write fails
	}
}

func handlePage(w http.ResponseWriter, r *http.Request, stur sportTypeURLResolver, upt urlPathTransformer, sth sportTypeHandlers) error {
	st, urlPath := upt(r.URL.Path, stur)
	sportTypeHandler, ok := sth[httpMethod(r.Method)][urlPath]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return nil
	}
	return sportTypeHandler(st, w, r)
}

func transformURLPath(urlPath string, stur sportTypeURLResolver) (db.SportType, string) {
	parts := strings.Split(urlPath, "/")
	if len(parts) < 2 {
		return 0, urlPath
	}
	firstPathSegment := parts[1]
	st := stur(firstPathSegment)
	if st != 0 {
		urlPath = strings.Replace(urlPath, firstPathSegment, "SportType", 1)
	}
	return st, urlPath
}

func handleHomePage(st db.SportType, w http.ResponseWriter, r *http.Request) error {
	title := fmt.Sprintf("%s Stats", serverName)
	homeTab := AdminTab{Name: "Home"}
	homePage := newPage(serverName, title, []Tab{homeTab}, false, TimesMessage{}, "home")
	return renderTemplate(w, homePage)
}

func handleStatsPage(st db.SportType, w http.ResponseWriter, r *http.Request) error {
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
	title := fmt.Sprintf("%s %s stats - %d", serverName, st.Name(), es.year)
	statsPage := newPage(serverName, title, tabs, true, timesMessage, "stats")
	return renderTemplate(w, statsPage)
}

func handleAdminPage(st db.SportType, w http.ResponseWriter, r *http.Request) error {
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
	title := fmt.Sprintf("%s %s [ADMIN MODE]", serverName, st.Name())
	adminPage := newPage(serverName, title, tabs, true, timesMessage, "admin")
	return renderTemplate(w, adminPage)
}

func handleAboutPage(st db.SportType, w http.ResponseWriter, r *http.Request) error {
	lastDeploy, err := request.About.PreviousDeployment()
	if err != nil {
		return err
	}
	timesMessage := TimesMessage{
		Messages: []string{"Server last deployed on", fmt.Sprintf("version %s", lastDeploy.Version)},
		Times:    []time.Time{lastDeploy.Time},
	}
	title := fmt.Sprintf("About %s Stats", serverName)
	aboutTab := AdminTab{Name: "About"}
	aboutPage := newPage(serverName, title, []Tab{aboutTab}, false, timesMessage, "about")
	return renderTemplate(w, aboutPage)
}

func handleExport(st db.SportType, w http.ResponseWriter, r *http.Request) error {
	es, err := getEtlStats(st)
	if err != nil {
		return err
	}
	asOfDate := es.etlTime.Format("2006-01-02")
	fileName := fmt.Sprintf("%s_%s-%d_%s.csv", serverName, es.sportTypeName, es.year, asOfDate)
	contentDisposition := fmt.Sprintf(`attachment; filename="%s"`, fileName)
	w.Header().Set("Content-Disposition", contentDisposition)
	return exportToCsv(es, serverName, w)
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

func handleAdminPost(st db.SportType, w http.ResponseWriter, r *http.Request) error {
	err := handleAdminPostRequest(st, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return nil
	}
	w.WriteHeader(http.StatusSeeOther)
	return nil
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
