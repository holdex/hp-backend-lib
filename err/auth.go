package liberr

import "errors"

func NewNotAuthorized(reason string) *NotAuthorized {
	return &NotAuthorized{errors.New(reason)}
}

type NotAuthorized struct {
	error
}

func (e *NotAuthorized) Error() string {
	return e.error.Error()
}

func NewNotAuthenticated() *NotAuthenticated {
	return &NotAuthenticated{}
}

type NotAuthenticated struct {
}

func (e *NotAuthenticated) Error() string {
	return "not authenticated"
}
