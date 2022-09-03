package router

import (
	"context"
	"github.com/ixtendio/gow/internal/path"
	"github.com/ixtendio/gow/request"
	"github.com/ixtendio/gow/response"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func Test_endpointMatcher_addEndpoint(t *testing.T) {
	tests := []struct {
		name            string
		endpointMatcher *endpointMatcher
		args            *endpoint
		expected        []string
		wantErr         bool
	}{
		{
			name:            "usecase_01",
			endpointMatcher: newEndpointMatcher(),
			args: &endpoint{
				method:   "GET",
				rootPath: parsePattern(t, "/"),
				handler:  emptyHandler(),
			},
			expected: []string{"GET:[(leaf)]"},
			wantErr:  false,
		},
		{
			name:            "usecase_02",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b"),
			args: &endpoint{
				method:   "GET",
				rootPath: parsePattern(t, "/a/{b}"),
				handler:  emptyHandler(),
			},
			expected: []string{"GET:[(a)(b)(leaf)({b})(leaf)]"},
			wantErr:  false,
		},
		{
			name:            "usecase_03",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b}/c"),
			args: &endpoint{
				method:   "GET",
				rootPath: parsePattern(t, "/a/{b}/{c}"),
				handler:  emptyHandler(),
			},
			expected: []string{"GET:[(a)(b)(leaf)({b})(leaf)(c)(leaf)({c})(leaf)]"},
			wantErr:  false,
		},
		{
			name:            "usecase_04",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b}/c", "GET:/a/{b}/{c}"),
			args: &endpoint{
				method:   "GET",
				rootPath: parsePattern(t, "/a/{b}/{d}"),
				handler:  emptyHandler(),
			},
			wantErr: true,
		},
		{
			name:            "usecase_05",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b}/c", "GET:/a/{b}/{c}"),
			args: &endpoint{
				method:   "GET",
				rootPath: parsePattern(t, "/b/a/{b}/{c}"),
				handler:  emptyHandler(),
			},
			expected: []string{"GET:[(a)(b)(leaf)({b})(leaf)(c)(leaf)({c})(leaf)][(b)(a)({b})({c})(leaf)]"},
			wantErr:  false,
		},
		{
			name:            "usecase_06",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d}"),
			args: &endpoint{
				method:   "GET",
				rootPath: parsePattern(t, "/a/*"),
				handler:  emptyHandler(),
			},
			expected: []string{"GET:[(a)(b)(leaf)({b:\\d})(leaf)({b})(leaf)(*)(leaf)]"},
			wantErr:  false,
		},
		{
			name:            "usecase_07",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b"),
			args: &endpoint{
				method:   "POST",
				rootPath: parsePattern(t, "/a/b"),
				handler:  emptyHandler(),
			},
			expected: []string{"GET:[(a)(b)(leaf)]", "POST:[(a)(b)(leaf)]"},
			wantErr:  false,
		},
		{
			name:            "usecase_08",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "POST:/a/b"),
			args: &endpoint{
				method:   "DELETE",
				rootPath: parsePattern(t, "/a/b"),
				handler:  emptyHandler(),
			},
			expected: []string{"DELETE:[(a)(b)(leaf)]", "GET:[(a)(b)(leaf)]", "POST:[(a)(b)(leaf)]"},
			wantErr:  false,
		},
		{
			name:            "usecase_09",
			endpointMatcher: newEndpointMatcherWithPatterns(t, "GET:/a/b", "GET:/a/{b}", "GET:/a/{b:\\d}", "GET:/a/*"),
			args: &endpoint{
				method:   "GET",
				rootPath: parsePattern(t, "/a/**"),
				handler:  emptyHandler(),
			},
			expected: []string{"GET:[(a)(b)(leaf)({b:\\d})(leaf)({b})(leaf)(*)(leaf)(**)(leaf)]"},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.endpointMatcher.addEndpoint(tt.args); (err != nil) != tt.wantErr {
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

func newEndpointMatcherWithPatterns(t *testing.T, patterns ...string) *endpointMatcher {
	endpointMatcher := newEndpointMatcher()
	for _, pattern := range patterns {
		parts := strings.SplitN(pattern, ":", 2)
		n := parsePattern(t, parts[1])
		if err := endpointMatcher.addEndpoint(&endpoint{
			method:   strings.ToUpper(parts[0]),
			rootPath: n,
			handler:  emptyHandler(),
		}); err != nil {
			t.Fatalf("failed to register: %s, err: %v", pattern, err)
		}
	}
	return endpointMatcher
}

func parsePattern(t *testing.T, pattern string) *path.Element {
	n, err := path.Parse(pattern, false)
	if err != nil {
		t.Fatalf("failed to parse pattern: %s, err: %v", pattern, err)
	}
	return n
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
		var preOrder func(root *trieNode)
		preOrder = func(root *trieNode) {
			if root == nil {
				return
			}
			if root.isLeaf() {
				sb.WriteString("(leaf)")
			} else {
				sb.WriteString("(")
				sb.WriteString(root.pathElement.RawVal)
				sb.WriteString(")")
			}

			for i := 0; i < len(root.children); i++ {
				preOrder(root.children[i])
			}
		}

		sb.WriteString(strings.ToUpper(method))
		sb.WriteString(":")
		for i := 0; i < len(root.children); i++ {
			sb.WriteString("[")
			preOrder(root.children[i])
			sb.WriteString("]")
		}
		result = append(result, sb.String())
	}
	sort.SliceStable(result, func(i, j int) bool {
		return strings.Compare(result[i], result[j]) < 0
	})
	return result
}
