package router

import (
	"context"
	"fmt"
	"github.com/ixtendio/gow/internal/path"
	"github.com/ixtendio/gow/request"
	"github.com/ixtendio/gow/response"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func Test_endpointMatcher_addEndpoint(t *testing.T) {
	type args struct {
		method                   string
		pathPattern              string
		caseInsensitivePathMatch bool
		handler                  Handler
	}
	tests := []struct {
		name            string
		endpointMatcher *matcher
		args            *args
		expected        []string
		wantErr         bool
	}{
		{
			name:            "usecase_01",
			endpointMatcher: newMatcher(),
			args: &args{
				method:      "GET",
				pathPattern: "/",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/#]"},
			wantErr:  false,
		},
		{
			name:            "usecase_02",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/{b}",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b}#]"},
			wantErr:  false,
		},
		{
			name:            "usecase_03",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b}/c"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/{b}/{c}",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b}#/c#{c}#]"},
			wantErr:  false,
		},
		{
			name:            "usecase_04",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b}/c", "GET:/a/{b}/{c}"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/{b}/{d}",
				handler:     emptyHandler(),
			},
			wantErr: true,
		},
		{
			name:            "usecase_05",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b}/c", "GET:/a/{b}/{c}"),
			args: &args{
				method:      "GET",
				pathPattern: "/b/a/{b}/{c}",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b}#/c#{c}#b/a/{b}/{c}#]"},
			wantErr:  false,
		},
		{
			name:            "usecase_06",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d}"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/*",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b:\\d}#{b}#*#]"},
			wantErr:  false,
		},
		{
			name:            "usecase_07",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b"),
			args: &args{
				method:      "POST",
				pathPattern: "/a/b",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#]", "POST:[/a/b#]"},
			wantErr:  false,
		},
		{
			name:            "usecase_08",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "POST:/a/b"),
			args: &args{
				method:      "DELETE",
				pathPattern: "/a/b",
				handler:     emptyHandler(),
			},
			expected: []string{"DELETE:[/a/b#]", "GET:[/a/b#]", "POST:[/a/b#]"},
			wantErr:  false,
		},
		{
			name:            "usecase_09",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d}", "GET:/a/*"),
			args: &args{
				method:      "GET",
				pathPattern: "/a/**",
				handler:     emptyHandler(),
			},
			expected: []string{"GET:[/a/b#{b:\\d}#{b}#*#**#]"},
			wantErr:  false,
		},
		{
			name:            "usecase_10",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a"),
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
			if err := tt.endpointMatcher.addEndpoint(tt.args.method, tt.args.pathPattern, tt.args.caseInsensitivePathMatch, tt.args.handler); (err != nil) != tt.wantErr {
				t.Errorf("addEndpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got := trieToString(tt.endpointMatcher.trieRoots)
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
	tests := []struct {
		name    string
		matcher *matcher
		args    *args
		want    string
	}{
		{
			name:    "usecase_01",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/a/ab"),
			},
			want: "GET:/a/{b}",
		},
		{
			name:    "usecase_02",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d{4}}"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/a/123"),
			},
			want: "GET:/a/{b}",
		},
		{
			name:    "usecase_03",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d{4}}"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/a/12345"),
			},
			want: "GET:/a/{b:\\d{4}}",
		},
		{
			name:    "usecase_04",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}/c"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/a/b/c"),
			},
			want: "GET:/a/{b}/c",
		},
		{
			name:    "usecase_05",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/*/*/d", "GET:/a/**/e"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/a/b/c/d/e"),
			},
			want: "GET:/a/**/e",
		},
		{
			name:    "usecase_06",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/bla/**/**/bla"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/bla/bla/bla/bla/bla/bla"),
			},
			want: "GET:/bla/**/**/bla",
		},
		{
			name:    "usecase_07",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/bla/**/bla"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/bla/testing/testing/bla/bla"),
			},
			want: "GET:/bla/**/bla",
		},
		{
			name:    "usecase_08",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/bla/**/**/bla", "GET:/bla/**/bla"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/bla/testing/testing/bla/bla"),
			},
			want: "GET:/bla/**/**/bla",
		},
		{
			name:    "usecase_09",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/*bla*/**/bla/**"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/XXXblaXXXX/testing/testing/bla/testing/testing/"),
			},
			want: "GET:/*bla*/**/bla/**",
		},
		{
			name:    "usecase_10",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/*bla*/**/bla/**"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/XXXblaXXXX/testing/testing/bla/testing/testing.jpg"),
			},
			want: "GET:/*bla*/**/bla/**",
		},
		{
			name:    "usecase_11",
			matcher: newEndpointMatcherWithPatterns(t, "GET:/docs/**/**/**"),
			args: &args{
				method: "GET",
				mc:     newMatchingContext(t, "/docs/cvs/other/commit.html"),
			},
			want: "GET:/docs/**/**/**",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.matcher.match(tt.args.method, tt.args.mc)
			var gotAsString string
			if got != nil {
				_, err := got(nil, nil)
				gotAsString = err.Error()
			}
			if tt.want != gotAsString {
				t.Errorf("match() got = %v, want %v", gotAsString, tt.want)
			}
		})
	}
}

func newMatchingContext(t *testing.T, urlPath string) *path.MatchingContext {
	rawUrl := fmt.Sprintf("https://www.somesite.com%s", urlPath)
	u, err := url.Parse(rawUrl)
	if err != nil {
		t.Fatalf("failed to parse url: %s", rawUrl)
	}
	return path.ParseRequestURL(u)
}

func newEndpointMatcherWithPatterns(t *testing.T, patterns ...string) *matcher {
	endpointMatcher := newMatcher()
	for _, pattern := range patterns {
		parts := strings.SplitN(pattern, ":", 2)
		if err := endpointMatcher.addEndpoint(strings.ToUpper(parts[0]), parts[1], false, errorHandler(pattern)); err != nil {
			t.Fatalf("failed to register: %s, err: %v", pattern, err)
		}
	}
	return endpointMatcher
}

func errorHandler(msg string) Handler {
	return func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return nil, fmt.Errorf(msg)
	}
}

func emptyHandler() Handler {
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
