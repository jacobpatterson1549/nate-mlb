package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
)

type (
	firestoreDatabase struct {
		client *firestore.Client
	}
	firestoreRow         struct{}
	firestoreRows        struct{

	}
	firestoreTransaction struct {
		actions []transactionAction
		client  *firestore.Client
	}
	firestoreResult struct{}

	transactionAction struct {
		name string
		args []interface{}
	}
)

const (
	firestoreDatabaseName   = "nate-mlb-db"
	firestoreContextTimeout = 15 * time.Second
)

// variables to ensure mongo structures implement the correct interfaces
var (
	_ database    = &firestoreDatabase{}
	_ row         = &firestoreRow{}
	_ rows        = &firestoreRows{}
	_ transaction = &firestoreTransaction{}
)

func newFirestoreDatabase(projectID string) (database, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID) // do not timeout context - the client is used by the backend
	if err != nil {
		return nil, fmt.Errorf("creating firestore client: %w", err)
	}
	return &firestoreDatabase{client}, nil
}

func (d *firestoreDatabase) Ping() error {
	return fmt.Errorf("not implemented")
}

func (d *firestoreDatabase) Query(query string, args ...interface{}) (rows, error) {
	// sqlFunction := newReadSQLFunction("get_sport_types", []string{"id", "name", "url"})
	// sqlFunction := newReadSQLFunction("get_player_types", []string{"id", "sport_type_id", "name", "description", "score_type"})
	// sqlFunction := newReadSQLFunction("get_years", []string{"year", "active"}, st)
	// sqlFunction := newReadSQLFunction("get_friends", []string{"id", "display_order", "name"}, st)
	// sqlFunction := newReadSQLFunction("get_players", []string{"id", "player_type_id", "source_id", "friend_id", "display_order"}, st)
	return &firestoreRows{}, fmt.Errorf("TODO: not implemented")
}

func (d *firestoreDatabase) QueryRow(query string, args ...interface{}) row {
	// sqlFunction := newReadSQLFunction("get_user_password", []string{"password"}, username)
	// sqlFunction := newReadSQLFunction("get_stat", []string{"year", "etl_timestamp", "etl_json"}, st)
	return &firestoreRow{} // TODO: not implemented
}

func (d *firestoreDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	// sqlFunction := newWriteSQLFunction("set_user_password", username, hashedPassword) // set password = hashedPassword
	// sqlFunction := newWriteSQLFunction("add_user", username, hashedPassword)
	// sqlFunction := newWriteSQLFunction("clr_stat", st) // set etl_json = '', etl_timestamp = ''
	// sqlFunction := newWriteSQLFunction("set_stat", stat.EtlTimestamp, stat.EtlJSON, stat.SportType, stat.Year)
	return nil, fmt.Errorf("TODO: not implemented")
}

func (d *firestoreDatabase) Begin() (transaction, error) {
	return &firestoreTransaction{}, fmt.Errorf("TODO: not implemented")
}

func (row *firestoreRow) Scan(dest ...interface{}) error {
	return fmt.Errorf("TODO: not implemented")
}

func (rs *firestoreRows) Close() error {
	return fmt.Errorf("TODO: not implemented")
}

func (rs *firestoreRows) Next() bool {
	return false // TODO: not implemented
}

func (rs *firestoreRows) Scan(dest ...interface{}) error {
	return fmt.Errorf("TODO: not implemented")
}

func (t *firestoreTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {

	// queries = append(queries, newWriteSQLFunction("del_friend", deleteFriendID))
	// queries = append(queries, newWriteSQLFunction("add_friend", insertFriend.DisplayOrder, insertFriend.Name, st))
	// queries = append(queries, newWriteSQLFunction("set_friend", updateFriend.DisplayOrder, updateFriend.Name, updateFriend.ID))

	// queries = append(queries, newWriteSQLFunction("del_player", deleteID))
	// queries = append(queries, newWriteSQLFunction("add_player", insertPlayer.DisplayOrder, insertPlayer.PlayerType, insertPlayer.SourceID, insertPlayer.FriendID))
	// queries = append(queries, newWriteSQLFunction("set_player", updatePlayer.DisplayOrder, updatePlayer.ID))

	// queries = append(queries, newWriteSQLFunction("clr_year_active", st))
	// queries = append(queries, newWriteSQLFunction("del_year", st, deleteYear))
	// queries = append(queries, newWriteSQLFunction("add_year", st, insertYear))
	// queries = append(queries, newWriteSQLFunction("set_year_active", st, activeYear))

	// NOOP for setup functions

	a := transactionAction{
		name: query,
		args: args,
	}
	t.actions = append(t.actions, a)
	var r firestoreResult
	return r, nil
}

func (t *firestoreTransaction) Commit() error {
	// return fmt.Errorf("TODO: not implemented")
	return withFirestoreTimeoutContext(context.Background(), func(ctx context.Context) error {
		return t.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
			// for _, a := range t.actions {
			// 	// TODO
			// }
			return nil
		})
	})
}

func (t *firestoreTransaction) Rollback() error {
	return nil
}

func (f firestoreResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (f firestoreResult) RowsAffected() (int64, error) {
	return 1, nil
}

func withFirestoreTimeoutContext(ctx context.Context, f func(ctx context.Context) error) error {
	ctx, cancelFunc := context.WithTimeout(ctx, firestoreContextTimeout)
	defer cancelFunc()
	return f(ctx)
}