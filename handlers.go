package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

// IndexGET handles GET requests for '/'
func IndexGET(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Hitting the home page")

	data := map[string]interface{}{
		"Page": "Home",
	}
	renderTemplate(w, r, "index.tmpl", data)
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
	data := map[string]interface{}{
		"Page": "Events",
	}
	renderTemplate(w, r, "events.tmpl", data)
}
