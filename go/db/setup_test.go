package db

import (
	"database/sql"
	"errors"
	"fmt"
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

func TestSetupTablesAndFunctions(t *testing.T) {
	setupTablesAndFunctionsTests := []struct {
		getSetupTableQueriesErr    error
		getSetupFunctionQueriesErr error
		beginErr                   error
		execErr                    error
		rollbackErr                error
		commitErr                  error
	}{
		{}, // happy path
		{
			getSetupTableQueriesErr: errors.New("getSetupTableQueries error"),
		},
		{
			getSetupFunctionQueriesErr: errors.New("getSetupFunctionQueries error"),
		},
		{
			beginErr: errors.New("begin error"),
		},
		{
			execErr: errors.New("exec error"),
		},
		{
			execErr:     errors.New("exec error"),
			rollbackErr: errors.New("rollback error"),
		},
		{
			commitErr: errors.New("commit error"),
		},
	}
	for i, test := range setupTablesAndFunctionsTests {
		readFileFunc := func(filename string) ([]byte, error) {
			if test.getSetupTableQueriesErr != nil {
				return nil, test.getSetupTableQueriesErr
			}
			return []byte("1;2;3;4;5;6;7"), nil
		}
		readDirFunc := func(dirname string) ([]os.FileInfo, error) {
			if test.getSetupFunctionQueriesErr != nil {
				return nil, test.getSetupFunctionQueriesErr
			}
			fileInfos := make([]os.FileInfo, 11)
			for i := range fileInfos {
				fileInfos[i] = mockFileInfo{
					NameFunc: func() string {
						return fmt.Sprintf("mock_file_%d", i)
					},
				}
			}
			return fileInfos, nil
		}
		commitCalled := false
		rollbackCalled := false
		execFuncCount := 0
		tx := mockTransaction{
			ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
				if test.execErr != nil {
					return nil, test.execErr
				}
				execFuncCount++
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
		db = mockDatabase{
			BeginFunc: func() (transaction, error) {
				if test.beginErr != nil {
					return nil, test.beginErr
				}
				return tx, nil
			},
		}
		gotErr := setupTablesAndFunctions(readFileFunc, readDirFunc)
		switch {
		case gotErr != nil:
			switch {
			case test.getSetupTableQueriesErr != nil:
				if !errors.Is(gotErr, test.getSetupTableQueriesErr) {
					t.Errorf("Test %v: wanted: %v, got: %v", i, test.getSetupTableQueriesErr, gotErr)
				}
			case test.getSetupFunctionQueriesErr != nil:
				if !errors.Is(gotErr, test.getSetupFunctionQueriesErr) {
					t.Errorf("Test %v: wanted: %v, got: %v", i, test.getSetupFunctionQueriesErr, gotErr)
				}
			case test.beginErr != nil:
				if !errors.Is(gotErr, test.beginErr) {
					t.Errorf("Test %v: wanted: %v, got: %v", i, test.beginErr, gotErr)
				}
			case test.execErr != nil:
				switch {
				case test.rollbackErr == nil && !errors.Is(gotErr, test.execErr):
					t.Errorf("Test %v: wanted: %v, got: %v", i, test.execErr, gotErr)
				case test.rollbackErr != nil:
					if !errors.Is(gotErr, test.rollbackErr) {
						t.Errorf("Test %v: wanted: %v, got: %v", i, test.rollbackErr, gotErr)
					}
					if !rollbackCalled {
						t.Errorf("Test %v: rollback not called", i)
					}
				}
			case test.commitErr != nil:
				if !errors.Is(gotErr, test.commitErr) {
					t.Errorf("Test %v: wanted: %v, got: %v", i, test.commitErr, gotErr)
				}
				if !commitCalled {
					t.Errorf("Test %v: commit not called", i)
				}
				if rollbackCalled {
					t.Errorf("Test %v: rollback called", i)
				}
			default:
				t.Errorf("Test %v: unexpected error: %v", i, gotErr)
			}
		default:
			if !commitCalled {
				t.Errorf("Test %v: commit not called", i)
			}
			if rollbackCalled {
				t.Errorf("Test %v: rollback called", i)
			}
			// 6 setup files, each with 7 queries
			// 5 function folders, each with 11 files
			wantFunctionCount := 6*7 + 5*11
			if wantFunctionCount != execFuncCount { // this will need to be updated every time additional setup query types are added
				t.Errorf("Test %v: wanted %v functions to be executed, got %v", i, wantFunctionCount, execFuncCount)
			}
		}
	}
}

func TestGetSetupFunctionQueries_fileReadErr(t *testing.T) {
	wantErr := errors.New("readFile error")
	readFileFunc := func(filename string) ([]byte, error) {
		return nil, wantErr
	}
	readDirFunc := func(dirname string) ([]os.FileInfo, error) {
		fileInfos := make([]os.FileInfo, 11)
		for i := range fileInfos {
			fileInfos[i] = mockFileInfo{
				NameFunc: func() string {
					return fmt.Sprintf("mock_file_%d", i)
				},
			}
		}
		return fileInfos, nil
	}
	_, gotErr := getSetupFunctionQueries(readFileFunc, readDirFunc)
	if gotErr == nil || !errors.Is(gotErr, wantErr) {
		t.Errorf("want %v, got: %v", wantErr, gotErr)
	}
}

func TestLimitPlayerTypes(t *testing.T) {
	limitPlayerTypesTests := []struct {
		initialPlayerTypes map[PlayerType]PlayerTypeInfo
		initialSportTypes  map[SportType]SportTypeInfo
		playerTypesCsv     string
		wantErr            bool
		wantPlayerTypes    map[PlayerType]PlayerTypeInfo
		wantSportTypes     map[SportType]SportTypeInfo
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
			initialPlayerTypes: map[PlayerType]PlayerTypeInfo{1: {}},
			playerTypesCsv:     "2",
			wantErr:            true,
		},
		{ // no filter
			initialPlayerTypes: map[PlayerType]PlayerTypeInfo{1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}},
			initialSportTypes:  map[SportType]SportTypeInfo{1: {}, 2: {}},
			playerTypesCsv:     "",
			wantPlayerTypes:    map[PlayerType]PlayerTypeInfo{1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}},
			wantSportTypes:     map[SportType]SportTypeInfo{1: {}, 2: {}},
		},
		{ // filter to one playerType
			initialPlayerTypes: map[PlayerType]PlayerTypeInfo{1: {SportType: 1}, 2: {SportType: 1}, 3: {SportType: 1}, 4: {SportType: 2}, 5: {SportType: 2}, 6: {SportType: 2}},
			initialSportTypes:  map[SportType]SportTypeInfo{1: {}, 2: {}},
			playerTypesCsv:     "4",
			wantPlayerTypes:    map[PlayerType]PlayerTypeInfo{4: {SportType: 2}},
			wantSportTypes:     map[SportType]SportTypeInfo{2: {}},
		},
	}
	for i, test := range limitPlayerTypesTests {
		playerTypes := test.initialPlayerTypes
		sportTypes := test.initialSportTypes
		err := LimitPlayerTypes(test.playerTypesCsv, sportTypes, playerTypes)
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
