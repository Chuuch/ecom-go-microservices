package email

import (
	"context"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/pkg/utils"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// Email UseCase interface
type EmailUseCase interface {
	SendEmail(ctx context.Context, delivery amqp.Delivery) error
	PublishEmail(ctx context.Context, email *models.Email) error
	FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error)
	FindEmailsByReceiver(ctx context.Context, receiverEmail string, paginationQuery *utils.PaginationQuery) (*models.EmailsList, error)
}
