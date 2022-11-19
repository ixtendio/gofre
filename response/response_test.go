package response

import (
	"net/http"
	"reflect"
	"testing"
)

func TestHttpCookies_Add(t *testing.T) {
	cookie1 := &http.Cookie{
		Name:  "cookie1",
		Value: "val1",
	}
	cookie2 := &http.Cookie{
		Name:   "cookie2",
		Value:  "val2",
		Path:   "path2",
		Domain: "domain2",
	}
	type args struct {
		cookies HttpCookies
	}
	tests := []struct {
		name string
		c    HttpCookies
		args args
		want HttpCookies
	}{
		{
			name: "add one cookie",
			c:    HttpCookies{},
			args: args{
				cookies: NewHttpCookie(cookie1),
			},
			want: HttpCookies{"cookie1::": cookie1},
		}, {
			name: "add the same cookie twice",
			c:    HttpCookies{},
			args: args{
				cookies: NewHttpCookie(cookie1, cookie1),
			},
			want: HttpCookies{"cookie1::": cookie1},
		}, {
			name: "add two cookies",
			c:    HttpCookies{},
			args: args{
				cookies: NewHttpCookie(cookie1, cookie2),
			},
			want: HttpCookies{
				"cookie1::":             cookie1,
				"cookie2:path2:domain2": cookie2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, cookie := range tt.args.cookies {
				tt.c.Add(cookie)
			}
			if !reflect.DeepEqual(tt.c, tt.want) {
				t.Errorf("NewEmptyHttpCookie.Add() got: %v, want: %v", tt.c, tt.want)
			}
		})
	}
}

func TestNewHttpCookies(t *testing.T) {
	tests := []struct {
		name string
		want HttpCookies
	}{
		{
			name: "construct with nil arg",
			want: HttpCookies{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEmptyHttpCookie(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEmptyHttpCookie() = %v, want %v", got, tt.want)
			}
		})
	}
}
