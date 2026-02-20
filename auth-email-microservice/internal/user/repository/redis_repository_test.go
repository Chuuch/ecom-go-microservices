package repository

import (
	"context"
	"log"
	"testing"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/alicebob/miniredis"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func SetupRedis() *userRedisRepo {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	userRedisRepository := NewUserRedisRepository(client, nil)
	return userRedisRepository.(*userRedisRepo)
}

func TestUserRedisRepository_SetUserCtx(t *testing.T) {
	t.Parallel()

	redisRepository := SetupRedis()

	t.Run("SetUserCtx", func(t *testing.T) {
		user := &models.User{
			UserID: uuid.New(),
		}

		err := redisRepository.SetUserCtx(context.Background(), redisRepository.createKey(user.UserID.String()), 10, user)
		require.NoError(t, err)
	})
}

func TestUserRedisRepository_GetByIDCtx(t *testing.T) {
	t.Parallel()

	redisRepo := SetupRedis()

	t.Run("GetByIDCtx", func(t *testing.T) {
		user := &models.User{
			UserID: uuid.New(),
		}

		err := redisRepo.SetUserCtx(context.Background(), redisRepo.createKey(user.UserID.String()), 10, user)
		require.NoError(t, err)

		foundUser, err := redisRepo.GetByIDCtx(context.Background(), redisRepo.createKey(user.UserID.String()))
		require.NoError(t, err)
		require.Equal(t, foundUser.UserID, user.UserID)
	})
}

func TestUserRedisRepository_DeleteUserCtx(t *testing.T) {
	t.Parallel()

	redisRepo := SetupRedis()

	t.Run("DeleteUserCtx", func(t *testing.T) {
		user := &models.User{
			UserID: uuid.New(),
		}

		err := redisRepo.DeleteUserCtx(context.Background(), redisRepo.createKey(user.UserID.String()))
		require.NoError(t, err)
	})
}
