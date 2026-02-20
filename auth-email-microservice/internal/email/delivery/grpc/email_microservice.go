package grpc

import (
	"context"

	"github.com/Chuuch/ecom-microservices/config"
	"github.com/Chuuch/ecom-microservices/internal/email"
	"github.com/Chuuch/ecom-microservices/internal/models"
	grpcerrors "github.com/Chuuch/ecom-microservices/pkg/grpc_errors"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/Chuuch/ecom-microservices/pkg/utils"
	userService "github.com/Chuuch/ecom-microservices/proto"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Email gRPC microservice
type EmailMicroservice struct {
	userService.UnimplementedEmailServiceServer
	emailUC email.EmailUseCase
	cfg     *config.Config
	logger  logger.Logger
}

// Email gRPC microservice constructor
func NewEmailMicroservice(emailUC email.EmailUseCase, cfg *config.Config, logger logger.Logger) *EmailMicroservice {
	return &EmailMicroservice{
		emailUC: emailUC,
		cfg:     cfg,
		logger:  logger,
	}
}

// Send email
func (e *EmailMicroservice) SendEmail(ctx context.Context, req *userService.SendEmailRequest) (*userService.SendEmailResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailMicroservice.SendEmail")
	defer span.Finish()

	mail := &models.Email{
		From:    "daniel@skyeystudio.com",
		To:      req.GetTo(),
		Subject: req.GetSubject(),
		Body:    req.GetBody(),
	}

	if err := mail.PrepareAndValidate(ctx); err != nil {
		e.logger.Errorf("prepareAndValidate: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "prepareAndValidate: %v", err)
	}

	if err := e.emailUC.PublishEmail(ctx, mail); err != nil {
		e.logger.Errorf("emailUC.PublishEmail: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "emailUC.PublishEmail: %v", err)
	}

	return &userService.SendEmailResponse{
		Status: "OK",
	}, nil
}

func (e *EmailMicroservice) FindEmailById(ctx context.Context, req *userService.FindEmailByIdRequest) (*userService.FindEmailByIdResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailMicroservice.FindEmailById")
	defer span.Finish()

	emailUUID, err := uuid.Parse(req.GetEmailUuid())
	if err != nil {
		e.logger.Errorf("parseUUID: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "parseUUID: %v", err)
	}

	emailById, err := e.emailUC.FindEmailById(ctx, emailUUID)
	if err != nil {
		e.logger.Errorf("emailUC.FindEmailById: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "emailUC.FindEmailById: %v", err)
	}

	return &userService.FindEmailByIdResponse{Email: e.convertEmailToProto(emailById)}, nil
}

func (e *EmailMicroservice) FindEmailsByReceiver(ctx context.Context, req *userService.FindEmailsByReceiverRequest) (*userService.FindEmailsByReceiverResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailMicroservice.FindEmailsByReceiver")
	defer span.Finish()

	paginationQuery := &utils.PaginationQuery{
		Size: req.GetSize(),
		Page: req.GetPage(),
	}

	emails, err := e.emailUC.FindEmailsByReceiver(ctx, req.GetReceiverEmail(), paginationQuery)
	if err != nil {
		e.logger.Errorf("emailUC.FindEmailsByReceiver: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "emailUC.FindEmailsByReceiver: %v", err)
	}

	return &userService.FindEmailsByReceiverResponse{
		Email:      e.convertEmailsListToProto(emails.Emails),
		TotalPages: emails.TotalPages,
		TotalCount: emails.TotalCount,
		HasMore:    emails.HasMore,
		Page:       emails.Page,
		Size:       emails.Size,
	}, nil
}

func (e *EmailMicroservice) convertEmailToProto(email *models.Email) *userService.Email {
	return &userService.Email{
		EmailId:     email.EmailID.String(),
		To:          email.To,
		From:        email.From,
		Body:        email.Body,
		Subject:     email.Subject,
		ContentType: email.ContentType,
		CreatedAt:   timestamppb.New(email.CreatedAt),
	}
}

func (e *EmailMicroservice) convertEmailsListToProto(emails []*models.Email) []*userService.Email {
	protoEmails := make([]*userService.Email, 0, len(emails))
	for _, email := range emails {
		protoEmails = append(protoEmails, e.convertEmailToProto(email))
	}
	return protoEmails
}
