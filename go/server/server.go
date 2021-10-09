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
		DisplayName    string
		Port           string
		NflAppKey      string
		LogRequestURIs bool
		HtmlFS         fs.FS
		JavascriptFS   fs.FS
		StaticFS       fs.FS
	}

	// Server contains data to serve pages for the user.
	Server struct {
		Config
		ds                ServerDatastore
		requestCache      *request.Cache
		sportEntries      []SportEntry
		sportTypesByURL   map[string]db.SportType
		log               *log.Logger
		scoreCategorizers map[db.PlayerType]request.ScoreCategorizer
		searchers         map[db.PlayerType]request.Searcher
		aboutRequester    AboutRequester
	}

	// ServerDatastore provides a way for the server to store and retrieve data.
	ServerDatastore interface {
		GetYears(st db.SportType) ([]db.Year, error)
		adminDatastore
		etlDatastore
	}

	// AboutRequester gets the previous deployment info for the app.
	AboutRequester interface {
		PreviousDeployment() (*request.Deployment, error)
	}
)

// New validates and creates a new Server from the config
func (cfg Config) New(log *log.Logger, ds ServerDatastore, httpClient request.HTTPClient) (*Server, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	switch {
	case log == nil:
		return nil, fmt.Errorf("log required")
	case ds == nil:
		return nil, fmt.Errorf("data store required")
	case httpClient == nil:
		return nil, fmt.Errorf("httpClient required")
	}
	sportTypes := ds.SportTypes()
	if _, ok := sportTypes[db.SportTypeNfl]; ok && len(cfg.NflAppKey) == 0 {
		return nil, fmt.Errorf("nfl app key required")
	}
	sportEntries := newSportEntries(sportTypes)
	sportTypesByURL := make(map[string]db.SportType, len(sportTypes))
	for st, sti := range sportTypes {
		sportTypesByURL[sti.URL] = st
	}
	c := request.NewCache(100)
	environment := cfg.DisplayName
	scoreCategorizers, searchers, aboutRequester := request.NewRequesters(httpClient, c, cfg.NflAppKey, environment, cfg.LogRequestURIs, log)
	s := Server{
		sportEntries:      sportEntries,
		sportTypesByURL:   sportTypesByURL,
		requestCache:      &c,
		scoreCategorizers: scoreCategorizers,
		searchers:         searchers,
		aboutRequester:    aboutRequester,
		log:               log,
		ds:                ds,
	}
	s.Config = cfg
	return &s, nil
}

func (cfg Config) validate() error {
	if _, err := strconv.Atoi(cfg.Port); err != nil {
		return fmt.Errorf("invalid port number: %s", cfg.Port)
	}
	switch {
	case cfg.HtmlFS == nil:
		return fmt.Errorf("html filesystem required")
	case cfg.JavascriptFS == nil:
		return fmt.Errorf("javascript filesystem required")
	case cfg.StaticFS == nil:
		return fmt.Errorf("static filesystem required")
	}
	return nil
}

// Run configures and starts the server
func (s Server) Run() error {
	h := s.handler()
	addr := fmt.Sprintf(":%s", s.Port)
	s.log.Println("starting server - locally running at http://127.0.0.1" + addr)
	if err := http.ListenAndServe(addr, h); err != http.ErrServerClosed { // BLOCKS
		return fmt.Errorf("server stopped unexpectedly: %w", err)
	}
	return nil
}

func (s Server) handler() http.Handler {
	mux := new(http.ServeMux)
	s.handleStatic(mux, "/robots.txt", "/favicon.ico")
	s.handleRoot(mux)
	return mux
}

func (s Server) handleStatic(mux *http.ServeMux, staticFilenames ...string) {
	staticFS := http.FileServer(http.FS(s.StaticFS))
	staticHandler := func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static" + r.URL.Path
		staticFS.ServeHTTP(w, r)
	}
	for _, staticFilename := range staticFilenames {
		mux.HandleFunc(staticFilename, staticHandler)
	}
}

func (s Server) handleRoot(mux *http.ServeMux) {
	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		st, path := s.transformURLPath(r)
		s.handleMethod(st, path, w, r)
	}
	mux.HandleFunc("/", rootHandler)
}

func (s Server) handleError(w http.ResponseWriter, err error) {
	s.log.Printf("server error: %q", err)
	http.Error(w, err.Error(), http.StatusInternalServerError) // will warn "http: superfluous response.WriteHeader call" if template write fails
}

func (s Server) handleMethod(st db.SportType, path string, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(st, path, w, r)
	case http.MethodPost:
		s.handlePost(st, path, w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s Server) handleGet(st db.SportType, path string, w http.ResponseWriter, r *http.Request) {
	switch path {
	case "/":
		s.handleHomePage(st, w, r)
	case "/about":
		s.handleAboutPage(st, w, r)
	case "/SportType":
		s.handleStatsPage(st, w, r)
	case "/SportType/export":
		s.handleExport(st, w, r)
	case "/SportType/admin":
		s.handleAdminPage(st, w, r)
	case "/SportType/admin/search":
		s.handleAdminSearch(st, w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s Server) handlePost(st db.SportType, path string, w http.ResponseWriter, r *http.Request) {
	switch path {
	case "/SportType/admin":
		s.handleAdminPost(st, w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s Server) handleHomePage(st db.SportType, w http.ResponseWriter, r *http.Request) {
	title := fmt.Sprintf("%s Stats", s.DisplayName)
	homeTab := AdminTab{Name: "Home"}
	homePage := newPage(s, title, []Tab{homeTab}, false, TimesMessage{}, "home")
	s.renderTemplate(w, homePage)
}

func (s Server) handleStatsPage(st db.SportType, w http.ResponseWriter, r *http.Request) {
	es, err := getEtlStats(st, s.ds, s.scoreCategorizers)
	if err != nil {
		s.handleError(w, err)
		return
	}
	tabs := make([]Tab, len(es.scoreCategories))
	stURL := s.ds.SportTypes()[es.sportType].URL
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
	stName := s.ds.SportTypes()[st].Name
	title := fmt.Sprintf("%s %s stats - %d", s.DisplayName, stName, es.year)
	statsPage := newPage(s, title, tabs, true, timesMessage, "stats")
	s.renderTemplate(w, statsPage)
}

func (s Server) handleAdminPage(st db.SportType, w http.ResponseWriter, r *http.Request) {
	es, err := getEtlStats(st, s.ds, s.scoreCategorizers)
	if err != nil {
		s.handleError(w, err)
		return
	}
	years, err := s.ds.GetYears(st)
	if err != nil {
		s.handleError(w, err)
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
	stName := s.ds.SportTypes()[st].Name
	title := fmt.Sprintf("%s %s [ADMIN MODE]", s.DisplayName, stName)
	adminPage := newPage(s, title, tabs, true, timesMessage, "admin")
	s.renderTemplate(w, adminPage)
}

func (s Server) handleAboutPage(st db.SportType, w http.ResponseWriter, r *http.Request) {
	lastDeploy, err := s.aboutRequester.PreviousDeployment()
	if err != nil {
		s.handleError(w, err)
		return
	}
	var timesMessage TimesMessage
	if lastDeploy != nil {
		timesMessage.Messages = []string{"Server last deployed on", fmt.Sprintf("version %s", lastDeploy.Version)}
		timesMessage.Times = []time.Time{lastDeploy.Time}
	}
	title := fmt.Sprintf("About %s Stats", s.DisplayName)
	aboutTab := AdminTab{Name: "About"}
	aboutPage := newPage(s, title, []Tab{aboutTab}, false, timesMessage, "about")
	s.renderTemplate(w, aboutPage)
}

func (s Server) handleExport(st db.SportType, w http.ResponseWriter, r *http.Request) {
	es, err := getEtlStats(st, s.ds, s.scoreCategorizers)
	if err != nil {
		s.handleError(w, err)
	}
	asOfDate := es.etlTime.Format("2006-01-02")
	fileName := fmt.Sprintf("%s_%s-%d_%s.csv", s.DisplayName, es.sportTypeName, es.year, asOfDate)
	contentDisposition := fmt.Sprintf(`attachment; filename="%s"`, fileName)
	w.Header().Set("Content-Disposition", contentDisposition)
	if err := exportToCsv(es, s.DisplayName, w); err != nil {
		s.handleError(w, err)
	}
}

func (s Server) renderTemplate(w http.ResponseWriter, p Page) {
	t, err := s.parseTemplate(w, p)
	if err != nil {
		s.handleError(w, fmt.Errorf("parsing template: %w", err))
		return
	}
	if err := t.Execute(w, p); err != nil {
		s.handleError(w, fmt.Errorf("rendering template: %w", err))
		return
	}
}

func (s Server) parseTemplate(w http.ResponseWriter, p Page) (*template.Template, error) {
	t := template.New("main.html")
	_, err := t.ParseFS(s.HtmlFS, "html/main/*.html")
	if err != nil {
		return nil, fmt.Errorf("loading template main files: %w", err)
	}
	_, err = t.ParseFS(s.HtmlFS, p.htmlFolderNameGlob())
	if err != nil {
		return nil, fmt.Errorf("loading template page files: %w", err)
	}
	jsFilenames, err := fs.Glob(s.JavascriptFS, "js/*/*.js")
	if err != nil {
		return nil, fmt.Errorf("reading js filenames: %w", err)
	}
	for _, jsFilename := range jsFilenames {
		t2, err := template.ParseFS(s.JavascriptFS, jsFilename)
		if err != nil {
			return nil, fmt.Errorf("loading template js file: %v: %w", jsFilename, err)
		}
		t.AddParseTree(jsFilename, t2.Tree)
	}
	if err != nil {
		return nil, fmt.Errorf("loading template js files: %w", err)
	}
	if _, err = t.ParseFS(s.StaticFS, "static/main.css"); err != nil {
		return nil, fmt.Errorf("loading template css file: %w", err)
	}
	return t, nil
}

func (s Server) handleAdminPost(st db.SportType, w http.ResponseWriter, r *http.Request) {
	if err := handleAdminPostRequest(s.ds, s.requestCache, st, r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Location", r.URL.Path)
	w.WriteHeader(http.StatusSeeOther)
}

func (s Server) handleAdminSearch(st db.SportType, w http.ResponseWriter, r *http.Request) {
	es, err := getEtlStats(st, s.ds, s.scoreCategorizers)
	if err != nil {
		s.handleError(w, err)
		return
	}
	playerSearchResults, err := handleAdminSearchRequest(es.year, s.searchers, r)
	if err != nil {
		s.handleError(w, err)
		return
	}
	if err := json.NewEncoder(w).Encode(playerSearchResults); err != nil {
		s.handleError(w, fmt.Errorf("converting PlayerSearchResults (%v) to json: %w", playerSearchResults, err))
		return
	}
}

func (s Server) transformURLPath(r *http.Request) (st db.SportType, path string) {
	urlPath := r.URL.Path
	parts := strings.Split(urlPath, "/")
	if len(parts) < 2 {
		return 0, urlPath
	}
	firstPathSegment := parts[1]
	st, ok := s.sportTypesByURL[firstPathSegment]
	if ok {
		urlPath = strings.Replace(urlPath, firstPathSegment, "SportType", 1)
	}
	return st, urlPath
}
