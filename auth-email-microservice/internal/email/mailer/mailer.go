package mailer

import (
	"context"

	"github.com/Chuuch/ecom-microservices/internal/models"
	"github.com/opentracing/opentracing-go"
	"github.com/resend/resend-go/v2"
)

// Mailer struct
type Mailer struct {
	resendClient *resend.Client
}

// New Mailer
func NewMailer(resendClient *resend.Client) *Mailer {
	return &Mailer{
		resendClient: resendClient,
	}
}

// Send email
func (m *Mailer) Send(ctx context.Context, email *models.Email) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Mailer.Send")
	defer span.Finish()

	params := &resend.SendEmailRequest{
		From:    email.From,
		To:      email.To,
		Subject: email.Subject,
		Html:    email.Body,
	}

	_, err := m.resendClient.Emails.Send(params)
	return err
}
