package repository

import (
	"context"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/pkg/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// Email repository
type EmailRepository struct {
	db *sqlx.DB
}

func NewEmailRepository(db *sqlx.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

func (e *EmailRepository) CreateEmail(ctx context.Context, email *models.Email) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailRepository.CreateEmail")
	defer span.Finish()

	var id uuid.UUID
	if err := e.db.QueryRowContext(ctx, createEmailQuery, email.To, email.From, email.Subject, email.Body, email.ContentType).Scan(&id); err != nil {
		return nil, errors.Wrap(err, "CreateEmail.QueryRowContext")
	}

	email.EmailID = id

	return email, nil
}

func (e *EmailRepository) FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailRepository.FindEmailById")
	defer span.Finish()

	var to string
	email := &models.Email{}

	if err := e.db.QueryRowContext(ctx, findEmailByIdQuery, emailID).Scan(&email.EmailID, &to, &email.From, &email.Subject, &email.Body, &email.ContentType, &email.CreatedAt); err != nil {
		return nil, errors.Wrap(err, "FindEmailById.QueryRowContext")
	}

	email.SetToFromString(to)
	return email, nil
}

// Find emails by receiver address
func (e *EmailRepository) FindEmailsByReceiver(ctx context.Context, receiverEmail string, paginationQuery *utils.PaginationQuery) (*models.EmailsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailRepository.FindEmailsByReceiver")
	defer span.Finish()

	var totalCount uint64
	if err := e.db.QueryRowContext(ctx, totalCountQuery, receiverEmail).Scan(&totalCount); err != nil {
		return nil, errors.Wrap(err, "FindEmailsByReceiver.QueryRowxContext")
	}

	if totalCount == 0 {
		return &models.EmailsList{Emails: []*models.Email{}}, nil
	}

	rows, err := e.db.QueryxContext(ctx, findEmailsByReceiverQuery, receiverEmail, paginationQuery.GetLimit(), paginationQuery.GetOffset())
	if err != nil {
		return nil, errors.Wrap(err, "FindEmailsByReceiver.QueryxContext")
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "FindEmailsByReceiver.QueryxContext.Err")
	}

	emails := make([]*models.Email, 0, paginationQuery.GetSize())
	for rows.Next() {
		var mailTo string
		email := &models.Email{}
		if err := rows.Scan(
			&email.EmailID,
			&mailTo,
			&email.From,
			&email.Subject,
			&email.Body,
			&email.ContentType,
			&email.CreatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "FindEmailsByReceiver.QueryxContext.Scan")
		}

		email.SetToFromString(mailTo)
		emails = append(emails, email)
	}

	return &models.EmailsList{
		TotalCount: totalCount,
		TotalPages: utils.GetTotalPages(totalCount, paginationQuery.GetSize()),
		HasMore:    utils.GetHasMore(paginationQuery.GetPage(), totalCount, paginationQuery.GetSize()),
		Page:       paginationQuery.GetPage(),
		Size:       paginationQuery.GetSize(),
		Emails:     emails,
	}, nil
}
