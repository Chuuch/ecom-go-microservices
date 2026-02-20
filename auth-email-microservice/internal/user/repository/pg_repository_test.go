package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/internal/user"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func SetupPG() (user.UserPGRepository, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)
	return userPGRepository, mock, err
}

func TestUserPGRepository_Register(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)

	// Column order must match RETURNING * (table definition order: user_id, first_name, last_name, email, password, role, created_at, updated_at, avatar)
	columns := []string{"user_id", "first_name", "last_name", "email", "password", "role", "avatar", "created_at", "updated_at"}
	userUUID := uuid.New()
	mockUser := models.User{
		UserID:    userUUID,
		FirstName: "Barry",
		LastName:  "Allen",
		Email:     "barry@flash.com",
		Password:  "superfast",
		Role:      "admin",
		Avatar:    "avatar",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		userUUID,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Role,
		mockUser.Avatar,
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery(createUserQuery).WithArgs(
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Role,
		mockUser.Avatar,
	).WillReturnRows(rows)

	createdUser, err := userPGRepository.Register(context.Background(), &mockUser)
	require.NoError(t, err)
	require.Equal(t, createdUser.UserID, userUUID)
	require.Equal(t, createdUser.FirstName, mockUser.FirstName)
	require.Equal(t, createdUser.LastName, mockUser.LastName)
	require.Equal(t, createdUser.Email, mockUser.Email)
	require.Equal(t, createdUser.Password, mockUser.Password)
	require.Equal(t, createdUser.Role, mockUser.Role)
	require.Equal(t, createdUser.Avatar, mockUser.Avatar)

	mock.ExpectationsWereMet()
}

func TestUserPGRepository_FindByEmail(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)

	columns := []string{"user_id", "email", "first_name", "last_name", "role", "avatar", "created_at", "updated_at"}
	userUUID := uuid.New()
	mockUser := models.User{
		UserID:    userUUID,
		Email:     "barry@flash.com",
		FirstName: "Barry",
		LastName:  "Allen",
		Role:      "admin",
		Avatar:    "avatar",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		userUUID,
		mockUser.Email,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Role,
		mockUser.Avatar,
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery(findByEmailQuery).WithArgs(mockUser.Email).WillReturnRows(rows)

	foundUser, err := userPGRepository.FindByEmail(context.Background(), mockUser.Email)
	require.NoError(t, err)
	require.Equal(t, foundUser.Email, mockUser.Email)

	mock.ExpectationsWereMet()
}

func TestUserPGRepository_FindById(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)

	columns := []string{"user_id", "email", "first_name", "last_name", "role", "avatar", "created_at", "updated_at"}
	userUUID := uuid.New()

	mockUser := models.User{
		UserID:    userUUID,
		FirstName: "Barry",
		LastName:  "Allen",
		Email:     "barry@flash.com",
		Role:      "admin",
		Avatar:    "avatar",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		userUUID,
		mockUser.Email,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Role,
		mockUser.Avatar,
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery(findByIdQuery).WithArgs(userUUID).WillReturnRows(rows)

	foundUser, err := userPGRepository.FindByID(context.Background(), userUUID)
	require.NoError(t, err)
	require.Equal(t, foundUser.UserID, userUUID)

	mock.ExpectationsWereMet()
}
