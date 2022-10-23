package response

import (
	"net/http"
	"reflect"
	"testing"
)

func TestHttpCookies_Add(t *testing.T) {
	type args struct {
		cookies []http.Cookie
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
				cookies: []http.Cookie{{
					Name:  "cookie1",
					Value: "val1",
				}},
			},
			want: HttpCookies{"cookie1::": http.Cookie{
				Name:  "cookie1",
				Value: "val1",
			}},
		}, {
			name: "add the same cookie twice",
			c:    HttpCookies{},
			args: args{
				cookies: []http.Cookie{{
					Name:  "cookie1",
					Value: "val1",
				}, {
					Name:  "cookie1",
					Value: "val1",
				}},
			},
			want: HttpCookies{"cookie1::": http.Cookie{
				Name:  "cookie1",
				Value: "val1",
			}},
		}, {
			name: "add two cookies",
			c:    HttpCookies{},
			args: args{
				cookies: []http.Cookie{{
					Name:  "cookie1",
					Value: "val1",
				}, {
					Name:   "cookie2",
					Value:  "val2",
					Path:   "path2",
					Domain: "domain2",
				}},
			},
			want: HttpCookies{
				"cookie1::": http.Cookie{
					Name:  "cookie1",
					Value: "val1",
				},
				"cookie2:path2:domain2": {
					Name:   "cookie2",
					Value:  "val2",
					Path:   "path2",
					Domain: "domain2",
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, cookie := range tt.args.cookies {
				tt.c.Add(cookie)
			}
			if !reflect.DeepEqual(tt.c, tt.want) {
				t.Errorf("NewHttpCookies.Add() got: %v, want: %v", tt.c, tt.want)
			}
		})
	}
}

func TestNewHttpCookies(t *testing.T) {
	type args struct {
		cookiesArray []http.Cookie
	}
	tests := []struct {
		name string
		args args
		want HttpCookies
	}{
		{
			name: "construct with nil arg",
			args: args{},
			want: HttpCookies{},
		},
		{
			name: "construct with empty array arg",
			args: args{
				cookiesArray: []http.Cookie{},
			},
			want: HttpCookies{},
		},
		{
			name: "construct with array",
			args: args{
				cookiesArray: []http.Cookie{{
					Name:   "cookie1",
					Value:  "val1",
					Path:   "path1",
					Domain: "domain1",
				}, {
					Name:   "cookie2",
					Value:  "val2",
					Path:   "path2",
					Domain: "domain2",
				}},
			},
			want: HttpCookies{
				"cookie1:path1:domain1": http.Cookie{
					Name:   "cookie1",
					Value:  "val1",
					Path:   "path1",
					Domain: "domain1",
				},
				"cookie2:path2:domain2": {
					Name:   "cookie2",
					Value:  "val2",
					Path:   "path2",
					Domain: "domain2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHttpCookies(tt.args.cookiesArray); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHttpCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}
