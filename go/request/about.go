package request

import (
	"fmt"
	"time"
)

// GithubRepoDeployment is used to unmarshal information about a github repository
type GithubRepoDeployment struct {
	Version string    `json:"ref"`
	Time    time.Time `json:"updated_at"`
}

// Deployment contains information about a deployment that is ready to be consumed
type Deployment struct {
	Version string
	Time    time.Time
}

// PreviousDeployment returns some information about the most recent deployment
func PreviousDeployment() (Deployment, error) {
	owner := "jacobpatterson1549"
	repo := "nate-mlb"
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/deployments", owner, repo)
	var grd []GithubRepoDeployment
	err := request.structPointerFromURL(url, &grd)

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
