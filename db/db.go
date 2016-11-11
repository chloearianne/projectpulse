package db

import (
	"database/sql"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/chloearianne/protestpulse/session"
)

// Database is a wrapper for the app database connection.
type Database struct {
	*sql.DB
}

// Config contains database connection parameters.
type Config struct {
	Host     string `yaml:"db_host"`
	Name     string `yaml:"db_name"`
	User     string `yaml:"db_user"`
	Password string `yaml:"db_password"`
}

// New takes a database configuration and returns a Database object, but any error
// during the initialization process will be deemed fatal.
func New(c Config) *Database {
	if c.User == "" || c.Name == "" || c.Host == "" {
		logrus.WithFields(logrus.Fields{
			"host": c.Host,
			"name": c.Name,
			"user": c.User,
		}).Fatal("Missing DB configurations parameters")
	}
	dbInfo := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=disable", c.User, c.Name, c.Host)
	if c.Password != "" {
		dbInfo = fmt.Sprintf("password=%s %s", c.Password, dbInfo)
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
func (db *Database) CreateUser(p *session.Profile, password string) (int, error) {
	query := `INSERT INTO account (
				email, password, first_name, last_name
			) VALUES (
				$1, $2, $3, $4
			) RETURNING id`

	var userID int
	err := db.QueryRow(query, p.Email, password, p.GivenName, p.FamilyName).Scan(&userID)
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

// GetMyEvents returns the events that the user has marked.
func (db *Database) GetMyEvents(email string) (int, error) {
	//TODO
	return 0, nil
}
