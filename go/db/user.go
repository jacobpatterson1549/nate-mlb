package db

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// getUserPassword gets the password for the specified user
func getUserPassword(username string) (string, error) {
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
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	sqlFunction := newWriteSQLFunction("set_user_password", username, hashedPassword)
	row := db.QueryRow(sqlFunction.sql(), sqlFunction.args...)
	return expectRowFound(row)
}

// AddUser creates a user with the specified username and password
func AddUser(username, password string) error {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}
	sqlFunction := newWriteSQLFunction("add_user", username, hashedPassword)
	row := db.QueryRow(sqlFunction.sql(), sqlFunction.args...)
	return expectRowFound(row)
}

// IsCorrectUserPassword determines whether the password for the user is correct
func IsCorrectUserPassword(username, password string) (bool, error) {
	hashedPassword, err := getUserPassword(username)
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	correctPassword := err == nil
	return correctPassword, err
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		err = fmt.Errorf("hashing password: %w", err)
	}
	return string(hashedPassword), err
}
