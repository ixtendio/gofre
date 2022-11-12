package middleware

import (
	"context"
	"github.com/ixtendio/gofre/errors"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"net/http"
)

type ResponseSupplier func(statusCode int, err error) response.HttpResponse

func ErrJsonResponse() Middleware {
	return ErrResponse(func(statusCode int, err error) response.HttpResponse {
		return response.JsonHttpResponse(statusCode, map[string]string{
			"error": err.Error(),
		})
	})
}

func ErrResponse(responseSupplier ResponseSupplier) Middleware {
	return func(handler handler.Handler) handler.Handler {
		return func(ctx context.Context, req request.HttpRequest) (response.HttpResponse, error) {
			resp, err := handler(ctx, req)
			if err != nil {
				statusCode := http.StatusInternalServerError
				if _, ok := err.(errors.ErrBadRequest); ok {
					statusCode = http.StatusBadRequest
				} else if _, ok := err.(errors.ErrObjectNotFound); ok {
					statusCode = http.StatusNotFound
				} else if err == errors.ErrUnauthorizedRequest {
					statusCode = http.StatusUnauthorized
				} else if err == errors.ErrWrongCredentials ||
					err == errors.ErrAccessDenied {
					statusCode = http.StatusForbidden
				}
				return responseSupplier(statusCode, err), nil
			}
			return resp, err
		}
	}
}
