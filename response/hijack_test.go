package response

import (
	"bufio"
	"github.com/ixtendio/gofre/router/path"

	"net"
	"net/http"
	"reflect"
	"testing"
)

var tcpCon = &net.TCPConn{}
var rw = &bufio.ReadWriter{}

type fakeHijackResponseWriter struct {
}

func (fakeHijackResponseWriter) Header() http.Header                          { return nil }
func (fakeHijackResponseWriter) Write([]byte) (int, error)                    { return 0, nil }
func (fakeHijackResponseWriter) WriteHeader(statusCode int)                   {}
func (fakeHijackResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) { return tcpCon, rw, nil }

type fakeNonHijackResponseWriter struct {
}

func (fakeNonHijackResponseWriter) Header() http.Header        { return nil }
func (fakeNonHijackResponseWriter) Write([]byte) (int, error)  { return 0, nil }
func (fakeNonHijackResponseWriter) WriteHeader(statusCode int) {}

func TestHttpHijackConnectionResponse_Write(t *testing.T) {
	var gotCon net.Conn
	var gotRw *bufio.ReadWriter
	var gotErr error

	type want struct {
		con net.Conn
		rw  *bufio.ReadWriter
	}
	type args struct {
		w              http.ResponseWriter
		hjCallbackFunc func(net.Conn, *bufio.ReadWriter, error)
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "HijackResponseWriter",
			args: args{
				w: fakeHijackResponseWriter{},
				hjCallbackFunc: func(conn net.Conn, writer *bufio.ReadWriter, err error) {
					gotCon = conn
					gotRw = writer
					gotErr = err
				},
			},
			want: want{
				con: tcpCon,
				rw:  rw,
			},
			wantErr: false,
		},

		{
			name: "NonHijackResponseWriter",
			args: args{
				w: fakeNonHijackResponseWriter{},
				hjCallbackFunc: func(conn net.Conn, writer *bufio.ReadWriter, err error) {
					gotCon = conn
					gotRw = writer
					gotErr = err
				},
			},
			want: want{
				con: nil,
				rw:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewHttpHijackConnectionResponse(tt.args.hjCallbackFunc)
			if err := r.Write(tt.args.w, path.MatchingContext{}); err != nil {
				t.Fatalf("Write() returned error: %v", err)
			}
			if r.StatusCode() != 0 {
				t.Errorf("StatusCode() expected value 0 got: %v", r.StatusCode())
			}
			if r.Headers() != nil {
				t.Errorf("Headers() expected value nil got: %v", r.Headers())
			}
			if r.Cookies() != nil {
				t.Errorf("Cookies() expected value nil got: %v", r.Cookies())
			}
			if tt.wantErr {
				if gotErr == nil {
					t.Errorf("Write() want error but got null")
				}
			} else {
				if gotErr != nil {
					t.Errorf("Write() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				got := want{
					con: gotCon,
					rw:  gotRw,
				}
				if !reflect.DeepEqual(tt.want, got) {
					t.Errorf("Write() got: %v, want: %v", got, tt.want)
				}
			}

		})
	}
}
