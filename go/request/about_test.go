package request

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type setDeploymentFromGithubRepoDeploymentsTest struct {
	environment string
	grds        []GithubRepoDeployment
	want        *Deployment
}

var setDeploymentFromGithubRepoDeploymentsTests = []setDeploymentFromGithubRepoDeploymentsTest{
	{
		// expect empty deployment when there are no deployments
	},
	{
		// one deployment
		grds: []GithubRepoDeployment{
			{Version: "xyz", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)},
		},
		want: &Deployment{Version: "xyz", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)},
	},
	{
		// two deployments
		grds: []GithubRepoDeployment{
			{Version: "v2", Time: time.Date(2019, time.August, 6, 14, 36, 9, 0, time.UTC)},
			{Version: "v1", Time: time.Date(2019, time.July, 31, 16, 00, 45, 0, time.UTC)},
		},
		want: &Deployment{Version: "v2", Time: time.Date(2019, time.August, 6, 14, 36, 9, 0, time.UTC)},
	},
	{
		// different environments
		environment: "old",
		grds: []GithubRepoDeployment{
			{Version: "v2", Time: time.Date(2019, time.August, 6, 14, 36, 9, 0, time.UTC), Environment: "new"},
			{Version: "v1", Time: time.Date(2019, time.July, 31, 16, 00, 45, 0, time.UTC), Environment: "old"},
		},
		want: &Deployment{Version: "v1", Time: time.Date(2019, time.July, 31, 16, 00, 45, 0, time.UTC)},
	},
	{
		// now for environment
		environment: "old",
		grds: []GithubRepoDeployment{
			{Version: "v2", Time: time.Date(2019, time.August, 6, 14, 36, 9, 0, time.UTC), Environment: "new"},
		},
		want: nil,
	},
	{
		// long version should be truncated like it is on github
		grds: []GithubRepoDeployment{
			{Version: "1234567890", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)},
		},
		want: &Deployment{Version: "1234567", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)},
	},
}

func TestSetDeploymentFromGithubRepoDeployments(t *testing.T) {
	for i, test := range setDeploymentFromGithubRepoDeploymentsTests {
		grds := githubRepoDeployments(test.grds)
		got := grds.previousDeployment(test.environment)
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("Test %v:\nwanted  %v,\nbut got %v", i, test.want, got)
		}
	}
}

func TestPreviousDeployment_RequesterError(t *testing.T) {
	m := mockRequester{
		structPointerFromURIFunc: func(uri string, v interface{}) error {
			return errors.New("requesterError")
		},
	}
	about := AboutRequester{requester: &m}
	_, err := about.PreviousDeployment()
	if err == nil {
		t.Error("expected request to fail, but did not")
	}
}

func TestPreviousDeployment_ok(t *testing.T) {
	jsonFunc := func(uri string) string {
		return `[{"ref":"1234567890","updated_at":"2019-09-19T17:45:08Z","environment":"foo"}]`
	}
	want := Deployment{
		Version: "1234567",
		Time:    time.Date(2019, time.September, 19, 17, 45, 8, 0, time.UTC),
	}
	r := newMockHTTPRequester(jsonFunc)
	about := AboutRequester{environment: "foo", requester: r}
	got, err := about.PreviousDeployment()
	switch {
	case err != nil:
		t.Errorf("request failed: %v", err)
	case !reflect.DeepEqual(&want, got):
		t.Errorf("non-equal Deployments:\nwanted: %v\ngot:    %v", want, got)
	}
}
