package db

import (
	"database/sql"
	"errors"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
)

var mockValidFS = fstest.MapFS{
	"sql/setup/users.pgsql":         &fstest.MapFile{Data: []byte("a")},
	"sql/setup/sport_types.pgsql":   &fstest.MapFile{Data: []byte("b")},
	"sql/setup/stats.pgsql":         &fstest.MapFile{Data: []byte("c")},
	"sql/setup/friends.pgsql":       &fstest.MapFile{Data: []byte("d")},
	"sql/setup/player_types.pgsql":  &fstest.MapFile{Data: []byte("e")},
	"sql/setup/players.pgsql":       &fstest.MapFile{Data: []byte("f")},
	"sql/functions/add/DUMMY.pgsql": &fstest.MapFile{Data: []byte("g")},
}

func TestSetupTablesAndFunctions(t *testing.T) {
	setupTablesAndFunctionsTests := []struct {
		fs          fs.ReadFileFS
		beginErr    error
		execErr     error
		rollbackErr error
		commitErr   error
		wantOk      bool
	}{
		{ // happy path
			fs:     mockValidFS,
			wantOk: true,
		},
		{ // getSetupTableQueries error
			fs: fstest.MapFS{
				"sql/functions/add/DUMMY.pgsql": &fstest.MapFile{Data: []byte("g")},
			},
		},
		{ //  getSetupFunctionQueries error
			fs: fstest.MapFS{
				"sql/setup/users.pgsql":        &fstest.MapFile{Data: []byte("a")},
				"sql/setup/sport_types.pgsql":  &fstest.MapFile{Data: []byte("b")},
				"sql/setup/stats.pgsql":        &fstest.MapFile{Data: []byte("c")},
				"sql/setup/friends.pgsql":      &fstest.MapFile{Data: []byte("d")},
				"sql/setup/player_types.pgsql": &fstest.MapFile{Data: []byte("e")},
				"sql/setup/players.pgsql":      &fstest.MapFile{Data: []byte("f")},
			},
		},
		{
			fs:       mockValidFS,
			beginErr: errors.New("begin error"),
		},
		{
			fs:      mockValidFS,
			execErr: errors.New("exec error"),
		},
		{
			fs:          mockValidFS,
			execErr:     errors.New("exec error"),
			rollbackErr: errors.New("rollback error"),
		},
		{
			fs:        mockValidFS,
			commitErr: errors.New("commit error"),
		},
	}
	for i, test := range setupTablesAndFunctionsTests {
		commitCalled := false
		rollbackCalled := false
		execFuncQueries := ""
		tx := mockTransaction{
			ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
				if test.execErr != nil {
					return nil, test.execErr
				}
				execFuncQueries += query
				return mockResult{
					RowsAffectedFunc: func() (int64, error) {
						return 1, nil
					},
				}, nil
			},
			CommitFunc: func() error {
				commitCalled = true
				return test.commitErr
			},
			RollbackFunc: func() error {
				rollbackCalled = true
				return test.rollbackErr
			},
		}
		db := mockDatabase{
			BeginFunc: func() (transaction, error) {
				if test.beginErr != nil {
					return nil, test.beginErr
				}
				return tx, nil
			},
		}
		ds := Datastore{
			db: db,
			fs: test.fs,
		}
		gotErr := ds.SetupTablesAndFunctions()
		switch {
		case !test.wantOk:
			if gotErr == nil {
				t.Errorf("Test %v: wanted error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unwanted error: %v", i, gotErr)
		default:
			if !commitCalled {
				t.Errorf("Test %v: commit not called", i)
			}
			if rollbackCalled {
				t.Errorf("Test %v: rollback called", i)
			}
			// 6 setup files, (a-f)
			// 1 function file (g)
			wantFuncQueries := "abcdefg"
			if wantFuncQueries != execFuncQueries { // this will need to be updated every time additional setup query types are added
				t.Errorf("Test %v: wanted %v queries, got %v", i, wantFuncQueries, execFuncQueries)
			}
		}
	}
}

// TODO: DELETEME
// func TestGetSetupFunctionQueries_fileReadErr(t *testing.T) {
// 	wantErr := errors.New("readFile error")
// 	readFileFunc := func(filename string) ([]byte, error) {
// 		return nil, wantErr
// 	}
// 	ds := Datastore{
// 		fs: mockFS{
// 			ReadFileFunc: readFileFunc,
// 		},
// 	}
// 	_, gotErr := ds.getSetupFunctionQueries()
// 	if gotErr == nil || !errors.Is(gotErr, wantErr) {
// 		t.Errorf("want %v, got: %v", wantErr, gotErr)
// 	}
// }

func TestLimitPlayerTypes(t *testing.T) {
	limitPlayerTypesTests := []struct {
		initialPlayerTypes PlayerTypeMap
		initialSportTypes  SportTypeMap
		playerTypesCsv     string
		wantErr            bool
		wantPlayerTypes    PlayerTypeMap
		wantSportTypes     SportTypeMap
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
			initialPlayerTypes: PlayerTypeMap{1: {}},
			playerTypesCsv:     "2",
			wantErr:            true,
		},
		{ // no filter
			initialPlayerTypes: PlayerTypeMap{1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}},
			initialSportTypes:  SportTypeMap{1: {}, 2: {}},
			playerTypesCsv:     "",
			wantPlayerTypes:    PlayerTypeMap{1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}},
			wantSportTypes:     SportTypeMap{1: {}, 2: {}},
		},
		{ // filter to one playerType
			initialPlayerTypes: PlayerTypeMap{1: {SportType: 1}, 2: {SportType: 1}, 3: {SportType: 1}, 4: {SportType: 2}, 5: {SportType: 2}, 6: {SportType: 2}},
			initialSportTypes:  SportTypeMap{1: {}, 2: {}},
			playerTypesCsv:     "4",
			wantPlayerTypes:    PlayerTypeMap{4: {SportType: 2}},
			wantSportTypes:     SportTypeMap{2: {}},
		},
	}
	for i, test := range limitPlayerTypesTests {
		ds := Datastore{
			playerTypes: test.initialPlayerTypes,
			sportTypes:  test.initialSportTypes,
		}
		err := ds.LimitPlayerTypes(test.playerTypesCsv)
		switch {
		case test.wantErr:
			if err == nil {
				t.Errorf("Test %v: wanted error, but did not get one", i)
			}
		case err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		default:
			switch {
			case !reflect.DeepEqual(test.wantPlayerTypes, ds.playerTypes):
				t.Errorf("Test %v: playerTypes:\nwanted: %v\ngot:    %v", i, test.wantPlayerTypes, ds.playerTypes)
			case !reflect.DeepEqual(test.wantSportTypes, ds.sportTypes):
				t.Errorf("Test %v: sportTypes:\nwanted: %v\ngot:    %v", i, test.wantSportTypes, ds.sportTypes)
			}
		}
	}
}
