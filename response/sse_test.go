package response

import (
	"context"
	"github.com/ixtendio/gofre/request"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

var nilEventGen = func(ctx context.Context, lastEventId string) <-chan ServerSentEvent {
	return nil
}

var testEventGen = func(startEventIndex int, done func()) EventGenerator {
	events := []ServerSentEvent{
		{
			Name:  "message",
			Id:    "1",
			Data:  []string{"msg1"},
			Retry: 0,
		},
		{
			Name:  "message",
			Id:    "2",
			Data:  []string{"msg2"},
			Retry: 0,
		},
		{
			Name:  "message",
			Id:    "3",
			Data:  []string{"msg3"},
			Retry: 0,
		},
	}
	return func(ctx context.Context, lastEventId string) <-chan ServerSentEvent {
		lastEventIdAsNumber, err := strconv.Atoi(lastEventId)
		if err == nil {
			startEventIndex = lastEventIdAsNumber
		}
		ch := make(chan ServerSentEvent)
		go func() {
			defer close(ch)
			for i := startEventIndex; i < len(events); i++ {
				ch <- events[i]
			}
			done()
		}()
		return ch
	}
}

func TestServerSentEvent_String(t *testing.T) {
	type args struct {
		Name  string
		Id    string
		Data  []string
		Retry int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "",
			args: args{},
			want: "",
		},
		{
			name: "with name",
			args: args{
				Name: "name",
			},
			want: "event: name\n\n",
		},
		{
			name: "with all fields",
			args: args{
				Name:  "name",
				Id:    "id",
				Data:  []string{"val1", "val2"},
				Retry: 5,
			},
			want: "event: name\ndata: val1\ndata: val2\nid: id\nretry: 5\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := ServerSentEvent{
				Name:  tt.args.Name,
				Id:    tt.args.Id,
				Data:  tt.args.Data,
				Retry: tt.args.Retry,
			}
			if got := evt.String(); got != tt.want {
				t.Errorf("ServerSentEvent.String() got: `%v`, want: `%v`", got, tt.want)
			}
		})
	}
}

func TestSSEHttpResponse(t *testing.T) {
	type args struct {
		ew EventGenerator
	}
	tests := []struct {
		name string
		args args
		want *HttpSSEResponse
	}{
		{
			name: "constructor",
			args: args{ew: nilEventGen},
			want: &HttpSSEResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 200,
					HttpHeaders:    http.Header{"Content-Type": {"text/event-stream"}, "Cache-Control": {"no-cache"}, "Connection": {"keep-alive"}},
					HttpCookies:    NewHttpCookies(nil),
				},
				EventGenerator: nilEventGen,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SSEHttpResponse(tt.args.ew)
			if !reflect.DeepEqual(got.HttpHeadersResponse, tt.want.HttpHeadersResponse) {
				t.Fatalf("SSEHttpResponse() = %v, want %v", got, tt.want)
			}
			if got.EventGenerator == nil {
				t.Fatalf("SSEHttpResponse() EventGenerator is nil")
			}
		})
	}
}

func TestSSEHttpResponseWithHeaders(t *testing.T) {
	type args struct {
		ew      EventGenerator
		headers http.Header
	}
	tests := []struct {
		name string
		args args
		want *HttpSSEResponse
	}{
		{
			name: "constructor",
			args: args{
				ew:      nilEventGen,
				headers: http.Header{"h1": {"v1"}},
			},
			want: &HttpSSEResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 200,
					HttpHeaders:    http.Header{"Content-Type": {"text/event-stream"}, "Cache-Control": {"no-cache"}, "Connection": {"keep-alive"}, "h1": {"v1"}},
					HttpCookies:    NewHttpCookies(nil),
				},
				EventGenerator: nilEventGen,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SSEHttpResponseWithHeaders(tt.args.ew, tt.args.headers)
			if !reflect.DeepEqual(got.HttpHeadersResponse, tt.want.HttpHeadersResponse) {
				t.Fatalf("SSEHttpResponseWithHeaders() = %v, want %v", got, tt.want)
			}
			if got.EventGenerator == nil {
				t.Fatalf("SSEHttpResponseWithHeaders() EventGenerator is nil")
			}
		})
	}
}

func TestSSEHttpResponseWithHeadersAndCookies(t *testing.T) {
	type args struct {
		ew      EventGenerator
		headers http.Header
		cookies []http.Cookie
	}
	tests := []struct {
		name string
		args args
		want *HttpSSEResponse
	}{
		{
			name: "constructor",
			args: args{
				ew:      nilEventGen,
				headers: http.Header{"h1": {"v1"}},
				cookies: []http.Cookie{{
					Name:  "cookie3",
					Value: "val3",
				}},
			},
			want: &HttpSSEResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: 200,
					HttpHeaders:    http.Header{"Content-Type": {"text/event-stream"}, "Cache-Control": {"no-cache"}, "Connection": {"keep-alive"}, "h1": {"v1"}},
					HttpCookies: NewHttpCookies([]http.Cookie{{
						Name:  "cookie3",
						Value: "val3",
					}}),
				},
				EventGenerator: nilEventGen,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SSEHttpResponseWithHeadersAndCookies(tt.args.ew, tt.args.headers, tt.args.cookies)
			if !reflect.DeepEqual(got.HttpHeadersResponse, tt.want.HttpHeadersResponse) {
				t.Fatalf("SSEHttpResponseWithHeadersAndCookies() = %v, want %v", got, tt.want)
			}
			if got.EventGenerator == nil {
				t.Fatalf("SSEHttpResponseWithHeadersAndCookies() EventGenerator is nil")
			}
		})
	}
}

func TestHttpSSEResponse_Write(t *testing.T) {
	type args struct {
		request                                *http.Request
		httpStatusCode                         int
		httpHeaders                            http.Header
		httpCookies                            []http.Cookie
		eventGeneratorStartIndex               int
		eventGeneratorCallRequestContextCancel bool
	}
	type want struct {
		httpCode    int
		httpHeaders http.Header
		body        string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name:    "http1 not supported",
			args:    args{request: &http.Request{ProtoMajor: 1}},
			wantErr: true,
		},
		{
			name: "request cancels",
			args: args{
				request:                                &http.Request{ProtoMajor: 2},
				httpStatusCode:                         200,
				eventGeneratorStartIndex:               0,
				eventGeneratorCallRequestContextCancel: true,
			},
			want: want{
				httpCode:    200,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}},
				body:        "event: message\ndata: msg1\nid: 1\n\nevent: message\ndata: msg2\nid: 2\n\n",
			},
		},
		{
			name: "data streams ends",
			args: args{
				request:                                &http.Request{ProtoMajor: 2},
				httpStatusCode:                         200,
				eventGeneratorStartIndex:               0,
				eventGeneratorCallRequestContextCancel: false,
			},
			want: want{
				httpCode:    200,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}},
				body:        "event: message\ndata: msg1\nid: 1\n\nevent: message\ndata: msg2\nid: 2\n\nevent: message\ndata: msg3\nid: 3\n\n",
			},
		},
		{
			name: "start from index form header",
			args: args{
				request:                                &http.Request{ProtoMajor: 2, Header: http.Header{"Last-Event-Id": {"1"}}},
				httpStatusCode:                         200,
				eventGeneratorStartIndex:               0,
				eventGeneratorCallRequestContextCancel: false,
			},
			want: want{
				httpCode:    200,
				httpHeaders: http.Header{"X-Content-Type-Options": {"nosniff"}},
				body:        "event: message\ndata: msg2\nid: 2\n\nevent: message\ndata: msg3\nid: 3\n\n",
			},
		},
	}
	for _, tt := range tests {
		ctx, ctxCancel := context.WithCancel(context.Background())
		eventGenerator := testEventGen(0, func() {
			if tt.args.eventGeneratorCallRequestContextCancel {
				ctxCancel()
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			resp := &HttpSSEResponse{
				HttpHeadersResponse: HttpHeadersResponse{
					HttpStatusCode: tt.args.httpStatusCode,
				},
				EventGenerator: eventGenerator,
			}
			if tt.args.httpHeaders != nil {
				for k, v := range tt.args.httpHeaders {
					for _, e := range v {
						resp.Headers().Add(k, e)
					}
				}
			}
			for _, k := range tt.args.httpCookies {
				resp.Cookies().Add(k)
			}
			responseRecorder := httptest.NewRecorder()
			err := resp.Write(responseRecorder, request.HttpRequest{R: tt.args.request.WithContext(ctx)})
			if tt.wantErr {
				if err == nil {
					t.Errorf("HttpSSEResponse() want error but got nil")
				}
			} else {
				got := want{
					httpCode:    responseRecorder.Code,
					httpHeaders: responseRecorder.Header(),
					body:        responseRecorder.Body.String(),
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("HttpSSEResponse.Write() got:  %v, want: %v", got, tt.want)
				}
			}
		})
	}
}
