package response

import (
	"context"
	"errors"
	"github.com/ixtendio/gow/request"
	"net/http"
	"strconv"
	"strings"
)

const headerLastEventId = "Last-Event-Id"

var errNotHttp2Request = errors.New("rejected, not a HTTP/2 request")

// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
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
	var sb strings.Builder
	sb.WriteString("event: ")
	sb.WriteString(evt.Name)
	for _, data := range evt.Data {
		sb.WriteString("\ndata: ")
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
	sb.WriteString("\n\n")
	return sb.String()
}

type EventGenerator func(ctx context.Context, lastEventId string) <-chan ServerSentEvent

type HttpSSEResponse struct {
	HttpHeadersResponse
	EventGenerator EventGenerator
}

func (r *HttpSSEResponse) Write(w http.ResponseWriter, reqContext *request.HttpRequest) error {
	if reqContext.RawRequest.ProtoMajor != 2 {
		w.WriteHeader(http.StatusInternalServerError)
		return errNotHttp2Request
	}

	// get the last event id that was sent
	lastEventId := reqContext.RawRequest.Header.Get(headerLastEventId)

	// write the headers
	if err := r.HttpHeadersResponse.Write(w, reqContext); err != nil {
		return err
	}
	reqCtx := reqContext.RawRequest.Context()
	ctx, cancelFunc := context.WithCancel(reqCtx)
	defer cancelFunc()

	// call the event generator
	evtChan := r.EventGenerator(ctx, lastEventId)

	// read and write the events
	for {
		select {
		case <-reqCtx.Done():
			return nil
		case evt := <-evtChan:
			if _, err := w.Write([]byte(evt.String())); err != nil {
				return nil
			}
			flusher, ok := w.(http.Flusher)
			if ok {
				flusher.Flush()
			}
		}
	}
}

// SSEHttpResponse creates a SSE response
func SSEHttpResponse(ew EventGenerator) *HttpSSEResponse {
	return &HttpSSEResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusOK,
			HttpHeaders: map[string]string{
				"Cache-Control": "no-cache",
				"Content-Type":  "text/event-stream",
			},
		},
		EventGenerator: ew,
	}
}

// SSEHttpResponseWithHeaders creates a SSE response with custom headers
func SSEHttpResponseWithHeaders(ew EventGenerator, headers map[string]string) *HttpSSEResponse {
	headers["Content-Type"] = "text/event-stream"
	return &HttpSSEResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusOK,
			HttpHeaders:    headers,
		},
		EventGenerator: ew,
	}
}

// SSEHttpResponseWithHeadersAndCookies creates a SSE response with custom headers and cookies
func SSEHttpResponseWithHeadersAndCookies(ew EventGenerator, headers map[string]string, cookies []*http.Cookie) *HttpSSEResponse {
	headers["Content-Type"] = "text/event-stream"
	return &HttpSSEResponse{
		HttpHeadersResponse: HttpHeadersResponse{
			HttpStatusCode: http.StatusOK,
			HttpHeaders:    headers,
			HttpCookies:    cookies,
		},
		EventGenerator: ew,
	}
}
