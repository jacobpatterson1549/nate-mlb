package server

import "testing"

type getSportTypeNameTest struct {
	url  string
	want string
}

var getSportTypeNameTests = []getSportTypeNameTest{
	{
		url:  "",
		want: "",
	},
	{
		url:  "/",
		want: "",
	},
	{
		url:  "/mlb",
		want: "mlb",
	},
	{
		url:  "/nfl/admin",
		want: "nfl",
	},
	{
		url:  "/admin",
		want: "admin", // not a valid sportName, but 
	},
}

func TestGetSportTypeName(t *testing.T) {
	for i, test := range getSportTypeNameTests {
		got := getSportTypeName(test.url)
		if test.want != got {
			t.Errorf("Test %d: wanted '%v', but got '%v' for url '%v'", i, test.want, got, test.url)
		}
	}
}
