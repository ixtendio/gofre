package response

import (
	"context"
	"errors"
	"github.com/ixtendio/gofre/router/path"

	"net/http"
	"strconv"
	"strings"
)

const headerLastEventId = "Last-Event-Id"
const eventStreamContentType = "text/event-stream"

var ErrNotHttp2Request = errors.New("rejected, not a HTTP/2 request")

// ServerSentEvent defines the server-sent event fields. More about server-sent events can be found here: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
type ServerSentEvent struct {
	// A string identifying the type of event described
	Name string
	// The event ID to set the EventSource object's last event ID value
	Id string
	// The data field for the message
	Data []string
	// The reconnection time in millis. If the connection to the server is lost, the browser will wait for the specified time before attempting to reconnect.
	Retry int
}

func (evt ServerSentEvent) String() string {
	var startDataWithNewLine bool
	var sb strings.Builder
	if evt.Name != "" {
		sb.WriteString("event: ")
		sb.WriteString(evt.Name)
		startDataWithNewLine = true
	}
	for _, data := range evt.Data {
		if !startDataWithNewLine {
			sb.WriteString("data: ")
			startDataWithNewLine = true
		} else {
			sb.WriteString("\ndata: ")
		}
		sb.WriteString(data)
	}
	if evt.Id != "" {
		sb.WriteString("\nid: ")
		sb.WriteString(evt.Id)
	}
	if evt.Retry > 0 {
		sb.WriteString("\nretry: ")
		sb.WriteString(strconv.Itoa(evt.Retry))
	}
	if sb.Len() > 0 {
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// The EventGenerator is a function that received a cancellation context.Context and the last event id received by the client and returns a channel of ServerSentEvent objects
type EventGenerator func(ctx context.Context, lastEventId string) <-chan ServerSentEvent

// HttpSSEResponse implements response.HttpResponse and provides support for SSE
type HttpSSEResponse struct {
	HttpHeadersResponse
	EventGenerator EventGenerator
}

func (r *HttpSSEResponse) Write(w http.ResponseWriter, mc path.MatchingContext) error {
	if mc.R.ProtoMajor != 2 {
		w.WriteHeader(http.StatusInternalServerError)
		return ErrNotHttp2Request
	}

	// write the headers
	if err := r.HttpHeadersResponse.Write(w, mc); err != nil {
		return err
	}

	header := w.Header()
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")

	defer r.flushResponse(w)
	// get the last event id that was sent
	lastEventId := mc.R.Header.Get(headerLastEventId)
	reqCtx := mc.R.Context()

	// read and write the events
	for evt := range r.EventGenerator(reqCtx, lastEventId) {
		select {
		case <-reqCtx.Done():
			return nil
		default:
			if err := writeTextResponse(w, evt.String()); err != nil {
				return nil
			}
			r.flushResponse(w)
		}
	}

	return nil
}

func (r *HttpSSEResponse) flushResponse(w http.ResponseWriter) {
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// SSEHttpResponse creates a SSE response
func SSEHttpResponse(ew EventGenerator) *HttpSSEResponse {
	return SSEHttpResponseWithHeadersAndCookies(ew, nil, nil)
}

// SSEHttpResponseWithHeaders creates a SSE response with custom headers
func SSEHttpResponseWithHeaders(ew EventGenerator, headers HttpHeaders) *HttpSSEResponse {
	return SSEHttpResponseWithHeadersAndCookies(ew, headers, nil)
}

// SSEHttpResponseWithHeadersAndCookies creates a SSE response with custom headers and cookies
// The headers and cookies, if present, once will be written to output will be added in the pool for re-use
func SSEHttpResponseWithHeadersAndCookies(ew EventGenerator, headers HttpHeaders, cookies HttpCookies) *HttpSSEResponse {
	return &HttpSSEResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusOK,
			ContentType:    eventStreamContentType,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		EventGenerator: ew,
	}
}
