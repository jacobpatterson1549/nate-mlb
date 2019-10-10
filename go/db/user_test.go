package db

import (
	"errors"
	"testing"
)

func TestGetUserPassword(t *testing.T) {
	type userPasswordQueryRow struct {
		Password string
	}
	getUserPasswordTests := []struct {
		username    string
		row         userPasswordQueryRow
		queryRowErr error
	}{
		// {},
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
		db = mockDatabase{
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
		}
		gotPassword, gotErr := getUserPassword(test.username)
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
