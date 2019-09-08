package server

import (
	"encoding/json"
	"errors"
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
func Run(portNumber int, databaseDriverName string, dataSourceName string) error {
	err := db.Init(databaseDriverName, dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	fileInfo, err := ioutil.ReadDir("static")
	if err != nil {
		return fmt.Errorf("problem reading static dir: %v", err)
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
		return fmt.Errorf("server stopped unexpectedly: %v", err)
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
		case "/", "/about", "/admin/password":
			break
		default:
			return fmt.Errorf("Unknown SportType: %v", firstPathSegment)
		}
	}
	var err error
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
		err = handleAdminSearch(st, w, r)
	case r.Method == "POST" && r.URL.Path == "/admin/password":
		err = handleAdminPassword(w, r)
	default:
		err = errors.New(http.StatusText(http.StatusNotFound))
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	return err
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
	homePage := newPage("Nate's Stats", []Tab{homeTab}, false, TimesMessage{}, "html/home.html")
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
		Messages: []string{"Stats reset daily after first page load is loaded after", "and last reset at"},
		Times:    []time.Time{es.etlRefreshTime, es.EtlTime},
	}
	title := fmt.Sprintf("Nate's %s pool - %d", st.Name(), es.year)
	statsPage := newPage(title, tabs, true, timesMessage, "html/statsTab.html")
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
	templateNames[0] = "html/adminTab.html"
	templateNames[1] = fmt.Sprintf("html/admin-form-inputs/player-search.html")
	for i, adminTab := range adminTabs {
		tabs[i] = adminTab
		templateNames[i+2] = fmt.Sprintf("html/admin-form-inputs/%s.html", adminTab.Action)
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
		Messages: []string{"Server last deployed on", fmt.Sprintf("version %s", lastDeploy.Version)},
		Times:    []time.Time{lastDeploy.Time},
	}
	aboutTab := AdminTab{Name: "About"}
	aboutPage := newPage("About Nate's Stats", []Tab{aboutTab}, false, timesMessage, "html/about.html")
	return renderTemplate(w, aboutPage)
}

func exportStats(st db.SportType, w http.ResponseWriter) error {
	es, err := getEtlStats(st)
	if err != nil {
		return err
	}

	asOfDate := es.EtlTime.Format("2006-01-02")
	fileName := fmt.Sprintf("nate-mlb_%s-%d_%s.csv", es.sportTypeName, es.year, asOfDate)
	contentDisposition := fmt.Sprintf(`attachment; filename="%s"`, fileName)
	w.Header().Set("Content-Disposition", contentDisposition)

	return exportToCsv(st, es, w)
}

func renderTemplate(w http.ResponseWriter, p Page) error {
	templateNames := append([]string{"html/main.html", "html/nav.html", "html/tabs.html", "html/footer.html"}, p.templateNames...)
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
	err := handleAdminPostRequest(st, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusSeeOther)
	}
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
		return fmt.Errorf("problem converting PlayerSearchResults (%v) to json: %v", playerSearchResults, err)
	}
	return nil
}

func handleAdminPassword(w http.ResponseWriter, r *http.Request) error {
	err := handleAdminPasswordRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	return err
}
