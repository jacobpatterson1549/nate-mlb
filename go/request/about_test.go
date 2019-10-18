package request

import (
	"errors"
	"testing"
	"time"
)

type setDeplomentFromGithubRepoDeploymentsTest struct {
	grd  []GithubRepoDeployment
	want Deployment
}

var setDeplomentFromGithubRepoDeploymentsTests = []setDeplomentFromGithubRepoDeploymentsTest{
	{
		// expect empty deployment when there are no deployments
	},
	{
		// one deployment
		grd: []GithubRepoDeployment{
			{Version: "xyz", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)},
		},
		want: Deployment{Version: "xyz", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)},
	},
	{
		// two deployments
		grd: []GithubRepoDeployment{
			{Version: "v2", Time: time.Date(2019, time.August, 6, 14, 36, 9, 0, time.UTC)},
			{Version: "v1", Time: time.Date(2019, time.July, 31, 16, 00, 45, 0, time.UTC)},
		},
		want: Deployment{Version: "v2", Time: time.Date(2019, time.August, 6, 14, 36, 9, 0, time.UTC)},
	},
	{
		// long version should be truncated like it is on github
		grd: []GithubRepoDeployment{
			{Version: "1234567890", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)},
		},
		want: Deployment{Version: "1234567", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)},
	},
}

func TestSetDeplomentFromGithubRepoDeployments(t *testing.T) {
	for i, test := range setDeplomentFromGithubRepoDeploymentsTests {
		var got Deployment
		got.setFromGithubRepoDeployments(test.grd)
		if got != test.want {
			t.Errorf("Test %v:\nwanted  %v,\nbut got %v", i, test.want, got)
		}
	}
}

func TestPreviousDeployment_RequestorError(t *testing.T) {
	m := mockRequestor{
		structPointerFromURIFunc: func(uri string, v interface{}) error {
			return errors.New("requestorError")
		},
	}
	about := AboutRequestor{requestor: &m}
	_, err := about.PreviousDeployment()
	if err == nil {
		t.Error("expected request to fail, but did not")
	}
}

func TestPreviousDeployment_ok(t *testing.T) {
	jsonFunc := func(uri string) string {
		return `[{"ref":"1234567890","updated_at":"2019-09-19T17:45:08Z"}]`
	}
	want := Deployment{
		Version: "1234567",
		Time:    time.Date(2019, time.September, 19, 17, 45, 8, 0, time.UTC),
	}
	r := newMockHTTPRequestor(jsonFunc)
	about := AboutRequestor{requestor: r}
	got, err := about.PreviousDeployment()
	switch {
	case err != nil:
		t.Errorf("request failed: %v", err)
	case want != got:
		t.Errorf("non-equal Deployments:\nwanted: %v\ngot:    %v", want, got)
	}
}
