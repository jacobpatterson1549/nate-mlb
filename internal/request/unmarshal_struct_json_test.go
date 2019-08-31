package request

import (
	"encoding/json"
	"testing"
)

type unmarshalStructJSONTest struct {
	interfaceJSON string
	got           interface{}
	want          interface{}
}

var unmarshalStructJSONTests = []unmarshalStructJSONTest{
	{
		interfaceJSON: `{"ref":"f370c06649a740542504b7ecb18031908f394fe8","updated_at":"2019-08-29T17:29:35Z"}`,
		got:           new(GithubRepoDeployment),
		want: &GithubRepoDeployment{
			Version: "f370c06649a740542504b7ecb18031908f394fe8",
			Time:    "2019-08-29T17:29:35Z", // TODO: can this deserialize to time.RFC3339 ?
		},
	},
}

func TestUnmarshalStructJson(t *testing.T) {
	for i, test := range unmarshalStructJSONTests {
		err1 := json.Unmarshal([]byte(test.interfaceJSON), &test.got)
		gotJSON, err2 := json.Marshal(test.got)
		wantJSON, err3 := json.Marshal(test.want)
		if err1 != nil && err2 != nil || err3 != nil {
			t.Errorf("Test %v:\n%v/%v/%v", i, err1, err2, err3)
		} else if string(wantJSON) != string(gotJSON) {
			t.Errorf("Test %v:\nwanted   %+v\nbut got  %+v\nfor json %v", i, test.want, test.got, test.interfaceJSON)
		}
	}
}
