package apperror

import (
	"errors"
	"fmt"
)

type Kind uint8

const (
	KindUnknown Kind = iota
	KindValidation
	KindNotFound
	KindConflict
	KindUnauthorized
	KindForbidden
	KindInternal
	KindExternal
	KindTimeout
)

const unknownErrorStr = "UNKNOWN_ERROR"

func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return unknownErrorStr
	case KindValidation:
		return "VALIDATION_ERROR"
	case KindNotFound:
		return "NOT_FOUND"
	case KindConflict:
		return "CONFLICT"
	case KindUnauthorized:
		return "UNAUTHORIZED"
	case KindForbidden:
		return "FORBIDDEN"
	case KindInternal:
		return "INTERNAL_ERROR"
	case KindExternal:
		return "EXTERNAL_SERVICE_ERROR"
	case KindTimeout:
		return "TIMEOUT"
	}

	return unknownErrorStr
}

type Error struct {
	kind    Kind
	code    string
	message string
	op      string
	err     error
	fields  map[string]string
}

func (e *Error) Error() string {
	if e.op != "" {
		return fmt.Sprintf("%s: %s", e.op, e.message)
	}

	return e.message
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Kind() Kind {
	return e.kind
}

func (e *Error) Code() string {
	return e.code
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) Op() string {
	return e.op
}

func (e *Error) Fields() map[string]string {
	return e.fields
}

type Option func(*Error)

func WithOp(op string) Option {
	return func(e *Error) {
		e.op = op
	}
}

func WithCode(code string) Option {
	return func(e *Error) {
		e.code = code
	}
}

func WithCause(err error) Option {
	return func(e *Error) {
		e.err = err
	}
}

func WithField(key, value string) Option {
	return func(e *Error) {
		if e.fields == nil {
			e.fields = make(map[string]string)
		}

		e.fields[key] = value
	}
}

func WithFields(fields map[string]string) Option {
	return func(e *Error) {
		e.fields = fields
	}
}

func newError(kind Kind, message string, opts ...Option) *Error {
	e := &Error{
		kind:    kind,
		message: message,
		code:    kind.String(),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func Validation(message string, opts ...Option) *Error {
	return newError(KindValidation, message, opts...)
}

func NotFound(message string, opts ...Option) *Error {
	return newError(KindNotFound, message, opts...)
}

func Conflict(message string, opts ...Option) *Error {
	return newError(KindConflict, message, opts...)
}

func Unauthorized(message string, opts ...Option) *Error {
	return newError(KindUnauthorized, message, opts...)
}

func Forbidden(message string, opts ...Option) *Error {
	return newError(KindForbidden, message, opts...)
}

func Internal(message string, opts ...Option) *Error {
	return newError(KindInternal, message, opts...)
}

func External(message string, opts ...Option) *Error {
	return newError(KindExternal, message, opts...)
}

func Timeout(message string, opts ...Option) *Error {
	return newError(KindTimeout, message, opts...)
}

func GetKind(err error) Kind {
	if err == nil {
		return KindUnknown
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Kind()
	}

	return KindUnknown
}

func Is(err error, kind Kind) bool {
	return GetKind(err) == kind
}

func IsNotFound(err error) bool {
	return Is(err, KindNotFound)
}

func IsValidation(err error) bool {
	return Is(err, KindValidation)
}

func IsConflict(err error) bool {
	return Is(err, KindConflict)
}

func IsUnauthorized(err error) bool {
	return Is(err, KindUnauthorized)
}

func IsForbidden(err error) bool {
	return Is(err, KindForbidden)
}

func IsInternal(err error) bool {
	return Is(err, KindInternal)
}

func IsExternal(err error) bool {
	return Is(err, KindExternal)
}

func IsTimeout(err error) bool {
	return Is(err, KindTimeout)
}

func AsAppError(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}

	return nil, false
}
