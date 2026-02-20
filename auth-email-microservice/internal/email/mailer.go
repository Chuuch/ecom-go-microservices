package email

import (
	"context"

	"github.com/Chuuch/ecom-microservices/internal/models"
)

// Mailer interface
type Mailer interface {
	Send(ctx context.Context, email *models.Email) error
}
