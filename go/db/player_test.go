package db

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestGetPlayers(t *testing.T) {
	getPlayersTests := []struct {
		requestSportType SportType
		rowsSportType    SportType
		queryErr         error
		rows             []interface{}
		wantSlice        []Player
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
				Player{
					ID:           "1",
					PlayerType:   1,
					SourceID:     1,
					FriendID:     "1",
					DisplayOrder: 1,
				},
			},
		},
		{ // happy path
			requestSportType: 3,
			rowsSportType:    3,
			rows: []interface{}{
				Player{
					ID:           "1",
					PlayerType:   1,
					SourceID:     1,
					FriendID:     "1",
					DisplayOrder: 1,
				},
				Player{
					ID:           "17",
					PlayerType:   3,
					SourceID:     6,
					FriendID:     "2",
					DisplayOrder: 3,
				},
				Player{
					ID:           "34",
					PlayerType:   3,
					SourceID:     4000,
					FriendID:     "2",
					DisplayOrder: 2,
				},
			},
			wantSlice: []Player{
				{
					ID:           "1",
					PlayerType:   1,
					SourceID:     1,
					FriendID:     "1",
					DisplayOrder: 1,
				},
				{
					ID:           "17",
					PlayerType:   3,
					SourceID:     6,
					FriendID:     "2",
					DisplayOrder: 3,
				},
				{
					ID:           "34",
					PlayerType:   3,
					SourceID:     4000,
					FriendID:     "2",
					DisplayOrder: 2,
				},
			},
		},
		{ // scan error
			requestSportType: 1,
			rowsSportType:    1,
			rows: []interface{}{
				struct {
					ID           ID
					PlayerType   PlayerType
					SourceID     SourceID
					FriendID     SourceID // should be ID
					DisplayOrder int
				}{
					ID:           "1",
					PlayerType:   1,
					SourceID:     1,
					FriendID:     1,
					DisplayOrder: 1,
				},
			},
			wantErr: true,
		},
	}
	for i, test := range getPlayersTests {
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
		gotSlice, gotErr := ds.GetPlayers(test.requestSportType)
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

func TestSavePlayers(t *testing.T) {
	savePlayersTests := []struct {
		st                      SportType
		futurePlayers           []Player
		previousPlayers         []interface{}
		getPlayersErr           error
		executeInTransactionErr error
		wantErr                 bool
		wantQueryArgs           [][]interface{}
	}{
		{},
		{ // happy path
			st: 3,
			futurePlayers: []Player{
				{
					ID:           "29",
					PlayerType:   1,
					SourceID:     9,
					FriendID:     "7",
					DisplayOrder: 2,
				},
				{
					ID:           "97",
					PlayerType:   1,
					SourceID:     81,
					FriendID:     "7",
					DisplayOrder: 1,
				},
				{ // not in previous implies new: id is ignored
					ID:           "66",
					PlayerType:   3,
					SourceID:     477,
					FriendID:     "4",
					DisplayOrder: 1,
				},
				{
					ID:           "63",
					PlayerType:   3,
					SourceID:     13,
					FriendID:     "3",
					DisplayOrder: 1,
				},
			},
			previousPlayers: []interface{}{
				Player{
					ID:           "29",
					PlayerType:   1,
					SourceID:     9,
					FriendID:     "7",
					DisplayOrder: 1,
				},
				Player{
					ID:           "97",
					PlayerType:   1,
					SourceID:     81,
					FriendID:     "7",
					DisplayOrder: 2,
				},
				Player{
					ID:           "14",
					PlayerType:   1,
					SourceID:     13,
					FriendID:     "4",
					DisplayOrder: 1,
				},
				Player{
					ID:           "63",
					PlayerType:   3,
					SourceID:     13,
					FriendID:     "3",
					DisplayOrder: 1,
				},
			},
			wantQueryArgs: [][]interface{}{
				{ID("14"), SportType(3)},
				{1, PlayerType(3), SourceID(477), ID("4"), SportType(3)},
				{2, ID("29"), SportType(3)},
				{1, ID("97"), SportType(3)},
			},
		},
		{
			getPlayersErr: errors.New("getPlayers error"),
		},
		{
			executeInTransactionErr: errors.New("executeInTransaction error"),
		},
		{ // playerType is for wrong SportType
			st: 3,
			futurePlayers: []Player{
				{
					PlayerType:   8,
					SourceID:     87,
					FriendID:     "3",
					DisplayOrder: 1,
				},
			},
			wantQueryArgs: [][]interface{}{
				{1, PlayerType(8), SourceID(87), ID("3")},
			},
			wantErr: true,
		},
	}
	playerTypes := PlayerTypeMap{
		PlayerType(1): PlayerTypeInfo{SportType: SportType(3)},
		PlayerType(3): PlayerTypeInfo{SportType: SportType(3)},
		PlayerType(8): PlayerTypeInfo{SportType: SportType(4)},
	}
	for i, test := range savePlayersTests {
		executeInTransactionFunc := func(queries []writeSQLFunction) {
			// delete playerIds, insert players {displayOrder, playerType, sourceID, friendID}, update players {displayOrder, id}
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
					return newMockRows(test.previousPlayers), test.getPlayersErr
				},
				BeginFunc: newMockBeginFunc(test.executeInTransactionErr, executeInTransactionFunc),
			},
			playerTypes: playerTypes,
		}
		wantErr := test.wantErr || test.getPlayersErr != nil || test.executeInTransactionErr != nil
		gotErr := ds.SavePlayers(test.st, test.futurePlayers)
		hadErr := gotErr != nil
		if wantErr != hadErr {
			t.Errorf("Test %v: wanted error %v, got: %v", i, wantErr, gotErr)
		}
		switch {
		case test.getPlayersErr != nil && !errors.Is(gotErr, test.getPlayersErr):
			t.Errorf("Test %v: wanted error to be %v, got %v", i, test.getPlayersErr, gotErr)
		case test.executeInTransactionErr != nil && !errors.Is(gotErr, test.executeInTransactionErr):
			t.Errorf("Test %v: wanted error to be %v, got %v", i, test.executeInTransactionErr, gotErr)
		}
	}
}
