package liberr

import (
	"errors"

	"google.golang.org/grpc/codes"
)

func NewInvalidArgument(reason string) *InvalidArgument {
	return &InvalidArgument{errors.New(reason)}
}

type InvalidArgument struct {
	error
}

func (e *InvalidArgument) Code() string {
	return codes.InvalidArgument.String()
}

func (e *InvalidArgument) Error() string {
	return e.error.Error()
}
