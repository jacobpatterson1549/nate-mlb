package request

import (
	"fmt"
	"time"
)

// PreviousDeployment returns some information about the most recent deployment
func PreviousDeployment() (Deployment, error) {
	owner := "jacobpatterson1549"
	repo := "nate-mlb"
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/deployments", owner, repo)
	grd := []GithubRepoDeployment{}
	err := requestJSON(url, &grd)
	previousDeployment := Deployment{}
	if err != nil || len(grd) == 0 {
		return previousDeployment, err
	}
	return grd[0].toDeployment()
}

// GithubRepoDeployment is used to unmarshal information about a github repository
type GithubRepoDeployment struct {
	Version string `json:"ref"`
	Time    string `json:"updated_at"`
}

// Deployment contains information about a deployment that is ready to be consumed
type Deployment struct {
	Version string
	Time    time.Time
}

func (grd *GithubRepoDeployment) toDeployment() (Deployment, error) {
	var (
		d   Deployment
		err error
	)
	d.Time, err = time.Parse(time.RFC3339, grd.Time)
	if err != nil {
		return d, fmt.Errorf("problem parsing %v into date: %v", grd.Time, err)
	}
	d.Version = grd.Version
	return d, nil
}
