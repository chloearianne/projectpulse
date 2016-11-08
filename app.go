package main

import (
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var templateMap = make(map[string]*template.Template)
var cookieStore *sessions.CookieStore
var ppdb *sql.DB

// AppConfig is a container for all app configuration parameters
// that are to be extracted from the YAML config file.
type AppConfig struct {
	// DB config
	DBUser     string `yaml:"db_user"`
	DBPassword string `yaml:"db_password"`
	DBName     string `yaml:"db_name"`
	DBHost     string `yaml:"db_host"`

	// Security config
	CookieKey string `yaml:"cookie_key"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configFile := fmt.Sprintf("conf/%s.yaml", os.Getenv("ENV"))
	if _, err := os.Stat(configFile); err != nil {
		logrus.WithField("path", configFile).WithError(err).Fatal("Could not find config file")
	}
	logrus.Infof("Using config file: %s", configFile)
	config, err := ioutil.ReadFile(configFile)
	c := AppConfig{}
	err = yaml.Unmarshal([]byte(config), &c)
	if err != nil {
		logrus.Fatal(err)
	}

	dbInfo := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=disable", c.DBUser, c.DBName, c.DBHost)
	if c.DBPassword != "" {
		dbInfo = fmt.Sprintf("password=%s %s", c.DBPassword, dbInfo)
	}
	ppdb, err = sql.Open("postgres", dbInfo)
	if err != nil {
		logrus.Fatal(err.Error())
	}
	ppdb.SetMaxIdleConns(100)
	err = ppdb.Ping()
	if err != nil {
		logrus.Fatal(err.Error())
	}
	defer ppdb.Close()

	cookieStore = sessions.NewCookieStore([]byte(c.CookieKey))
	gob.Register(map[string]interface{}{})

	// Generate templateMap from our 'templates' and 'layouts' directories
	templates, err := filepath.Glob("public/templates/*.tmpl")
	if err != nil {
		logrus.Fatal(err)
	}
	layouts, err := filepath.Glob("public/layouts/*.tmpl")
	if err != nil {
		logrus.Fatal(err)
	}
	for _, tmpl := range templates {
		files := append(layouts, tmpl)
		templateMap[filepath.Base(tmpl)] = template.Must(template.ParseFiles(files...))
	}

	// Set up routes
	r := mux.NewRouter()
	// Handle authentication.
	r.HandleFunc("/callback", CallbackHandler)
	r.HandleFunc("/login", LoginGET)
	// Handle app routes.
	r.HandleFunc("/", IndexGET)
	r.HandleFunc("/create", CreateGET).Methods("GET")
	r.HandleFunc("/create", CreatePOST).Methods("POST")
	r.HandleFunc("/events", EventsGET)
	r.HandleFunc("/events/{id}", EventGET).Methods("GET")

	// Set up middleware stack
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.HandlerFunc(IsAuthenticated),
		negroni.NewStatic(http.Dir("public")),
	)
	n.UseHandler(handlers.LoggingHandler(os.Stdout, r))
	n.Run(":" + os.Getenv("PORT"))
}

// renderTemplate is a wrapper around template.ExecuteTemplate.
func renderTemplate(w http.ResponseWriter, r *http.Request, filename string, data map[string]interface{}) {
	// Ensure the template exists in the map.
	tmpl, ok := templateMap[filename]
	if !ok {
		err := fmt.Errorf("The template %s does not exist.", filename)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		logrus.WithError(err).Error("Failed to ExecuteTemplate")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetUserID returns the ID for a given user.
func GetUserID(r *http.Request) (int, error) {
	session, err := cookieStore.Get(r, "auth-session")
	if err != nil {
		return 0, err
	}
	var userID int
	profile, ok := session.Values["profile"].(map[string]interface{})
	if !ok {
		return 0, errors.New("no profile data")
	}

	query := fmt.Sprintf("SELECT id FROM account WHERE email = '%s'", profile["email"])
	err = ppdb.QueryRow(query).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
