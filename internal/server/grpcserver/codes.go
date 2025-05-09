package grpcserver

import (
	"google.golang.org/grpc/codes"
	"schedule/pkg/failure"
)

func getCodeFromError(err error) codes.Code {
	switch {
	case failure.IsInternalError(err):
		return codes.Internal
	case failure.IsNotFoundError(err):
		return codes.NotFound
	case failure.IsInvalidRequestError(err):
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}
