package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

// IndexGET handles GET requests for the home page.
func IndexGET(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Hitting the home page")

	data := map[string]interface{}{
		"Page": "Home",
	}
	renderTemplate(w, r, "index.tmpl", data)
}
