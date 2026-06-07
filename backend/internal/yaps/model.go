package yaps

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("yap not found")
	ErrForbidden = errors.New("not allowed")
	ErrEmpty     = errors.New("yap is empty")
	ErrTooLong   = errors.New("yap is too long")
)

const MaxContentLen = 280

// KeyToPublicURL resolves a storage key to a browser-facing URL.
type KeyToPublicURL func(key string) string

// Author is the compact user view embedded in a yap.
type Author struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	AvatarURL   string    `json:"avatarUrl"`
}

type Media struct {
	ID       uuid.UUID `json:"id"`
	URL      string    `json:"url"`
	Position int       `json:"position"`
}

type Yap struct {
	ID        uuid.UUID  `json:"id"`
	Author    Author     `json:"author"`
	Content   string     `json:"content"`
	ReplyToID *uuid.UUID `json:"replyToId,omitempty"`
	QuoteOfID *uuid.UUID `json:"quoteOfId,omitempty"`
	QuoteOf   *Yap       `json:"quoteOf,omitempty"`
	Media     []Media    `json:"media"`

	LikesCount     int `json:"likesCount"`
	RepostsCount   int `json:"repostsCount"`
	RepliesCount   int `json:"repliesCount"`
	QuotesCount    int `json:"quotesCount"`
	BookmarksCount int `json:"bookmarksCount"`

	Liked      bool `json:"liked"`
	Reposted   bool `json:"reposted"`
	Bookmarked bool `json:"bookmarked"`

	// Repost surfacing context: set when a yap appears in a feed because
	// someone the viewer follows reposted it.
	RepostedBy *Author    `json:"repostedBy,omitempty"`
	RepostedAt *time.Time `json:"repostedAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
}

type Page struct {
	Items      []Yap   `json:"items"`
	NextCursor *string `json:"nextCursor,omitempty"`
}

type CreateInput struct {
	AuthorID  uuid.UUID
	Content   string
	ReplyToID *uuid.UUID
	QuoteOfID *uuid.UUID
	MediaKeys []string
}
