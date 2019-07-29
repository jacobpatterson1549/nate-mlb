package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func serveView(w http.ResponseWriter, r *http.Request) {
	friendPlayerInfo, err := getFriendPlayerInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	scoreCategories, err := getStats(friendPlayerInfo)
	// jsonStr := `[{"Name":"team","FriendScores":[{"Name":"Bob","PlayerScores":[{"Name":"Boston Red Sox","Score":59},{"Name":"Chicago Cubs","Score":56},{"Name":"Minnesota Twins","Score":64},{"Name":"Seattle Mariners","Score":46},{"Name":"Kansas City Royals","Score":40}],"Score":265},{"Name":"W","PlayerScores":[{"Name":"New York Yankees","Score":67},{"Name":"Philadelphia Phillies","Score":55},{"Name":"Colorado Rockies","Score":49},{"Name":"Pittsburgh Pirates","Score":46},{"Name":"Miami Marlins","Score":40}],"Score":257},{"Name":"Nate","PlayerScores":[{"Name":"Milwaukee Brewers","Score":56},{"Name":"St. Louis Cardinals","Score":56},{"Name":"Los Angeles Angels","Score":55},{"Name":"Chicago White Sox","Score":46},{"Name":"San Francisco Giants","Score":54}],"Score":267},{"Name":"Sam","PlayerScores":[{"Name":"Houston Astros","Score":68},{"Name":"Tampa Bay Rays","Score":60},{"Name":"New York Mets","Score":50},{"Name":"Toronto Blue Jays","Score":40},{"Name":"San Francisco Giants","Score":54}],"Score":272},{"Name":"Steve","PlayerScores":[{"Name":"Los Angeles Dodgers","Score":69},{"Name":"Atlanta Braves","Score":62},{"Name":"Arizona Diamondbacks","Score":53},{"Name":"Cincinnati Reds","Score":48},{"Name":"Detroit Tigers","Score":30}],"Score":262},{"Name":"Mike","PlayerScores":[{"Name":"Cleveland Indians","Score":62},{"Name":"Washington Nationals","Score":56},{"Name":"Oakland Athletics","Score":60},{"Name":"San Diego Padres","Score":49},{"Name":"San Francisco Giants","Score":54}],"Score":281}]},{"Name":"hitters","FriendScores":[]},{"Name":"pitchers","FriendScores":[]}]`
	// scoreCategories := []ScoreCategory{}
	// err := json.Unmarshal([]byte(jsonStr), &scoreCategories)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates, err := template.ParseFiles(
		"templates/main.html",
		"templates/view.html",
		"templates/scoreCategory.html",
		"templates/friendScore.html",
		"templates/playerScore.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.Execute(w, scoreCategories); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func startServer(portNumber int) {
	http.HandleFunc("/", serveView)
	addr := fmt.Sprintf(":%d", portNumber)
	http.ListenAndServe(addr, nil)
}
