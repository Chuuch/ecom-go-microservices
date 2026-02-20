package mailer

import (
	"github.com/Chuuch/ecom-microservices/config"
	"github.com/resend/resend-go/v2"
)

// New mailer dialer
func NewResendClient(cfg *config.Config) *resend.Client {
	return resend.NewClient(cfg.Resend.ApiKey)
}
