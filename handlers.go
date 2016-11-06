package main

import (
	"html/template"
	"net/http"
	"os"
	"time"
)

// IndexGET handles GET requests for '/'
func IndexGET(w http.ResponseWriter, r *http.Request) {
	session, err := cookieStore.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Page":    "Home",
		"Profile": session.Values["profile"],
	}
	renderTemplate(w, r, "index.tmpl", data)
}

// LoginGET handles GET requests for '/login'
func LoginGET(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Page":              "Login",
		"Auth0ClientId":     os.Getenv("AUTH0_CLIENT_ID"),
		"Auth0ClientSecret": os.Getenv("AUTH0_CLIENT_SECRET"),
		"Auth0Domain":       os.Getenv("AUTH0_DOMAIN"),
		"Auth0CallbackURL":  template.URL(os.Getenv("AUTH0_CALLBACK_URL")),
	}
	renderTemplate(w, r, "login.tmpl", data)
}

// CreateGET handles GET requests for '/create'
func CreateGET(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Page": "Create Event",
	}
	renderTemplate(w, r, "create.tmpl", data)
}

// EventsGET handles GET requests for '/events'
func EventsGET(w http.ResponseWriter, r *http.Request) {
	query := `SELECT
							title, start_timestamp
						FROM
							event;`
	rows, err := ppdb.Query(query)
	if err != nil {
		logrus.Error(err)
	}
	defer rows.Close()
	eventsMap := map[string]time.Time
	var title string
	var startTS time.Time
	for rows.Next() {
		if err := rows.Scan(&title, &startTS); err != nil {
			logrus.Error(err)
		}
		eventsMap[title] = startTS
	}

	data := map[string]interface{}{
		"Page": "Events",
		"Events": eventsMap,
	}
	renderTemplate(w, r, "events.tmpl", data)
}
