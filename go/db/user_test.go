package db

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
)

type mockPasswordHasher struct {
	isCorrectFunc func(p Password, hashedPassword string) (bool, error)
	hashFunc      func(p Password) (string, error)
}

func (m mockPasswordHasher) isCorrect(p Password, hashedPassword string) (bool, error) {
	return m.isCorrectFunc(p, hashedPassword)
}

func (m mockPasswordHasher) hash(p Password) (string, error) {
	return m.hashFunc(p)
}

func TestGetUserPassword(t *testing.T) {
	type userPasswordQueryRow struct {
		Password string
	}
	getUserPasswordTests := []struct {
		username    string
		row         userPasswordQueryRow
		queryRowErr error
	}{
		{},
		{
			row: userPasswordQueryRow{
				Password: "voodoo_cookie73",
			},
		},
		{
			queryRowErr: errors.New("scan error"),
		},
	}
	for i, test := range getUserPasswordTests {
		ds := Datastore{db: &sqlDB{
			db: mockDatabase{
				QueryRowFunc: func(query string, args ...interface{}) row {
					return mockRow{
						ScanFunc: func(dest ...interface{}) error {
							switch {
							case test.queryRowErr != nil:
								return test.queryRowErr
							default:
								return mockRowScanFunc(test.row, dest...)
							}
						},
					}
				},
			}},
		}
		gotPassword, gotErr := ds.getUserPassword(test.username)
		switch {
		case gotErr != nil:
			if !errors.Is(gotErr, test.queryRowErr) {
				t.Errorf("Test %v: wanted error to have %v; got %v", i, test.queryRowErr, gotErr)
			}
		default:
			if gotPassword != test.row.Password {
				t.Errorf("Test %v: wanted %v; got %v", i, test.row.Password, gotPassword)
			}
		}
	}
}

func TestSetUserPassword(t *testing.T) {
	userExecuteHelperTest(t, Datastore.SetUserPassword)
}
func TestAddUser(t *testing.T) {
	userExecuteHelperTest(t, Datastore.AddUser)
}

func userExecuteHelperTest(t *testing.T, testFunc func(Datastore, string, Password) error) {
	userExecuteTests := []struct {
		username     string
		p            Password
		hashErr      error
		execErr      error
		rowsAffected int64
		wantErr      bool
	}{
		{ // empty password
			wantErr: true,
		},
		{
			p:       "s3cr3t!",
			hashErr: errors.New("hash error"),
			wantErr: true,
		},
		{
			p:       "s3cr3t!",
			execErr: errors.New("exec error"),
			wantErr: true,
		},
		{ // no users with username
			p:            "s3cr3t!",
			rowsAffected: 0,
			wantErr:      true,
		},
		{ // happy path
			p:            "s3cr3t!",
			rowsAffected: 1,
		},
	}
	for i, test := range userExecuteTests {
		ds := Datastore{db: &sqlDB{
			db: mockDatabase{
				ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
					if test.execErr != nil {
						return nil, test.execErr
					}
					return mockResult{
						RowsAffectedFunc: func() (int64, error) {
							return test.rowsAffected, nil
						},
					}, nil
				},
			}},
			ph: mockPasswordHasher{
				hashFunc: func(p Password) (string, error) {
					if test.hashErr != nil {
						return "", test.hashErr
					}
					return string(test.p) + "-hashed!", nil
				},
			},
		}
		gotErr := testFunc(ds, test.username, test.p)
		switch {
		case test.wantErr:
			switch {
			case gotErr == nil:
				t.Errorf("Test %v: expected error", i)
			case test.hashErr != nil && !errors.Is(gotErr, test.hashErr):
				t.Errorf("Test %v, wanted error with %v, but got %v", i, test.hashErr, gotErr)
			case test.execErr != nil && !errors.Is(gotErr, test.execErr):
				t.Errorf("Test %v, wanted error with %v, but got %v", i, test.execErr, gotErr)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		}
	}
}

func TestIsCorrectUserPassword(t *testing.T) {
	isCorrectUserPasswordTests := []struct {
		username                     string
		p                            Password
		getUserPasswordFuncErr       error
		passwordHandlerIsCorrectBool bool
		passwordHandlerIsCorrectErr  error
		wantErr                      bool
	}{
		{},
		{
			getUserPasswordFuncErr: errors.New("getUserPasswordFuncErr error"),
			wantErr:                true,
		},
		{
			passwordHandlerIsCorrectErr: errors.New("getUserPasswordFuncErr error"),
			wantErr:                     true,
		},
		{
			passwordHandlerIsCorrectBool: true,
		},
	}
	for i, test := range isCorrectUserPasswordTests {
		ds := Datastore{
			db: &sqlDB{db: mockDatabase{
				QueryRowFunc: func(query string, args ...interface{}) row {
					return mockRow{
						ScanFunc: func(dest ...interface{}) error {
							if test.getUserPasswordFuncErr != nil {
								return test.getUserPasswordFuncErr
							}
							p := struct {
								Password string
							}{
								Password: string(test.p),
							}
							return mockRowScanFunc(p, dest...)
						},
					}
				},
			}},
			ph: mockPasswordHasher{
				isCorrectFunc: func(p Password, hashedPassword string) (bool, error) {
					return test.passwordHandlerIsCorrectBool, test.passwordHandlerIsCorrectErr
				},
			},
		}
		gotBool, gotErr := ds.IsCorrectUserPassword(test.username, test.p)
		switch {
		case test.wantErr:
			switch {
			case gotErr == nil:
				t.Errorf("Test %v: expected error", i)
			case test.getUserPasswordFuncErr != nil && !errors.Is(gotErr, test.getUserPasswordFuncErr):
				t.Errorf("Test %v, wanted error with %v, but got %v", i, test.getUserPasswordFuncErr, gotErr)
			case test.passwordHandlerIsCorrectErr != nil && !errors.Is(gotErr, test.passwordHandlerIsCorrectErr):
				t.Errorf("Test %v, wanted error with %v, but got %v", i, test.passwordHandlerIsCorrectErr, gotErr)
			}
		default:
			if test.passwordHandlerIsCorrectBool != gotBool {
				t.Errorf("Test %v: wanted %v, got %v", i, test.passwordHandlerIsCorrectBool, gotBool)
			}
		}
	}
}

func TestSetAdminPassword(t *testing.T) {
	setAdminPasswordTests := []struct {
		p                      Password
		getUserPasswordFuncErr error
		setUserPasswordFuncErr error
		addUserFuncErr         error
	}{
		{
			p: "valid_password17",
		}, // user exists, password successfully set
		{
			p:                      "valid_password17",
			setUserPasswordFuncErr: errors.New("setUserPassword error"),
		},
		{ // user new, password successfully set
			p:                      "valid_password17",
			getUserPasswordFuncErr: sql.ErrNoRows,
		},
		{
			p:                      "valid_password17",
			getUserPasswordFuncErr: sql.ErrNoRows,
			addUserFuncErr:         errors.New("addUser error"),
		},
		{
			p:                      "valid_password17",
			getUserPasswordFuncErr: errors.New("getUserPassword error"),
		},
	}
	for i, test := range setAdminPasswordTests {
		ds := Datastore{
			db: &sqlDB{db: mockDatabase{
				QueryRowFunc: func(query string, args ...interface{}) row {
					return mockRow{
						ScanFunc: func(dest ...interface{}) error {
							if test.getUserPasswordFuncErr != nil {
								return test.getUserPasswordFuncErr
							}
							p := struct {
								Password string
							}{
								Password: string(test.p),
							}
							return mockRowScanFunc(p, dest...)
						},
					}
				},
				ExecFunc: func(query string, args ...interface{}) (sql.Result, error) {
					r := mockResult{
						RowsAffectedFunc: func() (int64, error) {
							return 1, nil
						},
					}
					switch {
					case query[:10] == "SELECT set":
						return r, test.setUserPasswordFuncErr
					case query[:10] == "SELECT add":
						return r, test.addUserFuncErr
					default:
						return nil, fmt.Errorf("Unknown exec query: %v", query)
					}
				},
			}},
			ph: mockPasswordHasher{
				hashFunc: func(p Password) (string, error) {
					return string(p) + "-hashed!", nil
				},
			},
		}
		gotErr := ds.SetAdminPassword(test.p)
		switch {
		case test.getUserPasswordFuncErr == nil:
			if !errors.Is(gotErr, test.setUserPasswordFuncErr) {
				t.Errorf("Test %v: wanted %v, got %v", i, test.setUserPasswordFuncErr, gotErr)
			}
		case test.getUserPasswordFuncErr != sql.ErrNoRows:
			if !errors.Is(gotErr, test.getUserPasswordFuncErr) {
				t.Errorf("Test %v: wanted %v, got %v", i, test.getUserPasswordFuncErr, gotErr)

			}
		default:
			if !errors.Is(gotErr, test.addUserFuncErr) {
				t.Errorf("Test %v: wanted %v, got %v", i, test.addUserFuncErr, gotErr)

			}
		}
	}
}

func TestValidatePassword(t *testing.T) {
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
