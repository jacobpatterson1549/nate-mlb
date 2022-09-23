package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

type (
	// Password is a string that can be validated
	Password             string
	bcryptPasswordHasher struct{}
	passwordHasher       interface {
		hash(p Password) (string, error)
		isCorrect(p Password, hashedPassword string) (bool, error)
	}
)

var whitespaceRE = regexp.MustCompile(`\s`)

// getUserPassword gets the password for the specified user
func (ds Datastore) getUserPassword(username string) (string, error) {
	return ds.db.GetUserPassword(username)
}

func (d sqlDB) GetUserPassword(username string) (string, error) {
	sqlFunction := newReadSQLFunction("get_user_password", []string{"password"}, username)
	row := d.db.QueryRow(sqlFunction.sql(), sqlFunction.args...)
	var password string
	err := row.Scan(&password)
	if err != nil {
		return password, fmt.Errorf("getting password for user %v: %w", username, err)
	}
	return password, nil
}

// SetUserPassword gets the password for the specified user
func (ds Datastore) SetUserPassword(username string, p Password) error {
	if err := p.validate(); err != nil {
		return err
	}
	hashedPassword, err := ds.ph.hash(p)
	if err != nil {
		return err
	}
	return ds.db.SetUserPassword(username, hashedPassword)
}

func (d *sqlDB) SetUserPassword(username, hashedPassword string) error {
	sqlFunction := newWriteSQLFunction("set_user_password", username, hashedPassword)
	result, err := d.db.Exec(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return fmt.Errorf("setting user password: %w", err)
	}
	return expectSingleRowAffected(result)
}

// AddUser creates a user with the specified username and password
func (ds Datastore) AddUser(username string, p Password) error {
	if err := p.validate(); err != nil {
		return err
	}
	hashedPassword, err := ds.ph.hash(p)
	if err != nil {
		return err
	}
	return ds.db.AddUser(username, hashedPassword)
}

func (d *sqlDB) AddUser(username, hashedPassword string) error {
	sqlFunction := newWriteSQLFunction("add_user", username, hashedPassword)
	result, err := d.db.Exec(sqlFunction.sql(), sqlFunction.args...)
	if err != nil {
		return fmt.Errorf("adding user: %w", err)
	}
	return expectSingleRowAffected(result)
}

// IsCorrectUserPassword determines whether the password for the user is correct
func (ds Datastore) IsCorrectUserPassword(username string, p Password) (bool, error) {
	hashedPassword, err := ds.getUserPassword(username)
	if err != nil {
		return false, err
	}
	return ds.ph.isCorrect(p, hashedPassword)
}

// SetAdminPassword sets the admin password
// If the admin user does not exist, it is created.
func (ds Datastore) SetAdminPassword(p Password) error {
	username := "admin"
	_, err := ds.getUserPassword(username)
	switch {
	case err == nil: // user exists
		return ds.SetUserPassword(username, p)
	case ds.db.IsNotExist(err):
		return ds.AddUser(username, p)
	default: // problem checking if user exists
		return err
	}
}

func (d sqlDB) IsNotExist(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (p Password) validate() error {
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
