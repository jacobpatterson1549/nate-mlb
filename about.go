package main

import (
	"fmt"
	"time"
)

func getLastDeploy() (Deploy, error) {
	owner := "jacobpatterson1549"
	repo := "nate-mlb"
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/deployments", owner, repo)

	grd := []GithubRepoDeploymentJSON{}
	lastDeploy := Deploy{}
	err := requestJSON(url, &grd)
	if err == nil && len(grd) > 0 {
		lastDeploy.Version = grd[0].Version
		lastDeploy.Time, err = time.Parse(time.RFC3339, grd[0].Time)
	}
	return lastDeploy, err
}

// GithubRepoDeploymentJSON is used to unmarshal information about a github repository
type GithubRepoDeploymentJSON struct {
	Version string `json:"ref"`
	Time    string `json:"updated_at"`
}

// Deploy contains information about a deployment that is ready to be consumed
type Deploy struct {
	Version string
	Time    time.Time
}
