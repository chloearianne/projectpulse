package main

import (
	_ "crypto/sha512"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/chloearianne/protestpulse/session"

	"golang.org/x/oauth2"
)

// IsAuthenticated is middleware that checks to see whether the user is logged in.
func IsAuthenticated(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	session, err := cookieStore.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := session.Values["profile"]; !ok {
		// Only redirect if not currently requesting an /auth/ route or any
		// static assets to avoid an endless loop or blocking static resources.
		if !strings.Contains(r.URL.Path, "/auth/") && !strings.Contains(r.URL.Path, "/static/") {
			loginPath := "/auth/login"
			logrus.WithField("requestURL", r.URL.Path).Infof("Redirecting to %s", loginPath)
			// FIXME - last thing seen is /callback so we need to pass the path forward to deep link
			// http.Redirect(w, r, fmt.Sprintf("/auth/login?redir=%s", r.URL.Path), http.StatusSeeOther)
			http.Redirect(w, r, loginPath, http.StatusSeeOther)
			return
		}
	}

	next(w, r)
}

// LoginHandler handles requests for '/auth/login'.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Page":              "Login",
		"Auth0ClientId":     os.Getenv("AUTH0_CLIENT_ID"),
		"Auth0ClientSecret": os.Getenv("AUTH0_CLIENT_SECRET"),
		"Auth0Domain":       os.Getenv("AUTH0_DOMAIN"),
		"Auth0CallbackURL":  template.URL(os.Getenv("AUTH0_CALLBACK_URL")),
	}
	renderTemplate(w, r, "login.tmpl", data)
}

// LogoutHandler handles requests for '/auth/logout' by logging
// the user out from Auth0, clearing the session, and redirecting to login.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := cookieStore.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Clear the session and cookie
	delete(session.Values, "id_token")
	delete(session.Values, "access_token")
	delete(session.Values, "profile")
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to Auth0 logout endpoint followed by a redirect to login.
	logoutPath := fmt.Sprintf("http://%s/v2/logout", os.Getenv("AUTH0_DOMAIN"))
	logoutURL, err := url.Parse(logoutPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	q := logoutURL.Query()
	q.Set("returnTo", fmt.Sprintf("http://%s/auth/login", r.Host))
	q.Set("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	q.Set("secret", os.Getenv("AUTH0_CLIENT_SECRET"))
	logoutURL.RawQuery = q.Encode()

	http.Redirect(w, r, logoutURL.String(), http.StatusSeeOther)
}

// CallbackHandler will be called by Auth0 once it redirects to the app.
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	domain := os.Getenv("AUTH0_DOMAIN")

	conf := &oauth2.Config{
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("AUTH0_CALLBACK_URL"),
		Scopes:       []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + domain + "/authorize",
			TokenURL: "https://" + domain + "/oauth/token",
		},
	}

	code := r.URL.Query().Get("code")
	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Now getting the userinfo
	client := conf.Client(oauth2.NoContext, token)
	resp, err := client.Get("https://" + domain + "/userinfo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	raw, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var profile session.Profile
	if err = json.Unmarshal(raw, &profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := cookieStore.Get(r, "auth-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["id_token"] = token.Extra("id_token")
	session.Values["access_token"] = token.AccessToken
	session.Values["profile"] = profile
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to logged in page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
