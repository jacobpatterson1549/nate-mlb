package db

import "fmt"

// GetUserPassword gets the password for the specified user
func GetUserPassword(username string) (string, error) {
	sqlFunction := newReadSQLFunction("get_user_password", []string{"password"}, username)
	row := db.QueryRow(sqlFunction.sql(), sqlFunction.args...)
	var password string
	err := row.Scan(&password)
	if err != nil {
		return password, fmt.Errorf("getting password for user %v: %w", username, err)
	}
	return password, nil
}

// SetUserPassword gets the password for the specified user
func SetUserPassword(username, password string) error {
	sqlFunction := newWriteSQLFunction("set_user_password", username, password)
	row := db.QueryRow(sqlFunction.sql(), sqlFunction.args...)
	return expectRowFound(row)
}

// AddUser creates a user with the specified username and password
func AddUser(username, password string) error {
	sqlFunction := newWriteSQLFunction("add_user", username, password)
	row := db.QueryRow(sqlFunction.sql(), sqlFunction.args...)
	return expectRowFound(row)
}
