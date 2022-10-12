package router

import (
	"context"
	"fmt"
	handler2 "github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/internal/path"
	"github.com/ixtendio/gofre/request"
	"github.com/ixtendio/gofre/response"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func Benchmark_matcher_match(b *testing.B) {
	matcher := newEndpointMatcherWithPatternsB(b, "GET:/a/{b}/{c}/{d}/f/g", "GET:/a/{c}/{b}/**/f/{d}", "GET:/a/{c}/{b}/**/f/g", "GET:/a/{c}/{b}/*/{d}/f/g")
	mc := newMatchingContextB(b, "/a/b/c/d/e/f/g")
	for i := 0; i < b.N; i++ {
		matcher.match("GET", mc)
	}
}

func Test_endpointMatcher_addEndpoint(t *testing.T) {
	type args struct {
		method                   string
		pathPattern              string
		caseInsensitivePathMatch bool
		handler                  handler2.Handler
	}
	tests := []struct {
		name                    string
		existingEndpointMatcher *matcher
		args                    *args
		expected                []string
		wantErr                 bool
	}{
		{
			name:                    "usecase_01",
			existingEndpointMatcher: newMatcher(),
			args: &args{
				method:      "GET",
				pathPattern: "/",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/#]"},
			wantErr:  false,
		},
		{
			name:                    "usecase_02",
			existingEndpointMatcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/{b}",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b}#]"},
			wantErr:  false,
		},
		{
			name:                    "usecase_03",
			existingEndpointMatcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b}/c"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/{b}/{c}",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b}#/c#{c}#]"},
			wantErr:  false,
		},
		{
			name:                    "usecase_04",
			existingEndpointMatcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b}/c", "GET:/a/{b}/{c}"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/{b}/{d}",
				handler:     emptyHandler(),
			},
			wantErr: true,
		},
		{
			name:                    "usecase_05",
			existingEndpointMatcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b}/c", "GET:/a/{b}/{c}"),
			args: &args{
				method:      "GET",
				pathPattern: "/b/a/{b}/{c}",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b}#/c#{c}#b/a/{b}/{c}#]"},
			wantErr:  false,
		},
		{
			name:                    "usecase_06",
			existingEndpointMatcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d}"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/*",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b:\\d}#{b}#*#]"},
			wantErr:  false,
		},
		{
			name:                    "usecase_07",
			existingEndpointMatcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b"),
			args: &args{
				method:      "POST",
				pathPattern: "/a/b",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#]", "POST:[/a/b#]"},
			wantErr:  false,
		},
		{
			name:                    "usecase_08",
			existingEndpointMatcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "POST:/a/b"),
			args: &args{
				method:      "DELETE",
				pathPattern: "/a/b",
				handler:     emptyHandler(),
			},
			expected: []string{"DELETE:[/a/b#]", "GET:[/a/b#]", "POST:[/a/b#]"},
			wantErr:  false,
		},
		{
			name:                    "usecase_09",
			existingEndpointMatcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d}", "GET:/a/*"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/**",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b:\\d}#{b}#*#**#]"},
			wantErr:  false,
		},
		{
			name:                    "usecase_10",
			existingEndpointMatcher: newEndpointMatcherWithPatternsT(t, "GET:/a"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a#/#]"},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.existingEndpointMatcher.addEndpoint(tt.args.method, tt.args.pathPattern, tt.args.caseInsensitivePathMatch, tt.args.handler); (err != nil) != tt.wantErr {
				t.Errorf("addEndpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got := trieToString(tt.existingEndpointMatcher.trieRoots)
				if !reflect.DeepEqual(tt.expected, got) {
					t.Errorf("addEndpoint() got = '%v', want = '%v'", got, tt.expected)
				}
			}
		})
	}
}

func Test_matcher_match(t *testing.T) {
	type args struct {
		method string
		mc     *path.MatchingContext
	}
	type want struct {
		endpoint     string
		capturedVars map[string]string
	}
	tests := []struct {
		name    string
		matcher *matcher
		args    *args
		want    want
	}{
		{
			name:    "usecase_01",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/{b}"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/a/ab"),
			},
			want: want{
				endpoint:     "GET:/a/{b}",
				capturedVars: map[string]string{"b": "ab"},
			},
		},
		{
			name:    "usecase_02",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d{4}}"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/a/123"),
			},
			want: want{
				endpoint:     "GET:/a/{b}",
				capturedVars: map[string]string{"b": "123"},
			},
		},
		{
			name:    "usecase_03",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d{4}}"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/a/12345"),
			},
			want: want{
				endpoint:     "GET:/a/{b:\\d{4}}",
				capturedVars: map[string]string{"b": "12345"},
			},
		},
		{
			name:    "usecase_04",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/{b}/c"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/a/b/c"),
			},
			want: want{
				endpoint:     "GET:/a/{b}/c",
				capturedVars: map[string]string{"b": "b"},
			},
		},
		{
			name:    "usecase_05",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/a/b", "GET:/a/*/*/d", "GET:/a/**/e"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/a/b/c/d/e"),
			},
			want: want{
				endpoint:     "GET:/a/**/e",
				capturedVars: map[string]string{},
			},
		},
		{
			name:    "usecase_06",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/bla/**/**/bla"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/bla/bla/bla/bla/bla/bla"),
			},
			want: want{
				endpoint:     "GET:/bla/**/**/bla",
				capturedVars: map[string]string{},
			},
		},
		{
			name:    "usecase_07",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/bla/**/bla"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/bla/testing/testing/bla/bla"),
			},
			want: want{
				endpoint:     "GET:/bla/**/bla",
				capturedVars: map[string]string{},
			},
		},
		{
			name:    "usecase_08",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/bla/**/**/bla", "GET:/bla/**/bla"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/bla/testing/testing/bla/bla"),
			},
			want: want{
				endpoint:     "GET:/bla/**/**/bla",
				capturedVars: map[string]string{},
			},
		},
		{
			name:    "usecase_09",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/*bla*/**/bla/**"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/XXXblaXXXX/testing/testing/bla/testing/testing/"),
			},
			want: want{
				endpoint:     "GET:/*bla*/**/bla/**",
				capturedVars: map[string]string{},
			},
		},
		{
			name:    "usecase_10",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/*bla*/**/bla/**"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/XXXblaXXXX/testing/testing/bla/testing/testing.jpg"),
			},
			want: want{
				endpoint:     "GET:/*bla*/**/bla/**",
				capturedVars: map[string]string{},
			},
		},
		{
			name:    "usecase_11",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/docs/**/**/**"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/docs/cvs/other/commit.html"),
			},
			want: want{
				endpoint:     "GET:/docs/**/**/**",
				capturedVars: map[string]string{},
			},
		},
		{
			name:    "usecase_12",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/a/{b}/{c}/**/g/h"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/a/b/c/d/e/f/g/h"),
			},
			want: want{
				endpoint:     "GET:/a/{b}/{c}/**/g/h",
				capturedVars: map[string]string{"b": "b", "c": "c"},
			},
		},
		{
			name:    "usecase_13",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/a/{b}/{c}/*/f/g", "GET:/a/{c}/{b}/**/f/g"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/a/b/c/d/e/f/g"),
			},
			want: want{
				endpoint:     "GET:/a/{c}/{b}/**/f/g",
				capturedVars: map[string]string{"c": "b", "b": "c"},
			},
		},
		{
			name:    "usecase_14",
			matcher: newEndpointMatcherWithPatternsT(t, "GET:/a/{b}/{c}/{d}/f/g", "GET:/a/{c}/{b}/**/f/{d}", "GET:/a/{c}/{b}/**/f/g", "GET:/a/{c}/{b}/*/{d}/f/g"),
			args: &args{
				method: "GET",
				mc:     newMatchingContextT(t, "/a/b/c/d/e/f/g"),
			},
			want: want{
				endpoint:     "GET:/a/{c}/{b}/*/{d}/f/g",
				capturedVars: map[string]string{"c": "b", "b": "c", "d": "e"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, capturedVars := tt.matcher.match(tt.args.method, tt.args.mc)
			var gotAsString string
			if got != nil {
				_, err := got(nil, nil)
				gotAsString = err.Error()
			}
			if tt.want.endpoint != gotAsString {
				t.Errorf("match() got = %v, want %v", gotAsString, tt.want.endpoint)
			}
			if !reflect.DeepEqual(tt.want.capturedVars, capturedVars) {
				t.Errorf("match() got = %v, want %v", capturedVars, tt.want.capturedVars)
			}
		})
	}
}

func newMatchingContextB(t *testing.B, urlPath string) *path.MatchingContext {
	mc, err := newMatchingContext(urlPath)
	if err != nil {
		t.Fatal(err)
	}
	return mc
}

func newMatchingContextT(t *testing.T, urlPath string) *path.MatchingContext {
	mc, err := newMatchingContext(urlPath)
	if err != nil {
		t.Fatal(err)
	}
	return mc
}

func newMatchingContext(urlPath string) (*path.MatchingContext, error) {
	rawUrl := fmt.Sprintf("https://www.somesite.com%s", urlPath)
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %s", rawUrl)
	}
	return path.ParseURL(u), nil
}

func newEndpointMatcherWithPatternsB(t *testing.B, patterns ...string) *matcher {
	m, err := newEndpointMatcherWithPatterns(patterns...)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func newEndpointMatcherWithPatternsT(t *testing.T, patterns ...string) *matcher {
	m, err := newEndpointMatcherWithPatterns(patterns...)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func newEndpointMatcherWithPatterns(patterns ...string) (*matcher, error) {
	endpointMatcher := newMatcher()
	for _, pattern := range patterns {
		parts := strings.SplitN(pattern, ":", 2)
		if err := endpointMatcher.addEndpoint(strings.ToUpper(parts[0]), parts[1], false, errorHandler(pattern)); err != nil {
			return nil, fmt.Errorf("failed to register: %s, err: %v", pattern, err)
		}
	}
	return endpointMatcher, nil
}

func errorHandler(msg string) handler2.Handler {
	return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return nil, fmt.Errorf(msg)
	}
}

func emptyHandler() handler2.Handler {
	return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return nil, nil
	}
}

func trieToString(roots map[string]*trieNode) []string {
	var result []string
	for method, root := range roots {
		var sb strings.Builder
		var preOrder func(*trieNode)
		preOrder = func(root *trieNode) {
			if root == nil {
				return
			}
			if root.isLeaf() {
				sb.WriteString("#")
			} else {
				sb.WriteString(root.pathElement.RawVal)
			}

			for i := 0; i < len(root.children); i++ {
				preOrder(root.children[i])
			}
		}

		sb.WriteString(strings.ToUpper(method))
		sb.WriteString(":[")
		preOrder(root)
		sb.WriteString("]")
		result = append(result, sb.String())
	}
	sort.SliceStable(result, func(i, j int) bool {
		return strings.Compare(result[i], result[j]) < 0
	})
	return result
}
