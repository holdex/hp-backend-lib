package liberr

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
)

func NewNotAuthorized(reason string) *NotAuthorized {
	return &NotAuthorized{errors.New(reason)}
}

type NotAuthorized struct {
	error
}

func (e *NotAuthorized) Code() string {
	return codes.PermissionDenied.String()
}

func (e *NotAuthorized) Error() string {
	return e.error.Error()
}

func NewUnauthenticated(reason string, args ...interface{}) *Unauthenticated {
	return &Unauthenticated{fmt.Errorf(reason, args...)}
}

type Unauthenticated struct {
	error
}

func (e *Unauthenticated) Code() string {
	return codes.Unauthenticated.String()
}

func (e *Unauthenticated) Error() string {
	return e.error.Error()
}
