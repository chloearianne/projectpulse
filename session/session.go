package session

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

// Profile contains user data provided by the auth service.
type Profile struct {
	UserID     string
	Email      string
	GivenName  string
	FamilyName string
	Picture    string
}

func GetProfile(r *http.Request, cookie *sessions.CookieStore) (*Profile, error) {
	session, err := cookie.Get(r, "auth-session")
	if err != nil {
		return nil, fmt.Errorf("Could not get auth-session: %v", err)
	}

	profile, ok := session.Values["profile"].(map[string]string)
	if profile == nil {
		return nil, fmt.Errorf("Could not find profile key in session")
	}
	if !ok {
		return nil, fmt.Errorf("Failed to assert profile as string map")
	}

	p := &Profile{
		UserID:     profile["user_id"],
		Email:      profile["email"],
		GivenName:  profile["given_name"],
		FamilyName: profile["family_name"],
		Picture:    profile["picture"],
	}

	return p, nil
}
