package request

import (
	"fmt"
	"time"
)

type (
	// AboutRequester gets information about the most recent Deployment of the application
	AboutRequester struct {
		environment string
		requester   requester
	}

	// GithubRepoDeployment is used to unmarshal information about a github repository
	GithubRepoDeployment struct {
		Version     string    `json:"ref"`
		Time        time.Time `json:"updated_at"`
		Environment string    `json:"environment"`
	}

	githubRepoDeployments []GithubRepoDeployment

	// Deployment contains information about a deployment that is ready to be consumed
	Deployment struct {
		Version string
		Time    time.Time
	}
)

// PreviousDeployment returns some information about the most recent deployment
func (r AboutRequester) PreviousDeployment() (*Deployment, error) {
	owner := "jacobpatterson1549"
	repo := "nate-mlb"
	uri := fmt.Sprintf("https://api.github.com/repos/%s/%s/deployments", owner, repo)
	var s githubRepoDeployments
	err := r.requester.structPointerFromURI(uri, &s)
	if err != nil {
		return nil, err
	}
	return s.previousDeployment(r.environment), nil
}

func (grds githubRepoDeployments) previousDeployment(environment string) *Deployment {
	for _, grd := range grds {
		if grd.Environment == environment {
			return grd.deployment()
		}
	}
	return nil
}

func (grd GithubRepoDeployment) deployment() *Deployment {
	time := grd.Time
	version := grd.Version
	if len(version) > 7 {
		version = version[:7]
	}
	return &Deployment{
		Time:    time,
		Version: version,
	}
}
