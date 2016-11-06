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
	"github.com/codegangsta/negroni"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var templateMap = make(map[string]*template.Template)
var cookieStore *sessions.CookieStore

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

	// // Set up the environment-specific database handle
	// dbh := db.New(db.Params{
	// 	User:     c.DBUser,
	// 	Password: c.DBPassword,
	// 	DBName:   c.DBName,
	// 	Host:     c.DBHost,
	// })
	// defer dbh.Conn.Close()

	// Set up routes
	r := mux.NewRouter()
	// Handle authentication.
	r.HandleFunc("/callback", CallbackHandler)
	r.Handle("/", negroni.New(
		negroni.HandlerFunc(IsAuthenticated),
		negroni.Wrap(http.HandlerFunc(IndexGET)),
	))
	// Handle app routes.
	r.HandleFunc("/login", LoginGET)
	r.HandleFunc("/create", CreateGET)
	r.HandleFunc("/events", EventsGET)

	// Set up middleware stack
	n := negroni.New(
		negroni.NewRecovery(),
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
		logrus.WithError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
