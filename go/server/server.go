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
	// Config contains fields which describe the server
	Config struct {
		serverName        string
		ds                serverDatastore
		port              string
		sportEntries      []SportEntry
		sportTypesByURL   map[string]db.SportType
		requestCache      *request.Cache
		scoreCategorizers map[db.PlayerType]request.ScoreCategorizer
		searchers         map[db.PlayerType]request.Searcher
		aboutRequester    request.AboutRequester
	}
	serverDatastore interface {
		GetUtcTime() time.Time
		GetYears(st db.SportType) ([]db.Year, error)
		GetStat(st db.SportType) (*db.Stat, error)
		GetFriends(st db.SportType) ([]db.Friend, error)
		GetPlayers(st db.SportType) ([]db.Player, error)
		SetStat(stat db.Stat) error
		SaveYears(st db.SportType, futureYears []db.Year) error
		SaveFriends(st db.SportType, futureFriends []db.Friend) error
		SavePlayers(st db.SportType, futurePlayers []db.Player) error
		ClearStat(st db.SportType) error
		SetUserPassword(username string, p db.Password) error
		IsCorrectUserPassword(username string, p db.Password) (bool, error)
		// db.Datastore
		db.SportTypeGetter
		db.PlayerTypeGetter
	}
	urlPathTransformer func(sportTypesByURL map[string]db.SportType, url string) (db.SportType, string)
	httpMethod         string
	sportTypeHandler   func(st db.SportType, cfg Config, w http.ResponseWriter, r *http.Request) error
	sportTypeHandlers  map[httpMethod]map[string]sportTypeHandler
	httpHandler        func(w http.ResponseWriter, r *http.Request)
)

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

// NewConfig validates and creates a new configuration
func NewConfig(serverName string, ds serverDatastore, port string) (*Config, error) {
	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("Invalid port number: %s", port)
	}
	sportTypes := ds.SportTypes()
	sportTypesByURL := make(map[string]db.SportType, len(sportTypes))
	for st, sti := range sportTypes {
		sportTypesByURL[sti.URL] = st
	}
	c := request.NewCache(100)
	scoreCategorizers, searchers, aboutRequester := request.NewRequesters(c)
	return &Config{
		serverName:        serverName,
		ds:                ds,
		port:              port,
		sportEntries:      newSportEntries(ds.SportTypes()),
		sportTypesByURL:   sportTypesByURL,
		scoreCategorizers: scoreCategorizers,
		searchers:         searchers,
		aboutRequester:    aboutRequester,
	}, nil
}

// Run configures and starts the server
func Run(cfg Config) error {
	fileInfo, err := ioutil.ReadDir("static")
	if err != nil {
		return fmt.Errorf("reading static dir: %w", err)
	}
	for _, file := range fileInfo {
		path := "/" + file.Name()
		http.HandleFunc(path, handleStatic)
	}
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.HandleFunc("/", handleRoot(cfg))
	addr := fmt.Sprintf(":%s", cfg.port)
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

func handleRoot(cfg Config) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handlePage(cfg, w, r, transformURLPath, serverSportTypeHandlers)
		if err != nil {
			log.Printf("server error: %q", err)
			http.Error(w, err.Error(), http.StatusInternalServerError) // will warn "http: superfluous response.WriteHeader call" if template write fails
		}
	}
}

func handlePage(cfg Config, w http.ResponseWriter, r *http.Request, upt urlPathTransformer, sth sportTypeHandlers) error {
	st, urlPath := upt(cfg.sportTypesByURL, r.URL.Path)
	sportTypeHandler, ok := sth[httpMethod(r.Method)][urlPath]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return nil
	}
	return sportTypeHandler(st, cfg, w, r)
}

func transformURLPath(sportTypesByURL map[string]db.SportType, urlPath string) (db.SportType, string) {
	parts := strings.Split(urlPath, "/")
	if len(parts) < 2 {
		return 0, urlPath
	}
	firstPathSegment := parts[1]
	st, ok := sportTypesByURL[firstPathSegment]
	if ok {
		urlPath = strings.Replace(urlPath, firstPathSegment, "SportType", 1)
	}
	return st, urlPath
}

func handleHomePage(st db.SportType, cfg Config, w http.ResponseWriter, r *http.Request) error {
	title := fmt.Sprintf("%s Stats", cfg.serverName)
	homeTab := AdminTab{Name: "Home"}
	homePage := newPage(cfg.serverName, cfg.sportEntries, cfg.ds, title, []Tab{homeTab}, false, TimesMessage{}, "home")
	return renderTemplate(w, homePage)
}

func handleStatsPage(st db.SportType, cfg Config, w http.ResponseWriter, r *http.Request) error {
	es, err := getEtlStats(st, cfg.ds, cfg.scoreCategorizers)
	if err != nil {
		return err
	}
	tabs := make([]Tab, len(es.scoreCategories))
	stURL := cfg.ds.SportTypes()[es.sportType].URL
	for i, sc := range es.scoreCategories {
		tabs[i] = StatsTab{
			ScoreCategory: sc,
			ExportURL:     fmt.Sprintf("/%s/export", stURL),
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
	stName := cfg.ds.SportTypes()[st].Name
	title := fmt.Sprintf("%s %s stats - %d", cfg.serverName, stName, es.year)
	statsPage := newPage(cfg.serverName, cfg.sportEntries, cfg.ds, title, tabs, true, timesMessage, "stats")
	return renderTemplate(w, statsPage)
}

func handleAdminPage(st db.SportType, cfg Config, w http.ResponseWriter, r *http.Request) error {
	es, err := getEtlStats(st, cfg.ds, cfg.scoreCategorizers)
	if err != nil {
		return err
	}
	years, err := cfg.ds.GetYears(st)
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
	stName := cfg.ds.SportTypes()[st].Name
	title := fmt.Sprintf("%s %s [ADMIN MODE]", cfg.serverName, stName)
	adminPage := newPage(cfg.serverName, cfg.sportEntries, cfg.ds, title, tabs, true, timesMessage, "admin")
	return renderTemplate(w, adminPage)
}

func handleAboutPage(st db.SportType, cfg Config, w http.ResponseWriter, r *http.Request) error {
	lastDeploy, err := cfg.aboutRequester.PreviousDeployment()
	if err != nil {
		return err
	}
	timesMessage := TimesMessage{
		Messages: []string{"Server last deployed on", fmt.Sprintf("version %s", lastDeploy.Version)},
		Times:    []time.Time{lastDeploy.Time},
	}
	title := fmt.Sprintf("About %s Stats", cfg.serverName)
	aboutTab := AdminTab{Name: "About"}
	aboutPage := newPage(cfg.serverName, cfg.sportEntries, cfg.ds, title, []Tab{aboutTab}, false, timesMessage, "about")
	return renderTemplate(w, aboutPage)
}

func handleExport(st db.SportType, cfg Config, w http.ResponseWriter, r *http.Request) error {
	es, err := getEtlStats(st, cfg.ds, cfg.scoreCategorizers)
	if err != nil {
		return err
	}
	asOfDate := es.etlTime.Format("2006-01-02")
	fileName := fmt.Sprintf("%s_%s-%d_%s.csv", cfg.serverName, es.sportTypeName, es.year, asOfDate)
	contentDisposition := fmt.Sprintf(`attachment; filename="%s"`, fileName)
	w.Header().Set("Content-Disposition", contentDisposition)
	return exportToCsv(es, cfg.serverName, w)
}

func renderTemplate(w http.ResponseWriter, p Page) error {
	t := template.New("main.html")
	_, err := t.ParseGlob("html/main/*.html")
	if err != nil {
		return fmt.Errorf("loading template main files: %w", err)
	}
	_, err = t.ParseGlob(p.htmlFolderNameGlob())
	if err != nil {
		return fmt.Errorf("loading template page files: %w", err)
	}
	err = t.Execute(w, p)
	if err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}
	return nil
}

func handleAdminPost(st db.SportType, cfg Config, w http.ResponseWriter, r *http.Request) error {
	err := handleAdminPostRequest(cfg.ds, cfg.requestCache, st, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return nil
	}
	w.WriteHeader(http.StatusSeeOther)
	return nil
}

func handleAdminSearch(st db.SportType, cfg Config, w http.ResponseWriter, r *http.Request) error {
	es, err := getEtlStats(st, cfg.ds, cfg.scoreCategorizers)
	if err != nil {
		return err
	}
	playerSearchResults, err := handleAdminSearchRequest(cfg.ds, st, es.year, cfg.searchers, r)
	if err != nil {
		return err
	}
	err = json.NewEncoder(w).Encode(playerSearchResults)
	if err != nil {
		return fmt.Errorf("converting PlayerSearchResults (%v) to json: %w", playerSearchResults, err)
	}
	return nil
}
