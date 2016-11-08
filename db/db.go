package db

import (
	"database/sql"
	"fmt"

	"github.com/Sirupsen/logrus"
)

// Database is a wrapper for the app database connection.
type Database struct {
	*sql.DB
}

// Config contains database connection parameters.
type Config struct {
	host     string `yaml:"db_host"`
	name     string `yaml:"db_name"`
	user     string `yaml:"db_user"`
	password string `yaml:"db_password"`
}

// New takes a database configuration and returns a Database object, but any error
// during the initialization process will be deemed fatal.
func New(c *Config) *Database {
	dbInfo := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=disable", c.user, c.name, c.host)
	if c.password != "" {
		dbInfo = fmt.Sprintf("password=%s %s", c.password, dbInfo)
	}

	ppdb, err := sql.Open("postgres", dbInfo)
	if err != nil {
		logrus.Fatal(err.Error())
	}
	ppdb.SetMaxIdleConns(100)

	if err = ppdb.Ping(); err != nil {
		logrus.Fatal(err.Error())
	}

	return &Database{ppdb}
}

// CreateUser uses the given session profile to create a user record
// and returns the newly created user's ID.
func (db *Database) CreateUser(email, password, fname, lname string) (int, error) {
	query := `INSERT INTO account (
				email, password, first_name, last_name
			) VALUES (
				$1, $2, $3, $4
			) RETURNING id`

	var userID int
	err := db.QueryRow(query, email, password, fname, lname).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// GetUserID returns the user ID for a given email.
func (db *Database) GetUserID(email string) (int, error) {
	var userID int
	query := `SELECT id
			FROM account
			WHERE email = $1`
	if err := db.QueryRow(query, email).Scan(&userID); err != nil {
		return 0, err
	}
	return userID, nil
}
