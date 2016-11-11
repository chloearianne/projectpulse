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

// GetMyEvents returns the events that the user has marked.
func (db *Database) GetMyEvents(email string) (int, error) {
	//TODO
	return 0, nil
}
