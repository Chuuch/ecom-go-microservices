package grpcerrors

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound           = errors.New("Not found")
	ErrNoCtxMetadata      = errors.New("No ctx metadata")
	ErrInvalidSessionId   = errors.New("Invalid session id")
	ErrEmailAlreadyExists = errors.New("Email already exists")
)

// Parse error and get code
func ParseGRPCError(err error) codes.Code {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return codes.NotFound
	case errors.Is(err, context.Canceled):
		return codes.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return codes.DeadlineExceeded
	case errors.Is(err, ErrEmailAlreadyExists):
		return codes.AlreadyExists
	case errors.Is(err, ErrNoCtxMetadata):
		return codes.Unauthenticated
	case errors.Is(err, ErrInvalidSessionId):
		return codes.PermissionDenied
	case strings.Contains(err.Error(), "Validate"):
		return codes.InvalidArgument
	case strings.Contains(err.Error(), "redis"):
		return codes.NotFound
	case errors.Is(err, context.Canceled):
		return codes.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return codes.DeadlineExceeded
	}
	return codes.Internal
}

// Map grpc errors to http status codes
func MapGRPCErrStatusCodeToHttpStatus(code codes.Code) int {
	switch code {
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.AlreadyExists:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.InvalidArgument:
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}

// Error Response GRPC error response
func ErrorResponse(err error, msg string) error {
	return status.Errorf(ParseGRPCError(err), fmt.Sprintf("%s: %v", msg, err))
}
