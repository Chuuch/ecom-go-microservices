package server

import (
	"github.com/Chuuch/ecom-microservices/config"
	"github.com/Chuuch/ecom-microservices/internal/email"
	"github.com/Chuuch/ecom-microservices/internal/session"
	"github.com/Chuuch/ecom-microservices/internal/user"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/Chuuch/ecom-microservices/pkg/metric"
	userService "github.com/Chuuch/ecom-microservices/proto"
)

type usersService struct {
	userService.UnimplementedUserServiceServer
	logger    logger.Logger
	cfg       *config.Config
	userUC    user.UserUseCase
	sessionUC session.SessionUseCase
	metrics   metric.Metrics
	emailUC   email.EmailUseCase
}

// Auth server constructor

func NewAuthServerGRPC(
	logger logger.Logger,
	cfg *config.Config,
	userUC user.UserUseCase,
	sessionUC session.SessionUseCase,
	metrics metric.Metrics,
	emailUC email.EmailUseCase,
) *usersService {
	return &usersService{
		logger:    logger,
		cfg:       cfg,
		userUC:    userUC,
		sessionUC: sessionUC,
		metrics:   metrics,
		emailUC:   emailUC,
	}
}
