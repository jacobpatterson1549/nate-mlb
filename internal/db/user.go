package db

import "fmt"

// GetUserPassword gets the password for the specified user
func GetUserPassword(username string) (string, error) {
	var password string
	row := db.QueryRow(
		`SELECT get_user_password($1)`,
		username)
	err := row.Scan(&password)
	if err != nil {
		return password, fmt.Errorf("problem getting password for user %v: %v", username, err)
	}
	return password, nil
}

// SetUserPassword gets the password for the specified user
func SetUserPassword(username, password string) error {
	row := db.QueryRow(
		`SELECT set_user_password($1, $2)`,
		username,
		password)
	return expectRowFound(row)
}

// AddUser creates a user with the specified username and password
func AddUser(username, password string) error {
	row := db.QueryRow(
		`SELECT add_user($1, $2)`,
		username,
		password)
	return expectRowFound(row)
}
