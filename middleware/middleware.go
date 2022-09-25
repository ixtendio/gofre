package middleware

import "github.com/ixtendio/gow/router"

// A Middleware is a function that receives a NewMuxHandler and returns another NewMuxHandler
type Middleware func(handler router.Handler) router.Handler
