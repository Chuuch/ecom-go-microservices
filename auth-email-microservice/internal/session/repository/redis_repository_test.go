package repository

import (
	"context"
	"log"
	"testing"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/internal/session"
	"github.com/alicebob/miniredis"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func SetupRedis() session.SessionRepository {
	mr, err := miniredis.Run()

	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	sessionRepository := NewSessionRepository(client, nil)
	return sessionRepository
}

func TestCreateSession(t *testing.T) {
	t.Parallel()

	sessionRepository := SetupRedis()

	t.Run("CreateSession", func(t *testing.T) {
		sessionUUID := uuid.New()
		session := &models.Session{
			SessionID: sessionUUID.String(),
			UserID:    sessionUUID.String(),
		}

		s, err := sessionRepository.CreateSession(context.Background(), session, 10)
		require.NoError(t, err)
		require.NotEqual(t, s, "")
	})
}

func TestGetSessionByID(t *testing.T) {
	t.Parallel()
	sessionRepository := SetupRedis()

	t.Run("GetSessionByID", func(t *testing.T) {
		sessionUUID := uuid.New()
		session := &models.Session{
			SessionID: sessionUUID.String(),
			UserID:    sessionUUID.String(),
		}

		createdSession, err := sessionRepository.CreateSession(context.Background(), session, 10)
		require.NoError(t, err)
		require.NotEqual(t, createdSession, "")

		s, err := sessionRepository.GetSessionByID(context.Background(), createdSession)
		require.NoError(t, err)
		require.Equal(t, s.SessionID, createdSession)
		require.Equal(t, s.UserID, sessionUUID.String())
	})
}

func TestDeleteSession(t *testing.T) {
	t.Parallel()

	sessionRepository := SetupRedis()

	t.Run("DeleteSession", func(t *testing.T) {
		sessionID := uuid.New()
		err := sessionRepository.DeleteSession(context.Background(), sessionID.String())
		require.NoError(t, err)
	})
}
