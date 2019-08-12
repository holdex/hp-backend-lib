package liberr

import (
	"errors"
	"fmt"
)

func NewNotAuthorized(reason string) *NotAuthorized {
	return &NotAuthorized{errors.New(reason)}
}

type NotAuthorized struct {
	error
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

func (e *Unauthenticated) Error() string {
	return e.error.Error()
}
