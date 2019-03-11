package liberr

import "errors"

func NewInvalidArgument(reason string) *InvalidArgument {
	return &InvalidArgument{errors.New(reason)}
}

type InvalidArgument struct {
	error
}

func (e *InvalidArgument) Error() string {
	return e.error.Error()
}
