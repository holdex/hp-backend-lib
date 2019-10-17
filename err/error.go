package liberr

import (
	"errors"
	"fmt"
	"sync"
)

type Code interface {
	Code() string
}

type Error struct {
	code string
}

func (e *Error) Code() string {
	return e.code
}

func (e *Error) Error() string {
	return "CODE=" + e.code
}

func New(code string) error {
	return &Error{
		code: code,
	}
}

type Errors struct {
	m      sync.RWMutex
	Errors []error
}

func Group(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	var e string
	for _, err := range errs {
		e = fmt.Sprintln(e, err.Error())
	}
	return errors.New(e)
}

func Append(errs []error, err error) []error {
	if err != nil {
		return append(errs, err)
	}
	return errs
}

func (e *Errors) AppendMutex(err error) {
	if err == nil {
		return
	}
	e.m.Lock()
	e.Errors = append(e.Errors, err)
	e.m.Unlock()
}
