package session

import (
	"context"

	"github.com/Chuuch/ecom-microservices/internal/models"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, session *models.Session, expire int) (string, error)
	GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
}
