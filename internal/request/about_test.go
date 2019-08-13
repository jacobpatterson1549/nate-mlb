package request

import (
	"testing"
	"time"
)

type toDeploymentTest struct {
	grd       GithubRepoDeployment
	wantError bool
	want      Deployment
}

var toDeploymentTests = []toDeploymentTest{
	{grd: GithubRepoDeployment{}, wantError: true},
	{grd: GithubRepoDeployment{Version: "xyz", Time: "2019-08-13T02:47:32Z"}, want: Deployment{Version: "xyz", Time: time.Date(2019, time.August, 13, 2, 47, 32, 0, time.UTC)}},
}

func TestToDeployment(t *testing.T) {
	for i, test := range toDeploymentTests {
		got, err := test.grd.toDeployment()
		if test.wantError != (err != nil) {
			t.Errorf("Test %v: wanted %v, but got ERROR %v", i, test.want, err)
		}
		if !test.wantError && got != test.want {
			t.Errorf("Test %v: wanted %v, but got %v", i, test.want, got)
		}
	}
}
