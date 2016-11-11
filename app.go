package main

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/chloearianne/protestpulse/db"
	"github.com/chloearianne/protestpulse/session"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var templateMap = make(map[string]*template.Template)
var cookieStore *sessions.CookieStore

// var ppdb *sql.DB

// App bundles resources used by the application.
type App struct {
	db *db.Database
}

// AppConfig is a container for all app configuration parameters
// that are to be extracted from the YAML config file.
type AppConfig struct {
	CookieKey string    `yaml:"cookie_key"`
	DBConfig  db.Config `yaml:"db_config"`
}

func main() {
	// Load the .env file for environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load AppConfig
	// TODO make config path configurable via env var directly
	c := loadConfig(fmt.Sprintf("conf/%s.yaml", os.Getenv("ENV")))

	// Set up the database
	ppdb := db.New(c.DBConfig)
	defer ppdb.Close()

	cookieStore = sessions.NewCookieStore([]byte(c.CookieKey))
	// Register types to be stored on session
	gob.Register(map[string]interface{}{})
	gob.Register(&session.Profile{})

	// Populate templateMap by processing the 'templates' and 'layouts' directories.
	loadTemplates()

	// Create App object
	app := App{
		db: ppdb,
	}

	// Set up routes
	r := mux.NewRouter()
	// Handle authentication.
	r.HandleFunc("/auth/logout", LogoutHandler)
	r.HandleFunc("/auth/login", LoginHandler)
	r.HandleFunc("/auth/callback", app.CallbackHandler)
	// Handle app routes.
	r.HandleFunc("/", app.IndexGET)
	r.HandleFunc("/create", app.CreateGET).Methods("GET")
	r.HandleFunc("/create", app.CreatePOST).Methods("POST")
	r.HandleFunc("/events", app.EventsGET)
	r.HandleFunc("/events/{id}", app.EventGET).Methods("GET")

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

// loadTemplates generates a map of template file name to complete templates.
// If any failures occur when compiling the templates, a fatal error will be logged.
func loadTemplates() {
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
}

// loadConfig extracts the configuration file into an AppConfig object.
func loadConfig(path string) *AppConfig {
	if _, err := os.Stat(path); err != nil {
		logrus.WithField("path", path).WithError(err).Fatal("Could not find config file")
	}
	logrus.Infof("Using config file at %q", path)
	config, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Fatal(err)
	}

	c := &AppConfig{}
	err = yaml.Unmarshal([]byte(config), c)
	if err != nil {
		logrus.Fatal(err)
	}

	return c
}
