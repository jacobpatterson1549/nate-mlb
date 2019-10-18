package request

import (
	"fmt"
	"time"
)

type (
	// AboutRequester gets information about the most recent Deployment of the application
	AboutRequester struct {
		requester requester
	}

	// GithubRepoDeployment is used to unmarshal information about a github repository
	GithubRepoDeployment struct {
		Version string    `json:"ref"`
		Time    time.Time `json:"updated_at"`
	}

	// Deployment contains information about a deployment that is ready to be consumed
	Deployment struct {
		Version string
		Time    time.Time
	}
)

// PreviousDeployment returns some information about the most recent deployment
func (r *AboutRequester) PreviousDeployment() (Deployment, error) {
	owner := "jacobpatterson1549"
	repo := "nate-mlb"
	uri := fmt.Sprintf("https://api.github.com/repos/%s/%s/deployments", owner, repo)
	var grd []GithubRepoDeployment
	err := r.requester.structPointerFromURI(uri, &grd)

	var previousDeployment Deployment
	if err != nil {
		return previousDeployment, err
	}
	previousDeployment.setFromGithubRepoDeployments(grd)
	return previousDeployment, nil
}

func (d *Deployment) setFromGithubRepoDeployments(grd []GithubRepoDeployment) {
	if len(grd) != 0 {
		d.setFromGithubRepoDeployment(grd[0])
	}
}

func (d *Deployment) setFromGithubRepoDeployment(grd GithubRepoDeployment) {
	d.Time = grd.Time
	d.Version = grd.Version
	if len(d.Version) > 7 {
		d.Version = d.Version[:7]
	}
}
