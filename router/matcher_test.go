package router

import (
	"context"
	"github.com/ixtendio/gofre/internal/path"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"reflect"
	"strings"
	"testing"
)

func Test_matcheri_addEndpoint(t *testing.T) {
	type want struct {
		patternsMap map[string][]string
	}
	tests := []struct {
		name string
		args []string
		want want
	}{
		{
			name: "GET with multiple paths",
			args: []string{
				"GET:/",
				"GET:/a/{b}",
				"GET:/a/b/c",
				"GET:/c/b/{c:[a-z]+}",
				"GET:/b/a/{c}",
				"GET:/s/b/**",
				"GET:/t/b/**/d",
				"GET:/a/v/**/{d}",
				"GET:/a/o/*/{d}",
				"GET:/a/*/",
				"GET:/a/**",
				"GET:/a/**/b/**/c",
				"GET:/**/a",
			},
			want: want{
				map[string][]string{"GET": {
					"/",
					"/a/b/c",
					"/c/b/{c:[a-z]+}",
					"/b/a/{c}",
					"/a/o/*/{d}",
					"/t/b/**/d",
					"/a/v/**/{d}",
					"/s/b/**",
					"/a/{b}",
					"/a/*/",
					"/a/**/b/**/c",
					"/a/**",
					"/**/a",
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &matcher{
				patternsMap: make(map[string][]path.Pattern),
			}
			for _, arg := range tt.args {
				s := strings.SplitN(arg, ":", 2)
				err := m.addEndpoint(s[0], s[1], false,
					func(ctx context.Context, r request.HttpRequest) (response.HttpResponse, error) {
						return response.PlainTextHttpResponseOK("ok"), nil
					})
				if err != nil {
					t.Fatalf("addEndpoint() got err: %v , want nil", err)
				}
			}
			got := want{patternsMap: make(map[string][]string)}
			for k, v := range m.patternsMap {
				for _, e := range v {
					got.patternsMap[k] = append(got.patternsMap[k], e.RawValue)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("addEndpoint() got: %v , want: %v", got, tt.want)
			}
		})
	}
}
