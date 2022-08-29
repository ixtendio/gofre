package request

import (
	"net/http"
)

type HttpRequest struct {
	RawRequest *http.Request
	PathVars   map[string]string
}
