package db

import "fmt"

// GetPassword gets the password for the specified user
func GetPassword(username string) (string, error) {
	var v string
	row := db.QueryRow(
		`SELECT get_user_password($1)`,
		username)
	err := row.Scan(&v)
	if err != nil {
		return v, fmt.Errorf("problem getting password for user %v: %v", username, err)
	}
	return v, nil
}

// SavePassword gets the password for the specified user
func SavePassword(username, password string) error {
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
