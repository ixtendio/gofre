package errors

import "errors"

var ErrDenied = errors.New("denied request")
var ErrUnauthorized = errors.New("unauthorized request")
var ErrWrongCredentials = errors.New("wrong credentials")

type ErrInvalidRequest struct {
	err error
}

func (e ErrInvalidRequest) Error() string {
	return e.err.Error()
}

func NewErrInvalidRequestWithMessage(msg string) ErrInvalidRequest {
	return ErrInvalidRequest{
		err: errors.New(msg),
	}
}

func NewErrInvalidRequest(err error) ErrInvalidRequest {
	return ErrInvalidRequest{
		err: err,
	}
}

type ErrObjectNotFound struct {
	err error
}

func (e ErrObjectNotFound) Error() string {
	return e.err.Error()
}

func NewErrObjectNotFoundWithMessage(msg string) ErrObjectNotFound {
	return ErrObjectNotFound{
		err: errors.New(msg),
	}
}

func NewErrObjectNotFound(err error) ErrObjectNotFound {
	return ErrObjectNotFound{
		err: err,
	}
}
