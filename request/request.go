package request

import (
	"net/http"
)

type HttpRequest struct {
	R       *http.Request
	UriVars map[string]string
}
