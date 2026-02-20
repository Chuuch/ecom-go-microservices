package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/internal/user"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
)

const (
	basePrefix = "users:"
)

// Auth redis repository
type userRedisRepo struct {
	redisClient *redis.Client
	basePrefix  string
	logger      logger.Logger
}

func NewUserRedisRepository(redisClient *redis.Client, logger logger.Logger) user.UserRedisRepository {
	return &userRedisRepo{
		redisClient: redisClient,
		basePrefix:  basePrefix,
		logger:      logger,
	}
}

// Get user by id
func (r *userRedisRepo) GetByIDCtx(ctx context.Context, key string) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userRedisRepo.GetByIDCtx")
	defer span.Finish()

	userBytes, err := r.redisClient.Get(ctx, r.createKey(key)).Bytes()
	if err != nil {
		if err != redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	user := &models.User{}
	if err := json.Unmarshal(userBytes, &user); err != nil {
		r.logger.Error("userRedisRepo.GetByIDCtx: %v", err)
		return nil, err
	}

	return user, nil
}

// Cache user with duration in seconds
func (r *userRedisRepo) SetUserCtx(ctx context.Context, key string, seconds int, user *models.User) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userRedisRepo.SetUserCtx")
	defer span.Finish()

	userBytes, err := json.Marshal(&user)
	if err != nil {
		r.logger.Error("userRedisRepo.SetUserCtx.Json.Marshal: %v", err)
		return err
	}

	if err := r.redisClient.Set(ctx, r.createKey(key), userBytes, time.Duration(seconds)*time.Second).Err(); err != nil {
		r.logger.Error("userRedisRepo.SetUserCtx: %v", err)
		return err
	}

	return nil
}

// Delete user by key
func (r *userRedisRepo) DeleteUserCtx(ctx context.Context, key string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userRedisRepo.DeleteUserCtx")
	defer span.Finish()

	return r.redisClient.Del(ctx, r.createKey(key)).Err()
}

func (r *userRedisRepo) createKey(key string) string {
	return fmt.Sprintf("%s:%s", r.basePrefix, key)
}
