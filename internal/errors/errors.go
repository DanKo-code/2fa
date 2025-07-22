package errors

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

type ErrorType struct {
	t string
}

type CompositeError struct {
	external  error
	inner     error
	errorType ErrorType
}

func (c CompositeError) Error() string {
	return c.inner.Error()
}

func (c CompositeError) External() string {
	var comp CompositeError
	var msg string
	ok := errors.As(c.inner, &comp)
	if !ok {
		if c.external != nil {
			msg = c.external.Error()
		} else {
			msg = c.inner.Error()
		}
	} else {
		msg = comp.External()
	}

	r, size := utf8.DecodeRuneInString(msg)
	if r == utf8.RuneError {
		return msg
	}
	return string(unicode.ToUpper(r)) + msg[size:]
}

func (c CompositeError) ErrorType() ErrorType {
	var comp CompositeError
	ok := errors.As(c.inner, &comp)
	if ok {
		return comp.ErrorType()
	}
	return c.errorType
}

func NewCompositeWithContext(context string, inner error, external string) CompositeError {
	return CompositeError{
		inner:     Wrap(context, inner),
		external:  errors.New(external),
		errorType: ErrorTypeUnknown,
	}
}

func NewRecordNotFound(context string, inner error, external string) CompositeError {
	return CompositeError{
		inner:     Wrap(context, inner),
		external:  errors.New(external),
		errorType: ErrorTypeRecordNotFound,
	}
}

func NewRecordAlreadyExists(context string, inner error, external string) CompositeError {
	return CompositeError{
		inner:     Wrap(context, inner),
		external:  errors.New(external),
		errorType: ErrorTypeRecordAlreadyExists,
	}
}

func NewIncorrectInputError(context string, inner error, external string) CompositeError {
	return CompositeError{
		inner:     Wrap(context, inner),
		external:  errors.New(external),
		errorType: ErrorTypeIncorrectInput,
	}
}

func NewUnauthorized(context string, inner error, external string) CompositeError {
	return CompositeError{
		inner:     Wrap(context, inner),
		external:  errors.New(external),
		errorType: ErrorTypeUnauthorized,
	}
}

var (
	ErrorTypeUnknown             = ErrorType{"unknown"}
	ErrorTypeUnauthorized        = ErrorType{"unauthorized"}
	ErrorTypeIncorrectInput      = ErrorType{"incorrectInput"}
	ErrorTypeRecordNotFound      = ErrorType{"record not found"}
	ErrorTypeRecordAlreadyExists = ErrorType{"record already exists"}

	ErrUnauthorized                 = errors.New("unauthorized")
	ErrFailedSentFirstFactor        = errors.New("failed sent first factor")
	ErrNotFound                     = errors.New("not found")
	ErrUnknown                      = errors.New("unknown")
	ErrFirstFactorCodeAlreadyExist  = errors.New("code already exist")
	ErrSecondFactorCodeAlreadyExist = errors.New("code already exist")
	ErrPushRejected                 = errors.New("push rejected")
	ErrBadParams                    = errors.New("bad params")
)

func Wrap(context string, err error) error {
	return fmt.Errorf("%s: %w", context, err)
}
