package session

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

// Profile contains user data provided by the auth service.
type Profile struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
}

// GetProfile introspects the auth-session of the given cookie and request
// and returns a Profile of user data.
func GetProfile(r *http.Request, cookie *sessions.CookieStore) (*Profile, error) {
	session, err := cookie.Get(r, "auth-session")
	if err != nil {
		return nil, fmt.Errorf("Could not get auth-session: %v", err)
	}

	if session == nil || session.Values["profile"] == nil {
		return nil, fmt.Errorf("Could not find profile in session")
	}

	profile, ok := session.Values["profile"].(Profile)
	if !ok {
		return nil, fmt.Errorf("Could not assert profile")
	}

	p := &Profile{
		UserID:     profile.UserID,
		Email:      profile.Email,
		GivenName:  profile.GivenName,
		FamilyName: profile.FamilyName,
		Picture:    profile.Picture,
	}

	return p, nil
}
