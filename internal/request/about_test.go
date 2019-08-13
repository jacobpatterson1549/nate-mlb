package request

import (
	"testing"
	"time"
)

type setDeplomentFromGithubRepoDeploymentsTest struct {
	grd       []GithubRepoDeployment
	wantError bool
	want      Deployment
}

var setDeplomentFromGithubRepoDeploymentsTests = []setDeplomentFromGithubRepoDeploymentsTest{
	{
		// expect empty deployment when there are no deployments
	},
	{
		// one deployment
		grd: []GithubRepoDeployment{
			GithubRepoDeployment{Version: "xyz", Time: "2019-08-13T02:47:32Z"},
		},
		want: Deployment{Version: "xyz", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)},
	},
	{
		// two deployments
		grd: []GithubRepoDeployment{
			GithubRepoDeployment{Version: "v2", Time: "2019-08-06T14:36:09Z"},
			GithubRepoDeployment{Version: "v1", Time: "2019-07-31T16:00:45Z"},
		},
		want: Deployment{Version: "v2", Time: time.Date(2019, time.August, 6, 14, 36, 9, 0, time.UTC)},
	},
	{
		// two deployments
		grd: []GithubRepoDeployment{
			GithubRepoDeployment{Version: "v2", Time: "INVALID_DATE"},
			GithubRepoDeployment{Version: "v1", Time: "2019-07-31T16:00:45Z"},
		},
		wantError: false,
	},
}

func TestSetDeplomentFromGithubRepoDeploymentsTest(t *testing.T) {
	for i, test := range setDeplomentFromGithubRepoDeploymentsTests {
		var got Deployment
		err := got.setFromGithubRepoDeployments(test.grd)
		hadError := err != nil
		if test.wantError != hadError {
			t.Errorf("Test %v: wanted %v, but got ERROR %v", i, test.want, err)
		}
		if !test.wantError && got != test.want {
			t.Errorf("Test %v: wanted %v, but got %v", i, test.want, got)
		}
	}
}
