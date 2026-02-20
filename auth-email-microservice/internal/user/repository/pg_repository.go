package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
)

// Auth Repository
type UserPGRepository struct {
	db *sqlx.DB
}

func NewUserPGRepository(db *sqlx.DB) *UserPGRepository {
	return &UserPGRepository{
		db: db,
	}
}

func (u *UserPGRepository) Register(ctx context.Context, user *models.User) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserRepository.Register")
	defer span.Finish()

	createdUser := &models.User{}

	if err := u.db.QueryRowxContext(ctx, createUserQuery, user.FirstName, user.LastName, user.Email, user.Password, user.Role, user.Avatar).StructScan(createdUser); err != nil {
		return nil, errors.Wrap(err, "Register.QueryRowxContext")
	}

	return createdUser, nil
}

func (u *UserPGRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserRepository.FindByEmail")
	defer span.Finish()

	user := &models.User{}

	if err := u.db.GetContext(ctx, user, findByEmailQuery, email); err != nil {
		return nil, errors.Wrap(err, "FindByEmail.GetContext")
	}

	return user, nil
}

func (u *UserPGRepository) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserRepository.FindByID")
	defer span.Finish()

	user := &models.User{}

	if err := u.db.GetContext(ctx, user, findByIdQuery, userID); err != nil {
		return nil, errors.Wrap(err, "FindByID.GetContext")
	}

	return user, nil
}
