package common

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConvertGRPCError converts a GRPC error into a business error
func ConvertGRPCError(err error) error {
	s, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch s.Code() {
	case codes.AlreadyExists:
		return ErrAlreadyExists{Err: err}
	case codes.NotFound:
		return ErrNotFound{Err: err}
	default:
		return err
	}
}
