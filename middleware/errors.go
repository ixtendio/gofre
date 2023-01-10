package middleware

import (
	"context"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/router/path"

	"github.com/ixtendio/gofre/response"
	"net/http"
)

type ResponseSupplier func(statusCode int, err error) response.HttpResponse

// Error2HttpStatusCode translates an error to an HTTP status code
var Error2HttpStatusCode = func(err error) int {
	if _, ok := err.(errors.ErrBadRequest); ok {
		return http.StatusBadRequest
	} else if _, ok := err.(errors.ErrObjectNotFound); ok {
		return http.StatusNotFound
	} else if err == errors.ErrUnauthorizedRequest {
		return http.StatusUnauthorized
	} else if err == errors.ErrWrongCredentials ||
		err == errors.ErrAccessDenied {
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}

// ErrJsonResponse translates an error to a JSON response
func ErrJsonResponse() Middleware {
	return ErrResponse(func(statusCode int, err error) response.HttpResponse {
		return response.JsonHttpResponse(statusCode, map[string]string{
			"error": err.Error(),
		})
	})
}

// ErrResponse translates an error to an response.HttpResponse
func ErrResponse(responseSupplier ResponseSupplier) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, mc path.MatchingContext) (response.HttpResponse, error) {
			resp, err := handler(ctx, mc)
			if err != nil {
				statusCode := Error2HttpStatusCode(err)
				return responseSupplier(statusCode, err), nil
			}
			return resp, err
		}
	}
}
