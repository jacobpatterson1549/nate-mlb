package db

import "testing"

func TestPasswordValid(t *testing.T) {
	passwordIsValidTests := []struct {
		p       Password
		wantErr bool
	}{
		{
			p: "okPassword123",
		},
		{
			p:       "",
			wantErr: true,
		},
		{
			p:       "no spaces are allowed",
			wantErr: true,
		},
	}
	for i, test := range passwordIsValidTests {
		gotErr := test.p.validate()
		hadErr := gotErr != nil
		if test.wantErr != hadErr {
			t.Errorf("Test %d: wanted error: %v, but got %v for password.validate() on '%v'", i, test.wantErr, gotErr, test.p)
		}
	}
}
