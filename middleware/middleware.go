package middleware

import (
	"github.com/ixtendio/gofre/handler"
)

// A Middleware is a function that receives a NewMuxHandler and returns another NewMuxHandler
type Middleware func(handler handler.Handler) handler.Handler
