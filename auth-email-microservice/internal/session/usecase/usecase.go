package usecase

import (
	"context"

	"github.com/Chuuch/ecom-microservices/config"
	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/internal/session"
	"github.com/opentracing/opentracing-go"
)

// Session usecase
type sessionUC struct {
	sessionRepo session.SessionRepository
	cfg         *config.Config
}

func NewSessionUseCase(sessionRepo session.SessionRepository, cfg *config.Config) session.SessionUseCase {
	return &sessionUC{
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

// Create session
func (u *sessionUC) CreateSession(ctx context.Context, session *models.Session, expire int) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sessionUC.CreateSession")
	defer span.Finish()

	return u.sessionRepo.CreateSession(ctx, session, expire)
}

func (u *sessionUC) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sessionUC.GetSession")
	defer span.Finish()

	return u.sessionRepo.GetSessionByID(ctx, sessionID)
}

func (u *sessionUC) DeleteSession(ctx context.Context, sessionID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sessionUC.DeleteSession")
	defer span.Finish()

	return u.sessionRepo.DeleteSession(ctx, sessionID)
}
