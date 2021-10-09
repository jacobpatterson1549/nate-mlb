// Package server runs the http server after initializing the database.
package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
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
		log               *log.Logger
		htmlFS            fs.FS
		staticFS          fs.FS
		jsFS              fs.FS
	}
	serverDatastore interface {
		GetYears(st db.SportType) ([]db.Year, error)
		adminDatastore
		etlDatastore
	}
)

// NewConfig validates and creates a new configuration
func NewConfig(serverName string, ds serverDatastore, port, nflAppKey string, logRequestURIs bool, log *log.Logger, htmlFS, jsFS, staticFS fs.FS, httpClient request.HTTPClient) (*Config, error) {
	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("invalid port number: %s", port)
	}
	sportTypes := ds.SportTypes()
	sportEntries := newSportEntries(sportTypes)
	sportTypesByURL := make(map[string]db.SportType, len(sportTypes))
	for st, sti := range sportTypes {
		sportTypesByURL[sti.URL] = st
	}
	c := request.NewCache(100)
	environment := serverName
	scoreCategorizers, searchers, aboutRequester := request.NewRequesters(httpClient, c, nflAppKey, environment, logRequestURIs, log)
	return &Config{
		serverName:        serverName,
		ds:                ds,
		port:              port,
		sportEntries:      sportEntries,
		sportTypesByURL:   sportTypesByURL,
		requestCache:      &c,
		scoreCategorizers: scoreCategorizers,
		searchers:         searchers,
		aboutRequester:    aboutRequester,
		log:               log,
		htmlFS:            htmlFS,
		jsFS:              jsFS,
		staticFS:          staticFS,
	}, nil
}

// Run configures and starts the server
func Run(cfg Config) error {
	h := cfg.handler()
	addr := fmt.Sprintf(":%s", cfg.port)
	cfg.log.Println("starting server - locally running at http://127.0.0.1" + addr)
	if err := http.ListenAndServe(addr, h); err != http.ErrServerClosed { // BLOCKS
		return fmt.Errorf("server stopped unexpectedly: %w", err)
	}
	return nil
}

func (cfg Config) handler() http.Handler {
	mux := new(http.ServeMux)
	cfg.handleStatic(mux, "/robots.txt", "/favicon.ico")
	cfg.handleJavascriptFiles(mux)
	cfg.handleRoot(mux)
	return mux
}

func (cfg Config) handleStatic(mux *http.ServeMux, staticFilenames ...string) {
	staticFS := http.FileServer(http.FS(cfg.staticFS))
	staticHandler := func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static" + r.URL.Path
		staticFS.ServeHTTP(w, r)
	}
	for _, staticFilename := range staticFilenames {
		mux.HandleFunc(staticFilename, staticHandler)
	}
}

func (cfg Config) handleJavascriptFiles(mux *http.ServeMux) {
	jsHandler := http.FileServer(http.FS(cfg.jsFS))
	mux.Handle("/js/", jsHandler)
}

func (cfg Config) handleRoot(mux *http.ServeMux) {
	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		st, path := cfg.transformURLPath(r)
		cfg.handleMethod(st, path, w, r)
	}
	mux.HandleFunc("/", rootHandler)
}

func (cfg Config) handleError(w http.ResponseWriter, err error) {
	cfg.log.Printf("server error: %q", err)
	http.Error(w, err.Error(), http.StatusInternalServerError) // will warn "http: superfluous response.WriteHeader call" if template write fails
}

func (cfg Config) handleMethod(st db.SportType, path string, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg.handleGet(st, path, w, r)
	case http.MethodPost:
		cfg.handlePost(st, path, w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (cfg Config) handleGet(st db.SportType, path string, w http.ResponseWriter, r *http.Request) {
	switch path {
	case "/":
		cfg.handleHomePage(st, w, r)
	case "/about":
		cfg.handleAboutPage(st, w, r)
	case "/SportType":
		cfg.handleStatsPage(st, w, r)
	case "/SportType/export":
		cfg.handleExport(st, w, r)
	case "/SportType/admin":
		cfg.handleAdminPage(st, w, r)
	case "/SportType/admin/search":
		cfg.handleAdminSearch(st, w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (cfg Config) handlePost(st db.SportType, path string, w http.ResponseWriter, r *http.Request) {
	switch path {
	case "/SportType/admin":
		cfg.handleAdminPost(st, w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (cfg Config) handleHomePage(st db.SportType, w http.ResponseWriter, r *http.Request) {
	title := fmt.Sprintf("%s Stats", cfg.serverName)
	homeTab := AdminTab{Name: "Home"}
	homePage := newPage(cfg, title, []Tab{homeTab}, false, TimesMessage{}, "home")
	cfg.renderTemplate(w, homePage)
}

func (cfg Config) handleStatsPage(st db.SportType, w http.ResponseWriter, r *http.Request) {
	es, err := getEtlStats(st, cfg.ds, cfg.scoreCategorizers)
	if err != nil {
		cfg.handleError(w, err)
		return
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
	statsPage := newPage(cfg, title, tabs, true, timesMessage, "stats")
	cfg.renderTemplate(w, statsPage)
}

func (cfg Config) handleAdminPage(st db.SportType, w http.ResponseWriter, r *http.Request) {
	es, err := getEtlStats(st, cfg.ds, cfg.scoreCategorizers)
	if err != nil {
		cfg.handleError(w, err)
		return
	}
	years, err := cfg.ds.GetYears(st)
	if err != nil {
		cfg.handleError(w, err)
		return
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
	adminPage := newPage(cfg, title, tabs, true, timesMessage, "admin")
	cfg.renderTemplate(w, adminPage)
}

func (cfg Config) handleAboutPage(st db.SportType, w http.ResponseWriter, r *http.Request) {
	lastDeploy, err := cfg.aboutRequester.PreviousDeployment()
	if err != nil {
		cfg.handleError(w, err)
		return
	}
	var timesMessage TimesMessage
	if lastDeploy != nil {
		timesMessage.Messages = []string{"Server last deployed on", fmt.Sprintf("version %s", lastDeploy.Version)}
		timesMessage.Times = []time.Time{lastDeploy.Time}
	}
	title := fmt.Sprintf("About %s Stats", cfg.serverName)
	aboutTab := AdminTab{Name: "About"}
	aboutPage := newPage(cfg, title, []Tab{aboutTab}, false, timesMessage, "about")
	cfg.renderTemplate(w, aboutPage)
}

func (cfg Config) handleExport(st db.SportType, w http.ResponseWriter, r *http.Request) {
	es, err := getEtlStats(st, cfg.ds, cfg.scoreCategorizers)
	if err != nil {
		cfg.handleError(w, err)
	}
	asOfDate := es.etlTime.Format("2006-01-02")
	fileName := fmt.Sprintf("%s_%s-%d_%s.csv", cfg.serverName, es.sportTypeName, es.year, asOfDate)
	contentDisposition := fmt.Sprintf(`attachment; filename="%s"`, fileName)
	w.Header().Set("Content-Disposition", contentDisposition)
	if err := exportToCsv(es, cfg.serverName, w); err != nil {
		cfg.handleError(w, err)
	}
}

func (cfg Config) renderTemplate(w http.ResponseWriter, p Page) {
	// TODO: parse template when building handler
	t := template.New("main.html")
	_, err := t.ParseFS(cfg.htmlFS, "html/main/*.html")
	if err != nil {
		cfg.handleError(w, fmt.Errorf("loading template main files: %w", err))
		return
	}
	_, err = t.ParseFS(cfg.htmlFS, p.htmlFolderNameGlob())
	if err != nil {
		cfg.handleError(w, fmt.Errorf("loading template page files: %w", err))
		return
	}
	err = parseJavascriptFS(cfg.jsFS, t)
	if err != nil {
		cfg.handleError(w, fmt.Errorf("loading template js files: %w", err))
		return
	}
	if _, err = t.ParseFS(cfg.staticFS, "static/main.css"); err != nil {
		cfg.handleError(w, fmt.Errorf("loading template css file: %w", err))
		return
	}
	if err = t.Execute(w, p); err != nil {
		cfg.handleError(w, fmt.Errorf("rendering template: %w", err))
		return
	}
}

func parseJavascriptFS(jsFS fs.FS, t *template.Template) error {
	jsFilenames, err := fs.Glob(jsFS, "js/*/*.js")
	if err != nil {
		return fmt.Errorf("reading js filenames: %w", err)
	}
	for _, jsFilename := range jsFilenames {
		t2, err := template.ParseFS(jsFS, jsFilename)
		if err != nil {
			return err
		}
		t.AddParseTree(jsFilename, t2.Tree)
	}
	return nil
}

func (cfg Config) handleAdminPost(st db.SportType, w http.ResponseWriter, r *http.Request) {
	if err := handleAdminPostRequest(cfg.ds, cfg.requestCache, st, r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Location", r.URL.Path)
	w.WriteHeader(http.StatusSeeOther)
}

func (cfg Config) handleAdminSearch(st db.SportType, w http.ResponseWriter, r *http.Request) {
	es, err := getEtlStats(st, cfg.ds, cfg.scoreCategorizers)
	if err != nil {
		cfg.handleError(w, err)
		return
	}
	playerSearchResults, err := handleAdminSearchRequest(es.year, cfg.searchers, r)
	if err != nil {
		cfg.handleError(w, err)
		return
	}
	if err := json.NewEncoder(w).Encode(playerSearchResults); err != nil {
		cfg.handleError(w, fmt.Errorf("converting PlayerSearchResults (%v) to json: %w", playerSearchResults, err))
		return
	}
}

func (cfg Config) transformURLPath(r *http.Request) (st db.SportType, path string) {
	urlPath := r.URL.Path
	parts := strings.Split(urlPath, "/")
	if len(parts) < 2 {
		return 0, urlPath
	}
	firstPathSegment := parts[1]
	st, ok := cfg.sportTypesByURL[firstPathSegment]
	if ok {
		urlPath = strings.Replace(urlPath, firstPathSegment, "SportType", 1)
	}
	return st, urlPath
}
