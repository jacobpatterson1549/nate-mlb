package request

import (
	"fmt"
	"time"
)

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

// PreviousDeployment returns some information about the most recent deployment
func PreviousDeployment() (Deployment, error) {
	owner := "jacobpatterson1549"
	repo := "nate-mlb"
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/deployments", owner, repo)
	var grd []GithubRepoDeployment
	err := requestStruct(url, grd)

	var previousDeployment Deployment
	if err != nil {
		return previousDeployment, err
	}
	err = previousDeployment.setFromGithubRepoDeployments(grd)
	return previousDeployment, err
}

func (d *Deployment) setFromGithubRepoDeployments(grd []GithubRepoDeployment) error {
	if len(grd) == 0 {
		return nil
	}
	return d.setFromGithubRepoDeployment(grd[0])
}

func (d *Deployment) setFromGithubRepoDeployment(grd GithubRepoDeployment) error {
	var err error
	d.Time, err = time.Parse(time.RFC3339, grd.Time)
	if err != nil {
		return fmt.Errorf("problem parsing %v into date: %v", grd.Time, err)
	}
	d.Version = grd.Version
	return nil
}
