package db

import "fmt"

// GetUserPassword gets the password for the specified user
func GetUserPassword(username string) (string, error) {
	sqlFunction := newReadSQLFunction("get_user_password", []string{"password"}, username)
	row := db.QueryRow(sqlFunction.sql(), sqlFunction.args...)
	var password string
	err := row.Scan(&password)
	if err != nil {
		return password, fmt.Errorf("problem getting password for user %v: %v", username, err)
	}
	return password, nil
}

// GetUserExists gets whether or not a user exists with the specified username
func GetUserExists(username string) (bool, error) {
	sqlFunction := newReadSQLFunction("get_user_exists", []string{"username_exists"}, username)
	row := db.QueryRow(sqlFunction.sql(), sqlFunction.args...)
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return exists, fmt.Errorf("problem determining if user %v exists: %v", username, err)
	}
	return exists, nil
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
