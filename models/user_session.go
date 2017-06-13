package models

import (
	"encoding/json"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

type UserSession struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	SessionKey  string     `json:"session_key" db:"session_key"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	LoginTime   *time.Time `json:"login_time" db:"login_time"`
	LastSeeTime *time.Time `json:"last_see_time" db:"last_see_time"`
}

// String is not required by pop and may be deleted
func (u UserSession) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// UserSessions is not required by pop and may be deleted
type UserSessions []UserSession

// String is not required by pop and may be deleted
func (u UserSessions) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run everytime you call a "pop.Validate" method.
// This method is not required and may be deleted.
func (u *UserSession) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: u.SessionKey, Name: "SessionKey"},
	), nil
}

// ValidateSave gets run everytime you call "pop.ValidateSave" method.
// This method is not required and may be deleted.
func (u *UserSession) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run everytime you call "pop.ValidateUpdate" method.
// This method is not required and may be deleted.
func (u *UserSession) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
