package db

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestGetFriends(t *testing.T) {
	getFriendsTests := []struct {
		requestSportType SportType
		rowsSportType    SportType
		queryErr         error
		rows             []interface{}
		wantSlice        []Friend
		wantErr          bool
	}{
		{},
		{
			queryErr: fmt.Errorf("query error"),
			wantErr:  true,
		},
		{ // incorrect sportType
			requestSportType: 1,
			rowsSportType:    2,
			rows: []interface{}{
				Friend{
					ID:           1,
					DisplayOrder: 1,
					Name:         "arnold",
				},
			},
		},
		{ // happy path
			requestSportType: 3,
			rowsSportType:    3,
			rows: []interface{}{
				Friend{
					ID:           1,
					DisplayOrder: 1,
					Name:         "alfred",
				},
				Friend{
					ID:           6,
					DisplayOrder: 3,
					Name:         "aaron",
				},
				Friend{
					ID:           4,
					DisplayOrder: 2,
					Name:         "earl",
				},
			},
			wantSlice: []Friend{
				Friend{
					ID:           1,
					DisplayOrder: 1,
					Name:         "alfred",
				},
				Friend{
					ID:           6,
					DisplayOrder: 3,
					Name:         "aaron",
				},
				Friend{
					ID:           4,
					DisplayOrder: 2,
					Name:         "earl",
				},
			},
		},
		{ // scan error
			requestSportType: 1,
			rowsSportType:    1,
			rows: []interface{}{
				struct {
					ID           string
					DisplayOrder int
					Name         string
				}{
					ID:           "1",
					DisplayOrder: 1,
					Name:         "arnold",
				},
			},
			wantErr: true,
		},
	}
	for i, test := range getFriendsTests {
		ds := Datastore{
			db: mockDatabase{
				QueryFunc: func(query string, args ...interface{}) (rows, error) {
					if test.queryErr != nil {
						return nil, test.queryErr
					}
					if test.requestSportType != test.rowsSportType {
						return newMockRows([]interface{}{}), nil
					}
					return newMockRows(test.rows), nil
				},
			},
		}
		gotSlice, gotErr := ds.GetFriends(test.requestSportType)
		switch {
		case test.wantErr:
			if gotErr == nil {
				t.Errorf("Test %v: expected error", i)
			}
		case gotErr != nil:
			t.Errorf("Test %v: unexpected error: %v", i, gotErr)
		case test.requestSportType != test.rowsSportType:
			if len(gotSlice) != 0 {
				t.Errorf("Test %v: expected no rows for incorrect sportType, but got %v", i, gotSlice)
			}
		default:
			if len(test.rows) != len(gotSlice) {
				t.Errorf("Test %v: incorrect output rows", i)
			}
			for j, got := range gotSlice {
				want := test.wantSlice[j]
				if want != got {
					t.Errorf("Test %v, %T %v not equal: want %v, got %v", i, j, want, want, got)
				}
			}
		}
	}
}

func TestSaveFriends(t *testing.T) {
	saveFriendsTests := []struct {
		st                      SportType
		futureFriends           []Friend
		previousFriends         []interface{}
		getFriendsErr           error
		executeInTransactionErr error
		wantQueryArgs           [][]interface{}
	}{
		{},
		{ // happy path
			st: 9,
			futureFriends: []Friend{
				{
					ID:           8,
					DisplayOrder: 2,
					Name:         "bobby",
				},
				{
					DisplayOrder: 1,
					Name:         "new alice",
				},
				{
					ID:           7,
					DisplayOrder: 3,
					Name:         "curt",
				},
				{
					ID:           5,
					DisplayOrder: 4,
					Name:         "jeb",
				},
			},
			previousFriends: []interface{}{
				Friend{
					ID:           1,
					DisplayOrder: 1,
					Name:         "alfred",
				},
				Friend{
					ID:           8,
					DisplayOrder: 3,
					Name:         "bob",
				},
				Friend{
					ID:           7,
					DisplayOrder: 2,
					Name:         "curt",
				},
				Friend{
					ID:           5,
					DisplayOrder: 4,
					Name:         "jeb",
				},
			},
			wantQueryArgs: [][]interface{}{
				{ID(1)}, // alfred
				{1, "new alice", SportType(9)},
				{2, "bobby", ID(8)},
				{3, "curt", ID(7)},
			},
		},
		{
			getFriendsErr: errors.New("getFriends error"),
		},
		{
			executeInTransactionErr: errors.New("executeInTransaction error"),
		},
	}
	for i, test := range saveFriendsTests {
		executeInTransactionFunc := func(queries []writeSQLFunction) {
			// delete friendIds, insert friends {displayOrder, name}, update friends {displayOrder, name, id}
			if len(test.wantQueryArgs) != len(queries) {
				t.Errorf("Test %v: wanted %v queries, got %v", i, len(test.wantQueryArgs), len(queries))
			}
			for j, wantQueryArgs := range test.wantQueryArgs {
				queryArgs := queries[j].args
				if !reflect.DeepEqual(wantQueryArgs, queryArgs) {
					t.Errorf("Test %v: query %v args: wanted %v, got %v", i, j, wantQueryArgs, queryArgs)
				}
			}
		}
		ds := Datastore{
			db: mockDatabase{
				QueryFunc: func(query string, args ...interface{}) (rows, error) {
					if len(args) != 1 || !reflect.DeepEqual(test.st, args[0]) {
						t.Errorf("Test %v: wanted to get friends for SportType %v, but got %v", i, test.st, args)
					}
					return newMockRows(test.previousFriends), test.getFriendsErr
				},
				BeginFunc: newMockBeginFunc(test.executeInTransactionErr, executeInTransactionFunc),
			},
		}
		wantErr := test.getFriendsErr != nil || test.executeInTransactionErr != nil
		gotErr := ds.SaveFriends(test.st, test.futureFriends)
		hadErr := gotErr != nil
		if wantErr != hadErr {
			t.Errorf("Test %v: wanted error %v, got: %v", i, wantErr, gotErr)
		}
		switch {
		case test.getFriendsErr != nil && !errors.Is(gotErr, test.getFriendsErr):
			t.Errorf("Test %v: wanted error to be %v, got %v", i, test.getFriendsErr, gotErr)
		case test.executeInTransactionErr != nil && !errors.Is(gotErr, test.executeInTransactionErr):
			t.Errorf("Test %v: wanted error to be %v, got %v", i, test.executeInTransactionErr, gotErr)
		}
	}
}
