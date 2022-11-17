package request

import (
	"github.com/ixtendio/gofre/internal/path"
	"net/http"
	"reflect"
	"testing"
)

func TestNewHttpRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://www.domain.com", nil)
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want HttpRequest
	}{
		{
			name: "construct",
			args: args{r: req},
			want: HttpRequest{
				R:        req,
				pathVars: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHttpRequest(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHttpRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHttpRequestWithPathVars(t *testing.T) {
	m := []path.CaptureVar{{Name: "key", Value: "val"}}
	req, _ := http.NewRequest("GET", "https://www.domain.com", nil)
	type args struct {
		r        *http.Request
		pathVars []path.CaptureVar
	}
	tests := []struct {
		name string
		args args
		want HttpRequest
	}{
		{
			name: "construct",
			args: args{r: req, pathVars: m},
			want: HttpRequest{
				R:        req,
				pathVars: m,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHttpRequestWithPathVars(tt.args.r, tt.args.pathVars); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHttpRequestWithPathVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpRequest_PathVar(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://www.domain.com", nil)
	type fields struct {
		R        *http.Request
		pathVars []path.CaptureVar
	}
	type args struct {
		varName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "pathVars is nil",
			fields: fields{
				R:        req,
				pathVars: nil,
			},
			args: args{"key"},
			want: "",
		},
		{
			name: "pathVars is not nil but key does not exists",
			fields: fields{
				R:        req,
				pathVars: []path.CaptureVar{{Name: "key1", Value: "val"}},
			},
			args: args{"key"},
			want: "",
		}, {
			name: "pathVars is not nil and key exists",
			fields: fields{
				R:        req,
				pathVars: []path.CaptureVar{{Name: "key", Value: "val"}},
			},
			args: args{"key"},
			want: "val",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HttpRequest{
				R:        tt.fields.R,
				pathVars: tt.fields.pathVars,
			}
			if got := r.PathVar(tt.args.varName); got != tt.want {
				t.Errorf("PathVar() = %v, want %v", got, tt.want)
			}
		})
	}
}
