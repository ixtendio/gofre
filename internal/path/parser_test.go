package path

import (
	"log"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func BenchmarkTestParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParsePattern("/abc/def/{abc:def}/", false)
		if err != nil {
			b.Errorf("%v", err)
		}
	}
}

func BenchmarkTestParseRequestURL(b *testing.B) {
	requstUrl := mustParseURL("https://example.com/abc/def/ghe/as/df/aa/bb/../../../../cc/../../dd/ee/aa")
	for i := 0; i < b.N; i++ {
		_ = ParseURL(requstUrl)
	}
}

func Test_ParsePattern(t *testing.T) {
	type args struct {
		pathPattern     string
		caseInsensitive bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "path that not starts with slash",
			args:    args{pathPattern: "abc/{id:\\d}"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "path with capture variable without regex",
			args:    args{pathPattern: "/abc/{id}"},
			want:    "[/][abc][/][{id}]",
			wantErr: false,
		},
		{
			name:    "path with capture variable and regex",
			args:    args{pathPattern: "/abc/{id:\\d}"},
			want:    "[/][abc][/][{id:\\d}]",
			wantErr: false,
		},
		{
			name:    "root path",
			args:    args{pathPattern: "/"},
			want:    "[/]",
			wantErr: false,
		},
		{
			name:    "root path with double slash",
			args:    args{pathPattern: "//"},
			want:    "[/]",
			wantErr: false,
		},
		{
			name:    "path with many slash",
			args:    args{pathPattern: "/abc///cde////"},
			want:    "[/][abc][/][cde][/]",
			wantErr: false,
		},
		{
			name:    "path with %2f",
			args:    args{pathPattern: "/a%2fb"},
			want:    "[/][a][/][b]",
			wantErr: false,
		},
		{
			name:    "path with %2F",
			args:    args{pathPattern: "/a%2Fb"},
			want:    "[/][a][/][b]",
			wantErr: false,
		},
		{
			name:    "wrong path pattern 1",
			args:    args{pathPattern: "/a/{}"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "wrong path pattern 2",
			args:    args{pathPattern: "/a/{}/"},
			want:    "",
			wantErr: true,
		}, {
			name:    "wrong path pattern 3",
			args:    args{pathPattern: "/a/{}%2F"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePattern(tt.args.pathPattern, tt.args.caseInsensitive)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotAsString := elementsToString(got)
			if gotAsString != tt.want {
				t.Errorf("ParsePattern() got = %v, matchTypeWant %v", gotAsString, tt.want)
			}
		})
	}
}

func Test_ParseURL(t *testing.T) {
	type args struct {
		requestUrl *url.URL
	}
	tests := []struct {
		name string
		args args
		want *MatchingContext
	}{
		{
			name: "parse empty",
			args: args{requestUrl: mustParseURL("https://example.com")},
			want: &MatchingContext{
				OriginalPath: "",
				PathElements: nil,
			},
		},
		{
			name: "/",
			args: args{requestUrl: mustParseURL("https://example.com/?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath: "/",
				PathElements: []string{"/"},
			},
		},
		{
			name: "/abc",
			args: args{requestUrl: mustParseURL("https://example.com/abc?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath: "/abc",
				PathElements: []string{"/", "abc"},
			},
		},
		{
			name: "//",
			args: args{requestUrl: mustParseURL("https://example.com//?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath: "//",
				PathElements: []string{"/"},
			},
		},
		{
			name: "//abc",
			args: args{requestUrl: mustParseURL("https://example.com//abc?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath: "//abc",
				PathElements: []string{"/", "abc"},
			},
		},
		{
			name: "/abc/",
			args: args{requestUrl: mustParseURL("https://example.com/abc/?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath: "/abc/",
				PathElements: []string{"/", "abc", "/"},
			},
		},
		{
			name: "/abc//def",
			args: args{requestUrl: mustParseURL("https://example.com/abc//def?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath: "/abc//def",
				PathElements: []string{"/", "abc", "/", "def"},
			},
		},
		{
			name: "/foo%25fbar",
			args: args{requestUrl: mustParseURL("https://example.com/foo%25fbar?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath: "/foo%fbar",
				PathElements: []string{"/", "foo%fbar"},
			},
		},
		{
			name: "/foo%2fbar",
			args: args{requestUrl: mustParseURL("https://example.com/foo%2fbar?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath: "/foo/bar",
				PathElements: []string{"/", "foo", "/", "bar"},
			},
		},
		{
			name: "/path/to/new/../file",
			args: args{requestUrl: mustParseURL("https://example.com/path/to/new/../file")},
			want: &MatchingContext{
				OriginalPath: "/path/to/new/../file",
				PathElements: []string{"/", "path", "/", "to", "/", "file"},
			},
		},
		{
			name: "/foo/../../bar",
			args: args{requestUrl: mustParseURL("https://example.com/foo/../../bar")},
			want: &MatchingContext{
				OriginalPath: "/foo/../../bar",
				PathElements: []string{"/", "bar"},
			},
		},
		{
			name: "/foo/../..",
			args: args{requestUrl: mustParseURL("https://example.com/foo/../..")},
			want: &MatchingContext{
				OriginalPath: "/foo/../..",
				PathElements: []string{"/"},
			},
		},
		{
			name: "/foo/.%2e",
			args: args{requestUrl: mustParseURL("https://example.com/foo/.%2e")},
			want: &MatchingContext{
				OriginalPath: "/foo/..",
				PathElements: []string{"/"},
			},
		},
		{
			name: "/foo/%2e%2e",
			args: args{requestUrl: mustParseURL("https://example.com/foo/%2e%2e")},
			want: &MatchingContext{
				OriginalPath: "/foo/..",
				PathElements: []string{"/"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseURL(tt.args.requestUrl)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseURL() got = %#v, want = %#v", got, tt.want)
			}
		})
	}
}

func elementsToString(root *Element) string {
	var sb strings.Builder
	for root != nil {
		sb.WriteString("[")
		sb.WriteString(root.RawVal)
		sb.WriteString("]")
		root = root.Next
	}
	return sb.String()
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Failed parsing the url: %s, err:%v", rawURL, err)
	}
	return u
}

func Test_addPathSegment(t *testing.T) {
	type args struct {
		elements []string
		segment  string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty segment",
			args: args{
				elements: []string{"/", "a", "/", "b"},
				segment:  "",
			},
			want: []string{"/", "a", "/", "b"},
		},
		{
			name: "valid segment",
			args: args{
				elements: []string{"/", "a", "/"},
				segment:  "b",
			},
			want: []string{"/", "a", "/", "b"},
		},
		{
			name: ".. when array has no elements",
			args: args{
				elements: []string{""},
				segment:  "..",
			},
			want: []string{""},
		},
		{
			name: ".. when array has 1 element",
			args: args{
				elements: []string{"/"},
				segment:  "..",
			},
			want: []string{"/"},
		},
		{
			name: ".. when array has 2 elements",
			args: args{
				elements: []string{"/", "a"},
				segment:  "..",
			},
			want: []string{"/"},
		},
		{
			name: ".. when array has 3 elements",
			args: args{
				elements: []string{"/", "a", "/"},
				segment:  "..",
			},
			want: []string{"/"},
		},
		{
			name: ".. when array has 4 elements",
			args: args{
				elements: []string{"/", "a", "/", "b"},
				segment:  "..",
			},
			want: []string{"/", "a", "/"},
		},
		{
			name: ".. when array has 5 elements",
			args: args{
				elements: []string{"/", "a", "/", "b", "/"},
				segment:  "..",
			},
			want: []string{"/", "a", "/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addPathSegment(tt.args.elements, tt.args.segment); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addPathSegment() = %v, matchTypeWant %v", got, tt.want)
			}
		})
	}
}

func Test_parseElement(t *testing.T) {
	type args struct {
		pathPatternId   string
		element         string
		caseInsensitive bool
	}
	tests := []struct {
		name          string
		args          args
		matchTypeWant int
		wantErr       bool
	}{
		{
			name: "element is empty",
			args: args{
				pathPatternId:   "{a}",
				element:         "",
				caseInsensitive: false,
			},
			matchTypeWant: 0,
			wantErr:       false,
		},
		{
			name: "element is {}",
			args: args{
				pathPatternId:   "123",
				element:         "{}",
				caseInsensitive: false,
			},
			matchTypeWant: 0,
			wantErr:       true,
		},
		{
			name: "element is {a}",
			args: args{
				pathPatternId:   "123",
				element:         "{a}",
				caseInsensitive: false,
			},
			matchTypeWant: MatchVarCaptureType,
			wantErr:       false,
		},
		{
			name: "element is a",
			args: args{
				pathPatternId:   "123",
				element:         "a",
				caseInsensitive: false,
			},
			matchTypeWant: MatchLiteralType,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseElement(tt.args.pathPatternId, tt.args.element, tt.args.caseInsensitive)
			if tt.wantErr && err == nil {
				t.Fatalf("parseElement() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.matchTypeWant > 0 && got.MatchType != tt.matchTypeWant {
				t.Fatalf("parseElement() got = %#v, matchTypeWant = %#v", got, tt.matchTypeWant)
			}
		})
	}
}
