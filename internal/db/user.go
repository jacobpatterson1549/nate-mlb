package db

import "fmt"

// GetPassword gets the password for the specified user
func GetPassword(username string) (string, error) {
	var v string
	row := db.QueryRow(
		`SELECT password
		FROM users
		WHERE username = $1`,
		username)
	err := row.Scan(&v)
	if err != nil {
		return v, fmt.Errorf("problem getting password for user %v: %v", username, err)
	}
	return v, nil
}

// SavePassword gets the password for the specified user
func SavePassword(username, password string) error {
	result, err := db.Exec(
		`UPDATE users
		SET password = $1
		WHERE username = $2`,
		password, username)
	if err != nil {
		return fmt.Errorf("problem updating password for user %v: %v", username, err)
	}
	return expectSingleRowAffected(result)
}
