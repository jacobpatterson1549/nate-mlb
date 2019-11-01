package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	pingTests := []struct {
		pingErr error
	}{
		{},
		{
			pingErr: fmt.Errorf("Ping error"),
		},
	}
	for i, test := range pingTests {
		ds := Datastore{
			db: mockDatabase{
				PingFunc: func() error {
					return test.pingErr
				},
			},
		}
		wantErr := test.pingErr
		gotErr := ds.Ping()
		if wantErr != gotErr {
			t.Errorf("Test %v: wanted %v, got %v", i, wantErr, gotErr)
		}
	}
}

func TestGetUtcTime(t *testing.T) {
	ds := Datastore{}
	got := ds.GetUtcTime()
	defaultTime := time.Time{}
	if got.Unix() == defaultTime.Unix() {
		t.Errorf("expected time to not be same as default time, got %v (unix time)", defaultTime.Unix())
	}
}

func TestSportTypes(t *testing.T) {
	sportTypes := SportTypeMap{
		1: {Name: "baseball"},
		2: {URL: "nfl"},
	}
	ds := Datastore{
		sportTypes: sportTypes,
	}
	got := ds.SportTypes()
	if !reflect.DeepEqual(sportTypes, got) {
		t.Errorf("not equal\nwanted: %v\ngot:    %v", sportTypes, got)
	}
}

func TestPlayerTypes(t *testing.T) {
	playerTypes := PlayerTypeMap{
		1: {Name: "team"},
		2: {Description: "players that come up to plate to hit"},
		3: {ScoreType: "wins"},
	}
	ds := Datastore{
		playerTypes: playerTypes,
	}
	got := ds.PlayerTypes()
	if !reflect.DeepEqual(playerTypes, got) {
		t.Errorf("not equal\nwanted: %v\ngot:    %v", playerTypes, got)
	}
}

func TestExecuteInTransaction(t *testing.T) {
	executeInTransactionTests := []struct {
		queries     []writeSQLFunction
		beginErr    error
		execErr     error
		rollbackErr error
		commitErr   error
	}{
		{},
		{
			queries: []writeSQLFunction{
				newWriteSQLFunction("query1"),
				newWriteSQLFunction("query2"),
				newWriteSQLFunction("query3"),
			},
		},
		{
			beginErr: errors.New("begin error"),
		},
		{
			queries: []writeSQLFunction{
				newWriteSQLFunction("query1"),
			},
			execErr: errors.New("exec error"),
		},
		{
			queries: []writeSQLFunction{
				newWriteSQLFunction("query1"),
			},
			execErr:     errors.New("exec error"), // causes rollbackError
			rollbackErr: errors.New("rollback error"),
		},
		{
			queries: []writeSQLFunction{
				newWriteSQLFunction("query1"),
			},
			commitErr: errors.New("commit error"),
		},
	}
	for i, test := range executeInTransactionTests {
		commitCalled := false
		rollbackCalled := false
		tx := mockTransaction{
			ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
				if test.execErr != nil {
					return nil, test.execErr
				}
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
		ds := Datastore{db: db}
		gotErr := ds.executeInTransaction(test.queries)
		switch {
		case gotErr == nil:
			if test.beginErr != nil || test.execErr != nil || test.commitErr != nil {
				t.Errorf("Test %v: did not cause an error, but should have", i)
			}
		case test.beginErr != nil:
			if !errors.Is(gotErr, test.beginErr) {
				t.Errorf("Test %v: wanted error during begin, but got: %v", i, gotErr)
			}
		case test.execErr != nil:
			if test.rollbackErr == nil && !errors.Is(gotErr, test.execErr) {
				t.Errorf("Test %v: wanted error during exec, but got: %v", i, gotErr)
			}
		case test.rollbackErr != nil:
			if !errors.Is(gotErr, test.rollbackErr) {
				t.Errorf("Test %v: wanted error during rollback, but got: %v", i, gotErr)
			}
		case test.commitErr != nil:
			if !errors.Is(gotErr, test.commitErr) {
				t.Errorf("Test %v: wanted error during commit, but got: %v", i, gotErr)
			}
		default:
			t.Errorf("Test %v: unknown error: %v", i, gotErr)
		}
		// commit/rollback checks
		switch {
		case gotErr != nil && test.beginErr == nil && test.commitErr == nil:
			if !rollbackCalled {
				t.Errorf("Test %v: expected rollback", i)
			}
			if commitCalled && test.commitErr == nil {
				t.Errorf("Test %v: unexpected commit", i)
			}
		case gotErr == nil:
			if rollbackCalled {
				t.Errorf("Test %v: unexpected rollback", i)
			}
			if !commitCalled && len(test.queries) != 0 {
				t.Errorf("Test %v: expected commit", i)
			}
			if commitCalled && len(test.queries) == 0 {
				t.Errorf("Test %v: should not have been committed because there are no queries", i)
			}
		}
	}
}

func TestExpectSingleRowAffected(t *testing.T) {
	expectSingleRowAffectedTests := []struct {
		rowsAffectedErr error
		rows            int64
		wantErr         bool
	}{
		{
			rowsAffectedErr: errors.New("could not get rows affected"),
			rows:            1, // red herring :)
			wantErr:         true,
		},
		{
			rows: 1, // desired
		},
		{
			rows:    0,
			wantErr: true,
		},
		{
			rows:    3,
			wantErr: true,
		},
	}
	for i, test := range expectSingleRowAffectedTests {
		r := mockResult{
			RowsAffectedFunc: func() (int64, error) {
				return test.rows, test.rowsAffectedErr
			},
		}
		err := expectSingleRowAffected(r)
		switch {
		case test.wantErr && err == nil:
			t.Errorf("Test %v: expected error", i)
		case !test.wantErr && err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		}
	}
}

func TestExpectRowFound(t *testing.T) {
	expectRowFoundTests := []struct {
		found   bool
		scanErr error
		wantErr bool
	}{
		{
			found: true,
		},
		{
			scanErr: fmt.Errorf("scanError"),
		},
		{
			found:   false,
			wantErr: true,
		},
	}
	for i, test := range expectRowFoundTests {
		r := mockRow{
			ScanFunc: func(dest ...interface{}) error {
				if test.scanErr != nil {
					return test.scanErr
				}
				return mockScan(dest[0], test.found)
			},
		}
		gotErr := expectRowFound(r)
		switch {
		case gotErr == nil && (test.scanErr != nil || test.wantErr):
			t.Errorf("Test %v: expected error", i)
		case gotErr != nil && test.found:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		}
	}
}

func TestWaitForDb_numTries(t *testing.T) {
	waitForDbTests := []struct {
		successfulConnectTry int
		numFibonacciTries    int
		wantError            bool
	}{
		{ // should not fail when not attempted to connect
		},
		{
			successfulConnectTry: 1,
			numFibonacciTries:    1,
		},
		{
			successfulConnectTry: 2,
			numFibonacciTries:    3,
		},
		{
			successfulConnectTry: 4,
			numFibonacciTries:    3,
			wantError:            true,
		},
	}
	for i, test := range waitForDbTests {
		dbCheckCount := 0
		db := mockDatabase{
			PingFunc: func() error {
				dbCheckCount++
				if dbCheckCount != test.successfulConnectTry {
					return errors.New("check failed")
				}
				return nil
			},
		}
		log := log.New(ioutil.Discard, "test", log.LstdFlags)
		ds := Datastore{db: db, log: log}
		sleepFunc := func(waitTime int) {}
		err := ds.waitForDb(sleepFunc, test.numFibonacciTries)
		gotError := err != nil
		if test.wantError != gotError {
			t.Errorf("Test %v: wantedError = %v, gotError = %v", i, test.wantError, gotError)
		}
	}
}

func TestWaitForDb_fibonacci(t *testing.T) {
	wantFibonacciSleepSeconds := []int{0, 1, 1, 2, 3, 5, 8}
	dbCheckCount := 0
	db := mockDatabase{
		PingFunc: func() error {
			dbCheckCount++
			return fmt.Errorf("check failed")
		},
	}
	i := 0
	sleepFunc := func(sleepSeconds int) {
		if wantFibonacciSleepSeconds[i] != sleepSeconds {
			t.Errorf("unexpected %vth wait time: wanted %v, got %v", i, wantFibonacciSleepSeconds[i], sleepSeconds)
		}
		i++
	}
	log := log.New(ioutil.Discard, "test", log.LstdFlags)
	ds := Datastore{db: db, log: log}
	numFibonacciTries := len(wantFibonacciSleepSeconds)
	err := ds.waitForDb(sleepFunc, numFibonacciTries)
	if err == nil {
		t.Error("expected db wait check to error out")
	}
	if numFibonacciTries != i {
		t.Errorf("expected to wait for db to start %v times, got %v", numFibonacciTries, i)
	}
	if numFibonacciTries != dbCheckCount {
		t.Errorf("expected to check the db %v times, got %v", numFibonacciTries, dbCheckCount)
	}
}

var testNewDatastoreDriver *mockDriver

func init() {
	testNewDatastoreDriver = new(mockDriver)
	sql.Register("TestNewDatastore", testNewDatastoreDriver)
}
func TestNewDatastore(t *testing.T) {
	newDatastoreTests := []struct {
		newDatabaseErr             error
		waitForDbErr               error
		waitForDbErrStopIndex      int
		setupTablesAndFunctionsErr error
		getSportTypesErr           error
		getPlayerTypesErr          error
		wantErr                    bool
	}{
		{}, // happy path
		{
			newDatabaseErr: errors.New("newSQLDatabase error"),
			wantErr:        true,
		},
		{
			waitForDbErr:          errors.New("waitForDb error"),
			waitForDbErrStopIndex: 4000,
			wantErr:               true, // 4000 > 5
		},
		{
			waitForDbErr:          errors.New("waitForDb error"),
			waitForDbErrStopIndex: 3,
			wantErr:               false, // 3 < 5
		},
		{
			setupTablesAndFunctionsErr: errors.New("SetupTablesAndFunctions error"),
			wantErr:                    true,
		},
		{
			getSportTypesErr: errors.New("GetSportTypes error"),
			wantErr:          true,
		},
		{
			getPlayerTypesErr: errors.New("GetPlayerTypes error"),
			wantErr:           true,
		},
	}
	for i, test := range newDatastoreTests {
		cfg := datastoreConfig{
			driverName: "TestNewDatastore",
			readFileFunc: func(filename string) ([]byte, error) {
				return nil, test.setupTablesAndFunctionsErr
			},
			readDirFunc: func(dirname string) ([]os.FileInfo, error) {
				return nil, test.setupTablesAndFunctionsErr
			},
			pingFailureSleepFunc: func(sleepSeconds int) { /* NOOP */ },
			numFibonacciTries:    5,
			log:                  log.New(ioutil.Discard, "test", log.LstdFlags),
		}
		if test.newDatabaseErr != nil {
			cfg.driverName = "bad driver name"
		}
		mockDriverConn := mockDriverConn{
			PrepareFunc: func(query string) (driver.Stmt, error) {
				return mockDriverStmt{
					CloseFunc: func() error {
						return nil
					},
					NumInputFunc: func() int {
						return 0
					},
					ExecFunc: func(args []driver.Value) (driver.Result, error) {
						return mockResult{}, nil
					},
					QueryFunc: func(args []driver.Value) (driver.Rows, error) {
						var columns []string
						var srcRows [][]driver.Value
						var queryErr error
						switch {
						case strings.Contains(query, "get_sport_types"):
							columns = []string{"id", "name", "url"}
							srcRows = [][]driver.Value{
								{1, "st_1_name", "st_1_url"},
								{2, "st_2_name", "st_2_url"},
							}
							queryErr = test.getSportTypesErr
						case strings.Contains(query, "get_player_types"):
							columns = []string{"id", "sport_type_id", "name", "description", "score_type"}
							srcRows = [][]driver.Value{
								{1, 1, "pt_1_name", "pt_1_description", "pt_1_score_type"},
								{2, 1, "pt_2_name", "pt_2_description", "pt_2_score_type"},
								{3, 1, "pt_3_name", "pt_3_description", "pt_3_score_type"},
								{4, 2, "pt_4_name", "pt_4_description", "pt_4_score_type"},
								{5, 2, "pt_5_name", "pt_5_description", "pt_5_score_type"},
								{6, 2, "pt_6_name", "pt_6_description", "pt_6_score_type"},
							}
							queryErr = test.getPlayerTypesErr
						default:
							queryErr = fmt.Errorf("unknown query: %v", query)
						}
						i := 0
						return mockDriverRows{
							CloseFunc: func() error {
								return nil
							},
							ColumnsFunc: func() []string {
								return columns
							},
							NextFunc: func(dest []driver.Value) error {
								if i == len(srcRows) {
									return io.EOF
								}
								src := srcRows[i]
								for j := 0; j < len(src); j++ {
									dest[j] = src[j]
								}
								i++
								return nil
							},
						}, queryErr
					},
				}, nil
			},
			BeginFunc: func() (driver.Tx, error) {
				return mockDriverTx{
					CommitFunc: func() error {
						return nil
					},
				}, nil
			},
		}
		pingAttempt := 0
		testNewDatastoreDriver.OpenFunc = func(name string) (driver.Conn, error) {
			pingAttempt++
			switch {
			case test.newDatabaseErr != nil:
				return mockDriverConn, test.newDatabaseErr
			case pingAttempt < test.waitForDbErrStopIndex:
				return mockDriverConn, test.waitForDbErr
			default:
				return mockDriverConn, nil
			}
		}
		ds, err := newDatastore(cfg)
		switch {
		case test.wantErr:
			if err == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case err != nil:
			t.Errorf("Test %v: unexpected error: %v", i, err)
		default:
			if ds == nil {
				t.Errorf("Test %v: expected non-nil Datastore: %v", i, ds)
			}
			wantSportTypes := SportTypeMap{
				1: {Name: "st_1_name", URL: "st_1_url", DisplayOrder: 0},
				2: {Name: "st_2_name", URL: "st_2_url", DisplayOrder: 1},
			}
			gotSportTypes, err := ds.GetSportTypes()
			switch {
			case err != nil:
				t.Errorf("Test %v: %v", i, err)
			case !reflect.DeepEqual(wantSportTypes, gotSportTypes):
				t.Errorf("Test %v:\nwanted: %v\ngot:    %v", i, wantSportTypes, gotSportTypes)
			}
			wantPlayerTypes := PlayerTypeMap{
				1: {SportType: 1, Name: "pt_1_name", Description: "pt_1_description", ScoreType: "pt_1_score_type", DisplayOrder: 0},
				2: {SportType: 1, Name: "pt_2_name", Description: "pt_2_description", ScoreType: "pt_2_score_type", DisplayOrder: 1},
				3: {SportType: 1, Name: "pt_3_name", Description: "pt_3_description", ScoreType: "pt_3_score_type", DisplayOrder: 2},
				4: {SportType: 2, Name: "pt_4_name", Description: "pt_4_description", ScoreType: "pt_4_score_type", DisplayOrder: 3},
				5: {SportType: 2, Name: "pt_5_name", Description: "pt_5_description", ScoreType: "pt_5_score_type", DisplayOrder: 4},
				6: {SportType: 2, Name: "pt_6_name", Description: "pt_6_description", ScoreType: "pt_6_score_type", DisplayOrder: 5},
			}
			gotPlayerTypes, err := ds.GetPlayerTypes()
			switch {
			case err != nil:
				t.Errorf("Test %v: %v", i, err)
			case !reflect.DeepEqual(wantPlayerTypes, gotPlayerTypes):
				t.Errorf("Test %v:\nwanted: %v\ngot:    %v", i, wantPlayerTypes, gotPlayerTypes)
			}
		}
	}
}
