package models

import "errors"

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
	ErrNotFound           = errors.New("404 page not found")
	ErrNoAvatarImg        = errors.New("Avatar image missing in post data.")
)
