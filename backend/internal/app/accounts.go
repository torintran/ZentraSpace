package app

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var accountIdentifier = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,39}$`)

type StaffAccount struct {
	Username, Organization, Role, Password string
}

type staffLogin struct {
	organization, role string
	passwordHash       []byte
}

func loadStaff() (map[string]staffLogin, error) {
	accounts := []StaffAccount{
		{"owner", pilotOrg, "owner", os.Getenv("ZS_OWNER_PASSWORD")},
		{"meter", pilotOrg, "meter", os.Getenv("ZS_METER_PASSWORD")},
	}
	if raw := os.Getenv("ZS_STAFF_ACCOUNTS"); raw != "" {
		var extra []StaffAccount
		if err := json.Unmarshal([]byte(raw), &extra); err != nil {
			return nil, errors.New("Invalid ZS_STAFF_ACCOUNTS configuration.")
		}
		accounts = append(accounts, extra...)
	}
	staff := make(map[string]staffLogin, len(accounts))
	for _, account := range accounts {
		if !accountIdentifier.MatchString(account.Username) || !accountIdentifier.MatchString(account.Organization) ||
			(account.Role != "owner" && account.Role != "meter") || len(account.Password) < 16 || len(account.Password) > 72 {
			return nil, errors.New("Invalid staff account name, organization, role, or secret.")
		}
		if _, exists := staff[account.Username]; exists {
			return nil, errors.New("Duplicate staff username.")
		}
		pw, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		staff[account.Username] = staffLogin{account.Organization, account.Role, pw}
	}
	return staff, nil
}
