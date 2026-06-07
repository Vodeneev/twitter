package http

import "github.com/Vodeneev/twitter/backend/internal/auth"

// presentUser resolves avatar/header storage keys to public URLs for the response.
func (a *api) presentUser(u *auth.User) *auth.User {
	if u == nil {
		return nil
	}
	cp := *u
	cp.AvatarURL = a.PhotoURL(cp.AvatarKey)
	cp.HeaderURL = a.PhotoURL(cp.HeaderKey)
	return &cp
}

func (a *api) presentUsers(in []*auth.User) []*auth.User {
	out := make([]*auth.User, 0, len(in))
	for _, u := range in {
		out = append(out, a.presentUser(u))
	}
	return out
}

// presentPublicUser is like presentUser but hides the private email field.
func (a *api) presentPublicUser(u *auth.User) *auth.User {
	cp := a.presentUser(u)
	if cp != nil {
		cp.Email = ""
	}
	return cp
}

func (a *api) presentPublicUsers(in []*auth.User) []*auth.User {
	out := make([]*auth.User, 0, len(in))
	for _, u := range in {
		out = append(out, a.presentPublicUser(u))
	}
	return out
}
