package db

import "testing"

type passwordIsValidTest struct {
	password password
	want     bool
}

var passwordIsValidTests = []passwordIsValidTest{
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

func TestPasswordIsValid(t *testing.T) {
	for i, test := range passwordIsValidTests {
		got := test.password.isValid()
		if test.want != got {
			t.Errorf("Test %d: wanted '%v', but got '%v' for password.isValid() on '%v'", i, test.want, got, test.password)
		}
	}
}
