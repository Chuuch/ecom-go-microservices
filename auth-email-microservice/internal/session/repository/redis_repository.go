package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Chuuch/ecom-microservices/config"
	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/internal/session"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	basePrefix = "sessions:"
)

// Session repository
type sessionRepo struct {
	redisClient *redis.Client
	basePrefix  string
	cfg         *config.Config
}

func NewSessionRepository(redisClient *redis.Client, cfg *config.Config) session.SessionRepository {
	return &sessionRepo{
		redisClient: redisClient,
		basePrefix:  basePrefix,
		cfg:         cfg,
	}
}

func (s *sessionRepo) CreateSession(ctx context.Context, session *models.Session, expire int) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sessionRepo.CreateSession")
	defer span.Finish()

	session.SessionID = uuid.New().String()
	sessionKey := s.createKey(session.SessionID)

	sessionBytes, err := json.Marshal(&session)
	if err != nil {
		return "", errors.WithMessage(err, "sessionRepo.CreateSession.json.Marshal")
	}

	if err := s.redisClient.Set(ctx, sessionKey, sessionBytes, time.Second*time.Duration(expire)).Err(); err != nil {
		return "", errors.WithMessage(err, "sessionRepo.CreateSession.redisClient.Set")
	}

	return session.SessionID, nil
}

func (s *sessionRepo) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sessionRepo.GetSessionByID")
	defer span.Finish()

	sessionBytes, err := s.redisClient.Get(ctx, s.createKey(sessionID)).Bytes()
	if err != nil {
		return nil, errors.WithMessage(err, "sessionRepo.GetSessionByID.redisClient.Get")
	}

	session := &models.Session{}
	if err := json.Unmarshal(sessionBytes, &session); err != nil {
		return nil, errors.WithMessage(err, "sessionRepo.GetSessionByID.json.Unmarshal")
	}

	return session, nil
}

func (s *sessionRepo) DeleteSession(ctx context.Context, sessionID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "sessionRepo.DeleteSession")
	defer span.Finish()

	if err := s.redisClient.Del(ctx, sessionID).Err(); err != nil {
		return errors.WithMessage(err, "sessionRepo.DeleteSession.redisClient.Del")
	}

	return nil
}

func (s *sessionRepo) createKey(sessionID string) string {
	return fmt.Sprintf("%s: %s", s.basePrefix, sessionID)
}
