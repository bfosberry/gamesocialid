package models

import (
	"encoding/json"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

type Credential struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Provider     string    `json:"provider" db:"provider"`
	Name         string    `json:"name" db:"name"`
	Nickname     string    `json:"nickname" db:"nickname"`
	Email        string    `json:"email" db:"email"`
	ProfileUrl   string    `json:"profile_url" db:"profile_url"`
	ImageUrl     string    `json:"image_url" db:"image_url"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	TokenExpiry  string    `json:"token_expiry" db:"token_expiry"`
	Uid          string    `json:"uid" db:"uid"`
}

// String is not required by pop and may be deleted
func (c Credential) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Credentials is not required by pop and may be deleted
type Credentials []Credential

// String is not required by pop and may be deleted
func (c Credentials) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Validate gets run everytime you call a "pop.Validate" method.
// This method is not required and may be deleted.
func (c *Credential) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: c.Provider, Name: "Provider"},
		&validators.StringIsPresent{Field: c.Name, Name: "Name"},
		&validators.StringIsPresent{Field: c.Nickname, Name: "Nickname"},
		&validators.StringIsPresent{Field: c.Email, Name: "Email"},
		&validators.StringIsPresent{Field: c.ProfileUrl, Name: "ProfileUrl"},
		&validators.StringIsPresent{Field: c.ImageUrl, Name: "ImageUrl"},
		&validators.StringIsPresent{Field: c.RefreshToken, Name: "RefreshToken"},
		&validators.StringIsPresent{Field: c.TokenExpiry, Name: "TokenExpiry"},
		&validators.StringIsPresent{Field: c.Uid, Name: "Uid"},
	), nil
}

// ValidateSave gets run everytime you call "pop.ValidateSave" method.
// This method is not required and may be deleted.
func (c *Credential) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run everytime you call "pop.ValidateUpdate" method.
// This method is not required and may be deleted.
func (c *Credential) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
