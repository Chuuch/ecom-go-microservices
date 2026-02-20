package usecase

import (
	"context"
	"encoding/json"

	"github.com/Chuuch/ecom-microservices/config"
	"github.com/Chuuch/ecom-microservices/internal/email"
	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/Chuuch/ecom-microservices/pkg/utils"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Email UseCase
type EmailUseCase struct {
	emailRepo       email.EmailRepository
	logger          logger.Logger
	mailer          email.Mailer
	cfg             *config.Config
	emailsPublisher email.EmailsPublisher
}

// NewEmailUseCase returns a new EmailUseCase
func NewEmailUseCase(emailRepo email.EmailRepository, logger logger.Logger, mailer email.Mailer, cfg *config.Config, emailsPublisher email.EmailsPublisher) *EmailUseCase {
	return &EmailUseCase{
		emailRepo:       emailRepo,
		logger:          logger,
		mailer:          mailer,
		cfg:             cfg,
		emailsPublisher: emailsPublisher,
	}
}

// Send email
func (e *EmailUseCase) SendEmail(ctx context.Context, delivery amqp.Delivery) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.SendEmail")
	defer span.Finish()

	mail := &models.Email{}
	if err := json.Unmarshal(delivery.Body, mail); err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}

	mail.From = "daniel@skyeystudio.com"

	if err := utils.ValidateStruct(ctx, mail); err != nil {
		return errors.Wrap(err, "utils.ValidateStruct")
	}

	// e.logger.Infof("Sending email: %+v", mail)
	// if err := e.mailer.Send(ctx, mail); err != nil {
	// 	return errors.Wrap(err, "mailer.Send")
	// }

	createdEmail, err := e.emailRepo.CreateEmail(ctx, mail)
	if err != nil {
		return errors.Wrap(err, "emailRepo.CreateEmail")
	}

	span.LogFields(
		log.String("email_id", createdEmail.EmailID.String()),
	)

	e.logger.Infof("Email sent successfully: %+v", createdEmail.EmailID)

	return nil
}

// Publish email
func (e *EmailUseCase) PublishEmail(ctx context.Context, email *models.Email) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.PublishEmail")
	defer span.Finish()

	mailBytes, err := json.Marshal(email)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	return e.emailsPublisher.Publish(mailBytes, email.ContentType)
}

// FInd email by id
func (e *EmailUseCase) FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.FindEmailById")
	defer span.Finish()

	return e.emailRepo.FindEmailById(ctx, emailID)
}

// Find email by id
func (e *EmailUseCase) FindEmailsByReceiver(ctx context.Context, receiverEmail string, paginationQuery *utils.PaginationQuery) (*models.EmailsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.FindEmailsByReceiver")
	defer span.Finish()

	return e.emailRepo.FindEmailsByReceiver(ctx, receiverEmail, paginationQuery)
}
