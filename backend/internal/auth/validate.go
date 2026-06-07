package auth

import (
	"regexp"
	"strings"
)

var (
	emailRe    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	usernameRe = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)
)

// NormalizeEmail lower-cases and trims; CITEXT also makes lookups case-insensitive.
func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeUsername(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeIdentifier(s string) string {
	return NormalizeEmail(s)
}

func ValidateRegister(in RegisterInput) error {
	if !usernameRe.MatchString(in.Username) {
		return &ValidationError{Field: "username", Message: "username must be 3-20 chars: lowercase letters, digits, underscore"}
	}
	if !emailRe.MatchString(in.Email) {
		return &ValidationError{Field: "email", Message: "enter a valid email address"}
	}
	if dn := strings.TrimSpace(in.DisplayName); len(dn) > 50 {
		return &ValidationError{Field: "displayName", Message: "display name must be at most 50 characters"}
	}
	return validatePassword(in.Password)
}

func validatePassword(pw string) error {
	if len(pw) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	if len(pw) > 200 {
		return &ValidationError{Field: "password", Message: "password is too long"}
	}
	return nil
}
