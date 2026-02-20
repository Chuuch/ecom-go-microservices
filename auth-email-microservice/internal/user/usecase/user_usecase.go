package usecase

import (
	"context"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/internal/user"
	grpcerrors "github.com/Chuuch/ecom-microservices/pkg/grpc_errors"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	userByIdCacheDuration = 3600
)

// Auth UseCase
type UserUseCase struct {
	logger        logger.Logger
	userPGRepo    user.UserPGRepository
	userRedisRepo user.UserRedisRepository
}

// New Auth UseCase
func NewUserUseCase(logger logger.Logger, userPGRepo user.UserPGRepository, userRedisRepo user.UserRedisRepository) *UserUseCase {
	return &UserUseCase{
		logger:        logger,
		userPGRepo:    userPGRepo,
		userRedisRepo: userRedisRepo,
	}
}

func (u *UserUseCase) Register(ctx context.Context, user *models.User) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserUseCase.Register")
	defer span.Finish()

	existingUser, err := u.userPGRepo.FindByEmail(ctx, user.Email)
	if existingUser != nil || err == nil {
		return nil, grpcerrors.ErrEmailAlreadyExists
	}

	return u.userPGRepo.Register(ctx, user)
}

func (u *UserUseCase) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserUseCase.FindByEmail")
	defer span.Finish()

	findByEmail, err := u.userPGRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "userPGRepo.FindByEmail")
	}

	findByEmail.SanitizePassword()

	return findByEmail, nil
}

func (u *UserUseCase) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserUseCase.FindByID")
	defer span.Finish()

	cachedUser, err := u.userRedisRepo.GetByIDCtx(ctx, userID.String())
	if err != nil && !errors.Is(err, redis.Nil) {
		u.logger.Error("userUC.FindByID.userRedisRepo.GetByIDCtx: %v", err)
	}

	if cachedUser != nil {
		return cachedUser, nil
	}

	foundUser, err := u.userPGRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "userUC.FindById.userPGRepo.FindByID")
	}

	if err := u.userRedisRepo.SetUserCtx(ctx, userID.String(), userByIdCacheDuration, foundUser); err != nil {
		u.logger.Error("userUC.FindByID.SetUserCtx: %v", err)
	}

	return foundUser, nil
}

func (u *UserUseCase) Login(ctx context.Context, email, password string) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserUseCase.Login")
	defer span.Finish()

	foundUser, err := u.userPGRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "userUC.Login.userPGRepo.FindByEmail")
	}

	if err := foundUser.ComparePassword(password); err != nil {
		return nil, errors.Wrap(err, "user.ComparePassword")
	}

	return foundUser, nil
}
