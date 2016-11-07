package main

import (
	"html/template"
	"net/http"
	"os"
	"fmt"
	"time"
	"github.com/gorilla/mux"
	"github.com/Sirupsen/logrus"
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

// CreatePOST handles POST requests for '/create'
func CreatePOST(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Starting call")
	id, err := GetUserID(r)
	if err != nil {
		//FIXME - this should go elsewhere but trying to get stuff working
		session, err := cookieStore.Get(r, "auth-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		profile, ok := session.Values["profile"].(map[string]interface{})
		if !ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := fmt.Sprintf("INSERT INTO account (email, password, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id")
		ppdb.QueryRow(query, profile["email"], "dummy", profile["given_name"], profile["family_name"]).Scan(&id)
	}
	logrus.Info(id)

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
    logrus.Info(r.Form)

    startTimeDate, err := time.Parse(dateTimeFormat, fmt.Sprintf("%s %s", startDate, startTime))
    if err != nil {
    	logrus.WithError(err).Error("Failed to parse date time")
    }
    endTimeDate, err := time.Parse(dateTimeFormat, fmt.Sprintf("%s %s", endDate, endTime))
    if err != nil {
    	logrus.WithError(err).Error("Failed to parse date time")
    }

	// TODO verify that answerA does not equal answerB
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
	_, err = ppdb.Exec(query, id, title, startTimeDate, endTimeDate,
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

type Event struct {
	Title string
	Timestamp string 
	ID int
}

// EventsGET handles GET requests for '/events'
func EventsGET(w http.ResponseWriter, r *http.Request) {
	query := `SELECT title, start_timestamp, id FROM event;`
	rows, err := ppdb.Query(query)
	if err != nil {
		logrus.Error(err)
	}
	defer rows.Close()
	eventsMap := []Event{}

	var title string
	var id int
	var startTS time.Time
	for rows.Next() {
		if err := rows.Scan(&title, &startTS, &id); err != nil {
			logrus.Error(err)
		}
		eventsMap = append(eventsMap, Event{
			Title: title,
			Timestamp: startTS.Format(niceFormat),
			ID: id,
		})
	}

	data := map[string]interface{}{
		"Page": "Events",
		"Events": eventsMap,
	}
	renderTemplate(w, r, "events.tmpl", data)
}


func EventGET(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var title, desc, location string
	var start, end time.Time
	var eventType, topic int

	query := fmt.Sprintf(`SELECT title, start_timestamp, end_timestamp, description, event_type, event_topic, location FROM event WHERE id = %s;`, id)
	err := ppdb.QueryRow(query).Scan(&title, &start, &end, &desc, &eventType, &topic, &location)
	if err != nil {
		logrus.Error(err)
	}
	startTime := start.Format(niceFormat)
	endTime := end.Format(niceFormat)
	
	data := map[string]interface{}{
		"Page": "Events",
		"Title": title,
		"Start": startTime,
		"End": endTime,
		"Desc": desc,
		"Type": eventType,
		"Topic": topic,
		"Location": location,
	}
	renderTemplate(w, r, "event.tmpl", data)
}