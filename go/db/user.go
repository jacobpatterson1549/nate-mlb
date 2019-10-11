package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var ph passwordHasher

type (
	// Password is a string that can be validated
	Password             string
	bcryptPasswordHasher struct{}
	passwordHasher       interface {
		hash(p Password) (string, error)
		isCorrect(p Password, hashedPassword string) (bool, error)
	}
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
func SetUserPassword(username string, p Password) error {
	if err := p.validate(); err != nil {
		return err
	}
	hashedPassword, err := ph.hash(p)
	if err != nil {
		return err
	}
	sqlFunction := newWriteSQLFunction("set_user_password", username, hashedPassword)
	result, err := db.Exec(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return fmt.Errorf("setting user password: %w", err)
	}
	return expectSingleRowAffected(result)
}

// AddUser creates a user with the specified username and password
func AddUser(username string, p Password) error {
	if err := p.validate(); err != nil {
		return err
	}
	hashedPassword, err := ph.hash(p)
	if err != nil {
		return err
	}
	sqlFunction := newWriteSQLFunction("add_user", username, hashedPassword)
	result, err := db.Exec(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return fmt.Errorf("adding user: %w", err)
	}
	return expectSingleRowAffected(result)
}

// IsCorrectUserPassword determines whether the password for the user is correct
func IsCorrectUserPassword(username string, p Password) (bool, error) {
	return isCorrectUserPassword(username, p, getUserPassword)
}

func isCorrectUserPassword(
	username string, p Password,
	getUserPasswordFunc func(string) (string, error)) (bool, error) {
	hashedPassword, err := getUserPasswordFunc(username)
	if err != nil {
		return false, err
	}
	return ph.isCorrect(p, hashedPassword)
}

// SetAdminPassword sets the admin password
// If the admin user does not exist, it is created.
func SetAdminPassword(p Password) error {
	return setAdminPassword(p, getUserPassword, SetUserPassword, AddUser)
}

func setAdminPassword(p Password,
	getUserPasswordFunc func(string) (string, error),
	setUserPasswordFunc func(string, Password) error,
	addUserFunc func(string, Password) error) error {
	username := "admin"
	_, err := getUserPasswordFunc(username)
	switch {
	case err == nil:
		return setUserPasswordFunc(username, p)
	case !errors.Is(err, sql.ErrNoRows):
		return err
	default:
		return addUserFunc(username, p)
	}
}

func (p Password) validate() error {
	whitespaceRE := regexp.MustCompile("\\s") // TODO: cache this
	switch {
	case len(p) == 0:
		return fmt.Errorf("password must be nonempty")
	case whitespaceRE.MatchString(string(p)):
		return fmt.Errorf("password must not contain whitespace")
	}
	return nil
}

func (bcryptPasswordHasher) isCorrect(p Password, hashedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(p))
	switch {
	case err == bcrypt.ErrMismatchedHashAndPassword:
		return false, nil
	case err != nil:
		return false, fmt.Errorf("checking password: %w", err)
	}
	return true, nil
}

func (bcryptPasswordHasher) hash(p Password) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		err = fmt.Errorf("hashing password: %w", err)
	}
	return string(hashedPassword), err
}
