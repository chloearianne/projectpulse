package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/chloearianne/protestpulse/db"
	"github.com/gorilla/mux"
)

var humanDateFormat = "Jan 02, 2006"

// IndexGET handles GET requests for '/'.
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

// LoginGET handles GET requests for '/login'.
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

// CreatePOST handles POST requests for '/create'.
func CreatePOST(w http.ResponseWriter, r *http.Request) {
	session, err := cookieStore.Get(r, "auth-session")
	if err != nil {
		return 0, err
	}
	profile, ok := session.Values["profile"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("no profile data")
	}
	userID, err := db.GetUserID(profile["email"])
	if err != nil {
		logrus.WithErr(err).Error("User ID not found")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	title := r.Form["title"][0]
	eventType := r.Form["event_type"][0]
	eventTopic := r.Form["event_topic"][0]
	description := r.Form["description"][0]
	location := r.Form["location"][0]
	startDate := r.Form["start_date"][0]
	endDate := r.Form["end_date"][0]
	startTime := r.Form["start_time"][0]
	endTime := r.Form["end_time"][0]

	var dateTimeFormat = "2006-01-02 15:04"
	startTimeDate, err := time.Parse(dateTimeFormat, fmt.Sprintf("%s %s", startDate, startTime))
	if err != nil {
		logrus.WithError(err).Error("Failed to parse date time")
	}
	endTimeDate, err := time.Parse(dateTimeFormat, fmt.Sprintf("%s %s", endDate, endTime))
	if err != nil {
		logrus.WithError(err).Error("Failed to parse date time")
	}

	query := `INSERT INTO event (
		        creator_id, title, start_timestamp, end_timestamp,
		        description, event_topic, event_type, location,
		        stars
			  )
			  VALUES (
			  	$1, $2, $3, $4,
			  	$5, $6, $7, $8,
			  	$9
			  )`
	_, err = ppdb.Exec(query, userID, title, startTimeDate, endTimeDate,
		description, eventTopic, eventType, location,
		0)
	if err != nil {
		logrus.WithError(err).Error("Failed to save event")
	} else {
		logrus.Info("Successfully created new event")
	}
	EventsGET(w, r)
	return
}

// Event contains the metadata related to an activism event.
type Event struct {
	ID        int
	Title     string
	Timestamp string
}

// EventsGET handles GET requests for '/events'.
func EventsGET(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, title, start_timestamp FROM event`
	rows, err := ppdb.Query(query)
	if err != nil {
		logrus.Error(err)
	}
	defer rows.Close()

	eventsMap := []Event{}
	var id int
	var title string
	var startTS time.Time
	for rows.Next() {
		if err := rows.Scan(&id, &title, &startTS); err != nil {
			logrus.Error(err)
		}
		eventsMap = append(eventsMap, Event{
			ID:        id,
			Title:     title,
			Timestamp: startTS.Format(humanDateFormat),
		})
	}

	data := map[string]interface{}{
		"Page":   "Events",
		"Events": eventsMap,
	}
	renderTemplate(w, r, "events.tmpl", data)
}

// EventGET handles GET requests for a single event at '/events/{id}'.
func EventGET(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var title, desc, location string
	var startTime, endTime time.Time
	var eventType, topic int

	query := `SELECT
				title, start_timestamp, end_timestamp,
				description, event_type, event_topic,
				location
			FROM event
			WHERE id = $1`
	err := ppdb.QueryRow(query, id).Scan(
		&title, &start, &end,
		&desc, &eventType, &topic,
		&location,
	)
	if err != nil {
		logrus.Error(err)
	}

	data := map[string]interface{}{
		"Page":     "Events",
		"Title":    title,
		"Start":    startTime.Format(humanDateFormat),
		"End":      endTime.Format(humanDateFormat),
		"Desc":     desc,
		"Type":     eventType,
		"Topic":    topic,
		"Location": location,
	}
	renderTemplate(w, r, "event.tmpl", data)
}
