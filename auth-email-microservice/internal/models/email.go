package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Chuuch/ecom-microservices/pkg/utils"
	"github.com/google/uuid"
)

// Email model
type Email struct {
	EmailID     uuid.UUID `json:"email_id" db:"email_id" validate:"omitempty"`
	To          []string  `json:"to" db:"to" validate:"required"`
	From        string    `json:"from" db:"from" validate:"required,email"`
	Body        string    `json:"body" db:"body" validate:"required"`
	Subject     string    `json:"subject" db:"subject" validate:"required"`
	ContentType string    `json:"content_type" db:"content_type" validate:"required"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" validate:"omitempty"`
}

// Get string from addresses
func (e *Email) GetToString() string {
	return strings.Join(e.To, ", ")
}

// Set array from string value
func (e *Email) SetToFromString(to string) {
	e.To = strings.Split(to, ", ")
}

// Prepare email for creation
func (e *Email) PrepareAndValidate(ctx context.Context) error {
	e.From = strings.TrimSpace(strings.ToLower(e.From))
	for _, mail := range e.To {
		if !utils.ValidateEmail(mail) {
			return fmt.Errorf("invalid email address: %s", mail)
		}
		mail = strings.TrimSpace(strings.ToLower(mail))
	}
	e.ContentType = "text/html"

	return utils.ValidateStruct(ctx, e)
}

// Emails list with pagination
type EmailsList struct {
	TotalCount uint64   `json:"total_count"`
	TotalPages uint64   `json:"total_pages"`
	HasMore    bool     `json:"has_more"`
	Page       uint64   `json:"page"`
	Size       uint64   `json:"size"`
	Emails     []*Email `json:"emails"`
}
