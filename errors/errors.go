package errors

import "errors"

var ErrAccessDenied = errors.New("access denied")
var ErrUnauthorizedRequest = errors.New("unauthorized request")
var ErrWrongCredentials = errors.New("wrong credentials")

type ErrBadRequest struct {
	err error
}

func (e ErrBadRequest) Error() string {
	return e.err.Error()
}

func NewBadRequestWithMessage(msg string) ErrBadRequest {
	return ErrBadRequest{
		err: errors.New(msg),
	}
}

func NewBadRequest(err error) ErrBadRequest {
	return ErrBadRequest{
		err: err,
	}
}

type ErrObjectNotFound struct {
	err error
}

func (e ErrObjectNotFound) Error() string {
	return e.err.Error()
}

func NewObjectNotFoundWithMessage(msg string) ErrObjectNotFound {
	return ErrObjectNotFound{
		err: errors.New(msg),
	}
}

func NewObjectNotFound(err error) ErrObjectNotFound {
	return ErrObjectNotFound{
		err: err,
	}
}
