package db

import "testing"

func TestPasswordIsValid(t *testing.T) {
	passwordIsValidTests := []struct {
		password password
		want     bool
	}{
		{
			password: "okPassword123",
			want:     true,
		},
		{
			password: "",
			want:     false,
		},
		{
			password: "no spaces are allowed",
			want:     false,
		},
	}
	for i, test := range passwordIsValidTests {
		got := test.password.isValid()
		if test.want != got {
			t.Errorf("Test %d: wanted '%v', but got '%v' for password.isValid() on '%v'", i, test.want, got, test.password)
		}
	}
}
