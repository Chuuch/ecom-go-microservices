package email

import (
	"context"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/pkg/utils"
	"github.com/google/uuid"
)

// Email Repository interface
type EmailRepository interface {
	CreateEmail(ctx context.Context, email *models.Email) (*models.Email, error)
	FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error)
	FindEmailsByReceiver(ctx context.Context, receiverEmail string, paginationQuery *utils.PaginationQuery) (*models.EmailsList, error)
}
