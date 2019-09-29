package db

import (
	"os"
	"reflect"
	"testing"
	"time"
)

type mockFileInfo struct {
	NameFunc    func() string
	SizeFunc    func() int64
	ModeFunc    func() os.FileMode
	ModTimeFunc func() time.Time
	IsDirFunc   func() bool
	SysFunc     func() interface{}
}

func (m mockFileInfo) Name() string {
	return m.NameFunc()
}
func (m mockFileInfo) Size() int64 {
	return m.SizeFunc()
}
func (m mockFileInfo) Mode() os.FileMode {
	return m.ModeFunc()
}
func (m mockFileInfo) ModTime() time.Time {
	return m.ModTimeFunc()
}
func (m mockFileInfo) IsDir() bool {
	return m.IsDirFunc()
}
func (m mockFileInfo) Sys() interface{} {
	return m.SysFunc()
}

/*
getSetupFileContents = func(filename string) ([]byte, error) {
	return []byte{}, nil
}
getSetupFunctionDirContents = func(dirname string) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}
*/

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

func TestLimitPlayerTypes(t *testing.T) {
	limitPlayerTypesTests := []struct {
		initialPlayerTypes map[PlayerType]playerType
		initialSportTypes  map[SportType]sportType
		playerTypesCsv     string
		wantErr            bool
		wantPlayerTypes    map[PlayerType]playerType
		wantSportTypes     map[SportType]sportType
	}{
		{ // no playerTypes ok
		},
		{ // bad playerTypesCsv
			playerTypesCsv: "one",
			wantErr:        true,
		},
		{ // no playerTypes
			playerTypesCsv: "1",
			wantErr:        true,
		},
		{ // wanted playerType that is not loaded
			initialPlayerTypes: map[PlayerType]playerType{1: {}},
			playerTypesCsv:     "2",
			wantErr:            true,
		},
		{ // no filter
			initialPlayerTypes: map[PlayerType]playerType{1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}},
			initialSportTypes:  map[SportType]sportType{1: {}, 2: {}},
			playerTypesCsv:     "",
			wantPlayerTypes:    map[PlayerType]playerType{1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}},
			wantSportTypes:     map[SportType]sportType{1: {}, 2: {}},
		},
		{ // filter to one playerType
			initialPlayerTypes: map[PlayerType]playerType{1: {sportType: 1}, 2: {sportType: 1}, 3: {sportType: 1}, 4: {sportType: 2}, 5: {sportType: 2}, 6: {sportType: 2}},
			initialSportTypes:  map[SportType]sportType{1: {}, 2: {}},
			playerTypesCsv:     "4",
			wantPlayerTypes:    map[PlayerType]playerType{4: {sportType: 2}},
			wantSportTypes:     map[SportType]sportType{2: {}},
		},
	}
	for i, test := range limitPlayerTypesTests {
		playerTypes = test.initialPlayerTypes
		sportTypes = test.initialSportTypes
		err := LimitPlayerTypes(test.playerTypesCsv)
		switch {
		case test.wantErr:
			if err == nil {
				t.Errorf("Test %v: wanted error, but did not get one", i)
			}
		case err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		default:
			switch {
			case !reflect.DeepEqual(test.wantPlayerTypes, playerTypes):
				t.Errorf("Test %v: playerTypes:\nwanted: %v\ngot:    %v", i, test.wantPlayerTypes, playerTypes)
			case !reflect.DeepEqual(test.wantSportTypes, sportTypes):
				t.Errorf("Test %v: sportTypes:\nwanted: %v\ngot:    %v", i, test.wantSportTypes, sportTypes)
			}
		}
	}
}
