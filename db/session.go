package main

import (
	"fmt"
	"net/http"
)

type Profile struct {
	UserID     string
	Email      string
	GivenName  string
	FamilyName string
	Picture    string
}

func GetProfile(r *http.Request) (*Profile, err) {
	session, err := cookieStore.Get(r, "auth-session")
	if err != nil {
		return fmt.Errorf("Could not get auth-session: %v", err)
	}

	profile := session.Values["profile"]
	if profile == nil {
		return fmt.Errorf("Could not find profile key in session")
	}

	return &Profile{
		UserID:     profile["user_id"],
		Email:      profile["email"],
		GivenName:  profile["given_name"],
		FamilyName: profile["family_name"],
		Picture:    profile["picture"],
	}
}
