package db

// CreateUser uses the given session profile to create a user record
// and returns the newly created user's ID.
func CreateUser(p *Profile) (int, error) {
	query := `INSERT INTO account (
				email, password, first_name, last_name
			) VALUES (
				$1, $2, $3, $4
			) RETURNING id`

	var userID int
	err := ppdb.QueryRow(query, p.Email, p.Password, p.GivenName, p.FamilyName).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// GetUserID returns the user ID for a given email.
func GetUserID(email string) (int, error) {
	var userID int
	query = `SELECT id
			FROM account
			WHERE email = $1`
	if err = ppdb.QueryRow(query, email).Scan(&userID); err != nil {
		return 0, err
	}
	return userID, nil
}
