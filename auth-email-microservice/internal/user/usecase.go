package user

import (
	"context"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/google/uuid"
)

// User usecase interface
type UserUseCase interface {
	Register(ctx context.Context, user *models.User) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.User, error)
}
