package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func startServer(portNumber int) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handleView)

	addr := fmt.Sprintf(":%d", portNumber)
	http.ListenAndServe(addr, nil)
}

func handleView(w http.ResponseWriter, r *http.Request) {
	err := writeScoreSategories(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func writeScoreSategories(w http.ResponseWriter) error {
	friendPlayerInfo, err := getFriendPlayerInfo()
	if err != nil {
		return err
	}

	scoreCategories, err := getStats(friendPlayerInfo)
	// jsonBytes, err := json.Marshal(scoreCategories)
	// if jsonBytes != nil {
	// 	fmt.Println(string(jsonBytes))
	// }
	// scoreCategories := []ScoreCategory{}
	// err := json.Unmarshal([]byte(`[{"Name":"team","FriendScores":[{"Name":"Bob","PlayerScores":[{"Name":"Boston Red Sox","Score":59},{"Name":"Chicago Cubs","Score":56},{"Name":"Minnesota Twins","Score":64},{"Name":"Seattle Mariners","Score":46},{"Name":"Kansas City Royals","Score":40}],"Score":265},{"Name":"W","PlayerScores":[{"Name":"New York Yankees","Score":67},{"Name":"Philadelphia Phillies","Score":55},{"Name":"Colorado Rockies","Score":50},{"Name":"Pittsburgh Pirates","Score":46},{"Name":"Miami Marlins","Score":41}],"Score":259},{"Name":"Nate","PlayerScores":[{"Name":"Milwaukee Brewers","Score":56},{"Name":"St. Louis Cardinals","Score":56},{"Name":"Los Angeles Angels","Score":55},{"Name":"Chicago White Sox","Score":46},{"Name":"San Francisco Giants","Score":54}],"Score":267},{"Name":"Sam","PlayerScores":[{"Name":"Houston Astros","Score":68},{"Name":"Tampa Bay Rays","Score":60},{"Name":"New York Mets","Score":50},{"Name":"Toronto Blue Jays","Score":41},{"Name":"San Francisco Giants","Score":54}],"Score":273},{"Name":"Steve","PlayerScores":[{"Name":"Los Angeles Dodgers","Score":69},{"Name":"Atlanta Braves","Score":62},{"Name":"Arizona Diamondbacks","Score":53},{"Name":"Cincinnati Reds","Score":49},{"Name":"Detroit Tigers","Score":31}],"Score":264},{"Name":"Mike","PlayerScores":[{"Name":"Cleveland Indians","Score":62},{"Name":"Washington Nationals","Score":57},{"Name":"Oakland Athletics","Score":60},{"Name":"San Diego Padres","Score":50},{"Name":"San Francisco Giants","Score":54}],"Score":283}]},{"Name":"hitting","FriendScores":[{"Name":"Bob","PlayerScores":[{"Name":"J.D. Martinez","Score":22},{"Name":"Mookie Betts","Score":18},{"Name":"Jose Ramirez","Score":14}],"Score":40},{"Name":"W","PlayerScores":[{"Name":"Bryce Harper","Score":18},{"Name":"Rhys Hoskins","Score":22},{"Name":"Ronald Acuna Jr.","Score":25}],"Score":47},{"Name":"Nate","PlayerScores":[{"Name":"Mike Trout","Score":34},{"Name":"Nolan Arenado","Score":22},{"Name":"Manny Machado","Score":25}],"Score":59},{"Name":"Sam","PlayerScores":[{"Name":"Khris Davis","Score":16},{"Name":"Joey Gallo","Score":22},{"Name":"Francisco Lindor","Score":18}],"Score":40},{"Name":"Steve","PlayerScores":[{"Name":"Giancarlo Stanton","Score":1},{"Name":"Edwin Encarnacion","Score":30},{"Name":"Christian Yelich","Score":36}],"Score":66},{"Name":"Mike","PlayerScores":[{"Name":"Aaron Judge","Score":11},{"Name":"Trevor Story","Score":22},{"Name":"Paul Goldschmidt","Score":24}],"Score":46}]},{"Name":"pitching","FriendScores":[{"Name":"Bob","PlayerScores":[{"Name":"Blake Snell","Score":6}],"Score":6},{"Name":"W","PlayerScores":[{"Name":"Jacob deGrom","Score":6},{"Name":"Walker Buehler","Score":9},{"Name":"James Paxton","Score":5}],"Score":15},{"Name":"Nate","PlayerScores":[{"Name":"Corey Kluber","Score":2},{"Name":"Chris Sale","Score":5},{"Name":"Aaron Nola","Score":9}],"Score":14},{"Name":"Sam","PlayerScores":[{"Name":"Corey Kluber","Score":2},{"Name":"Carlos Carrasco","Score":4},{"Name":"Jon Lester","Score":9}],"Score":13},{"Name":"Steve","PlayerScores":[{"Name":"Chris Sale","Score":5},{"Name":"David Price","Score":7},{"Name":"Noah Syndergaard","Score":7}],"Score":14},{"Name":"Mike","PlayerScores":[{"Name":"Justin Verlander","Score":13},{"Name":"Gerrit Cole","Score":12},{"Name":"Trevor Bauer","Score":9}],"Score":25}]}]`), &scoreCategories)
	if err != nil {
		return err
	}

	template, err := template.ParseFiles(
		"templates/main.html",
		"templates/view.html",
	)
	if err != nil {
		return err
	}

	return template.Execute(w, scoreCategories)
}
