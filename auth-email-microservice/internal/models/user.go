package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User base model

type User struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id" validate:"omitempty"`
	Email     string    `json:"email" db:"email" validate:"omitempty,lte=60,email"`
	FirstName string    `json:"first_name" db:"first_name" validate:"required"`
	LastName  string    `json:"last_name" db:"last_name" validate:"required"`
	Role      string    `json:"role" db:"role"`
	Avatar    string    `json:"avatar" db:"avatar"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// Sanitize password
func (u *User) SanitizePassword() {
	u.Password = ""
}

// Hash user password
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// Compare user password
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// Prepare user for register
func (u *User) PrepareRegister() error {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Password = strings.TrimSpace(u.Password)

	if err := u.HashPassword(); err != nil {
		return err
	}

	if u.Role != "" {
		u.Role = strings.ToLower(strings.TrimSpace(u.Role))
	}

	return nil
}
