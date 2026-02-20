package server

import (
	"context"
	"errors"

	"github.com/Chuuch/ecom-microservices/internal/models"
	grpcerrors "github.com/Chuuch/ecom-microservices/pkg/grpc_errors"
	"github.com/Chuuch/ecom-microservices/pkg/utils"
	userService "github.com/Chuuch/ecom-microservices/proto"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Register a new user
func (u *usersService) Register(ctx context.Context, r *userService.RegisterRequest) (*userService.RegisterResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "user.Register")
	defer span.Finish()

	user, err := u.registerRequestToUser(r)
	if err != nil {
		u.logger.Error("registerRequestToUser: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "registerRequestToUser: %v", err)
	}

	if err := utils.ValidateStruct(ctx, user); err != nil {
		u.logger.Errorf("validateStruct: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "validateStruct: %v", err)
	}

	createdUser, err := u.userUC.Register(ctx, user)

	if err != nil {
		u.logger.Error("userUC.Register: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "userUC.Register: %v", err)
	}

	return &userService.RegisterResponse{
		User: u.userToProto(createdUser),
	}, nil
}

// Find user by email address
func (u *usersService) FindByEmail(ctx context.Context, r *userService.FindByEmailRequest) (*userService.FindByEmailResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "user.FindByEmail")
	defer span.Finish()

	email := r.GetEmail()

	if !utils.ValidateEmail(email) {
		u.logger.Errorf("validateEmail: %v", email)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(errors.New("Invalid email address")), "validateEmail: %v", email)
	}

	user, err := u.userUC.FindByEmail(ctx, email)
	if err != nil {
		u.logger.Error("userUC.FindByEmail: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "userUC.FindByEmail: %v", err)
	}

	return &userService.FindByEmailResponse{
		User: u.userToProto(user),
	}, nil
}

// Find user by id
func (u *usersService) FindByID(ctx context.Context, r *userService.FindByIDRequest) (*userService.FindByIDResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "user.FindByID")
	defer span.Finish()

	userID, err := uuid.Parse(r.GetUserId())
	if err != nil {
		u.logger.Errorf("parseUserID: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "parseUserID: %v", err)
	}

	user, err := u.userUC.FindByID(ctx, userID)
	if err != nil {
		u.logger.Error("userUC.FindByID: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "userUC.FindByID: %v", err)
	}

	return &userService.FindByIDResponse{
		User: u.userToProto(user),
	}, nil
}

// Login user with email and password
func (u *usersService) Login(ctx context.Context, r *userService.LoginRequest) (*userService.LoginResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "user.Login")
	defer span.Finish()

	email := r.GetEmail()
	if !utils.ValidateEmail(email) {
		u.logger.Errorf("validateEmail: %v", email)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(errors.New("Invalid email address")), "validateEmail: %v", email)
	}

	user, err := u.userUC.Login(ctx, email, r.GetPassword())
	if err != nil {
		u.logger.Error("userUC.Login: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "userUC.Login: %v", err)
	}

	session, err := u.sessionUC.CreateSession(ctx, &models.Session{
		UserID: user.UserID.String(),
	}, u.cfg.Session.Expire)
	if err != nil {
		u.logger.Error("sessionUC.CreateSession: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "sessionUC.CreateSession: %v", err)
	}

	return &userService.LoginResponse{
		User:      u.userToProto(user),
		SessionId: session,
	}, nil
}

// Get session id from ctx metadata, find user by uuid and return it
func (u *usersService) GetMe(ctx context.Context, r *userService.GetMeRequest) (*userService.GetMeResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "user.GetMe")
	defer span.Finish()

	sessionID, err := u.getSessionIDFromCtx(ctx)
	if err != nil {
		u.logger.Error("getSessionIDFromCtx: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "getSessionIDFromCtx: %v", err)
	}

	session, err := u.sessionUC.GetSessionByID(ctx, sessionID)
	if err != nil {
		u.logger.Errorf("sessionUC.GetSessionByID: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "sessionUC.GetSessionByID: %v", err)
	}

	user, err := u.userUC.FindByID(ctx, uuid.MustParse(session.UserID))
	if err != nil {
		u.logger.Errorf("userUC.FindByID: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "userUC.FindByID: %v", err)
	}

	return &userService.GetMeResponse{
		User: u.userToProto(user),
	}, nil
}

func (u *usersService) registerRequestToUser(r *userService.RegisterRequest) (*models.User, error) {
	candidate := &models.User{
		Email:     r.GetEmail(),
		FirstName: r.GetFirstName(),
		LastName:  r.GetLastName(),
		Password:  r.GetPassword(),
		Role:      r.GetRole(),
		Avatar:    r.GetAvatar(),
	}

	if err := candidate.PrepareRegister(); err != nil {
		return nil, err
	}

	return candidate, nil
}

// Logout user
func (u *usersService) Logout(ctx context.Context, r *userService.LogoutRequest) (*userService.LogoutResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "user.Logout")
	defer span.Finish()

	sessionID, err := u.getSessionIDFromCtx(ctx)
	if err != nil {
		u.logger.Error("getSessionIDFromCtx: %v", err)
		return nil, err
	}

	if err := u.sessionUC.DeleteSession(ctx, sessionID); err != nil {
		u.logger.Error("sessionUC.DeleteByID: %v", err)
		return nil, status.Errorf(grpcerrors.ParseGRPCError(err), "sessionUC.DeleteByID: %v", err)
	}

	return &userService.LogoutResponse{}, nil
}

func (u *usersService) getSessionIDFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "No ctx metadata")
	}

	sessionID := md.Get("session_id")
	if len(sessionID) == 0 {
		return "", status.Error(codes.PermissionDenied, "No session id in ctx metadata")
	}
	return sessionID[0], nil
}

func (u *usersService) userToProto(user *models.User) *userService.User {
	userProto := &userService.User{
		Uuid:      user.UserID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
		Avatar:    user.Avatar,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	return userProto
}
