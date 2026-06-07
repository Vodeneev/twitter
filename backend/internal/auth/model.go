package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrEmailTaken           = errors.New("email already registered")
	ErrUsernameTaken        = errors.New("username already taken")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrAlreadyVerified      = errors.New("email already verified")
	ErrUserBanned           = errors.New("user banned")
	ErrSessionExpired       = errors.New("session expired")
	ErrVerificationInvalid  = errors.New("verification token invalid or expired")
	ErrPasswordResetInvalid = errors.New("password reset token invalid or expired")
)

const (
	DefaultSessionTTL    = 24 * time.Hour
	RememberMeSessionTTL = 30 * 24 * time.Hour
)

// User is the public-facing user representation returned by the API.
type User struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	DisplayName    string    `json:"displayName"`
	Bio            string    `json:"bio"`
	Location       string    `json:"location"`
	Website        string    `json:"website"`
	AvatarURL      string    `json:"avatarUrl"`
	HeaderURL      string    `json:"headerUrl"`
	IsAdmin        bool      `json:"isAdmin"`
	IsBanned       bool      `json:"isBanned"`
	FollowersCount int       `json:"followersCount"`
	FollowingCount int       `json:"followingCount"`
	YapsCount      int       `json:"yapsCount"`
	EmailVerified  bool      `json:"emailVerified"`
	CreatedAt      time.Time `json:"createdAt"`

	// Following is set relative to the current viewer in directory/profile responses.
	Following bool `json:"following"`

	// Raw storage keys; resolved to URLs at the HTTP boundary.
	AvatarKey string `json:"-"`
	HeaderKey string `json:"-"`
}

type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	ExpiresAt  time.Time
	RememberMe bool
}

type RegisterInput struct {
	Username    string
	Email       string
	DisplayName string
	Password    string
}

type LoginInput struct {
	Identifier string
	Password   string
	RememberMe bool
	UserAgent  string
	IP         string
}

// ValidationError carries a field-specific, user-presentable validation problem.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }
