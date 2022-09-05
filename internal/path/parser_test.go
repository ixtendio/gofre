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
		_, err := Parse("/abc/def/{abc:def}/", false)
		if err != nil {
			b.Errorf("%v", err)
		}
	}
}

func BenchmarkTestParseRequestURL(b *testing.B) {
	requstUrl := mustParseURL("https://example.com/abc/def/ghe")
	for i := 0; i < b.N; i++ {
		_ = ParseRequestURL(requstUrl)
	}
}

func TestParse(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.pathPattern, tt.args.caseInsensitive)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotAsString := elementsToString(got)
			if gotAsString != tt.want {
				t.Errorf("Parse() got = %v, want %v", gotAsString, tt.want)
			}
		})
	}
}

func TestParseRequestURL(t *testing.T) {
	type args struct {
		requestUrl *url.URL
	}
	tests := []struct {
		name string
		args args
		want *MatchingContext
	}{
		{
			name: "/",
			args: args{requestUrl: mustParseURL("https://example.com/?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath:          "/",
				PathElements:          []string{"/"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/abc",
			args: args{requestUrl: mustParseURL("https://example.com/abc?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath:          "/abc",
				PathElements:          []string{"/", "abc"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "//",
			args: args{requestUrl: mustParseURL("https://example.com//?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath:          "//",
				PathElements:          []string{"/"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "//abc",
			args: args{requestUrl: mustParseURL("https://example.com//abc?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath:          "//abc",
				PathElements:          []string{"/", "abc"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/abc/",
			args: args{requestUrl: mustParseURL("https://example.com/abc/?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath:          "/abc/",
				PathElements:          []string{"/", "abc", "/"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/abc//def",
			args: args{requestUrl: mustParseURL("https://example.com/abc//def?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath:          "/abc//def",
				PathElements:          []string{"/", "abc", "/", "def"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/foo%25fbar",
			args: args{requestUrl: mustParseURL("https://example.com/foo%25fbar?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath:          "/foo%fbar",
				PathElements:          []string{"/", "foo%fbar"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/foo%2fbar",
			args: args{requestUrl: mustParseURL("https://example.com/foo%2fbar?q=morefoo%25bar")},
			want: &MatchingContext{
				OriginalPath:          "/foo/bar",
				PathElements:          []string{"/", "foo", "/", "bar"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/path/to/new/../file",
			args: args{requestUrl: mustParseURL("https://example.com/path/to/new/../file")},
			want: &MatchingContext{
				OriginalPath:          "/path/to/new/../file",
				PathElements:          []string{"/", "path", "/", "to", "/", "file"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/foo/../../bar",
			args: args{requestUrl: mustParseURL("https://example.com/foo/../../bar")},
			want: &MatchingContext{
				OriginalPath:          "/foo/../../bar",
				PathElements:          []string{"/", "bar"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/foo/../..",
			args: args{requestUrl: mustParseURL("https://example.com/foo/../..")},
			want: &MatchingContext{
				OriginalPath:          "/foo/../..",
				PathElements:          []string{"/"},
				ExtractedUriVariables: map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRequestURL(tt.args.requestUrl)
			if got.OriginalPath != tt.want.OriginalPath {
				t.Errorf("ParseRequestURL() OriginalPath = %v, want %v", got.OriginalPath, tt.want.OriginalPath)
			}
			if !reflect.DeepEqual(got.PathElements, tt.want.PathElements) {
				t.Errorf("ParseRequestURL() PathElements = %v, want %v", got.PathElements, tt.want.PathElements)
			}
			if !reflect.DeepEqual(got.ExtractedUriVariables, tt.want.ExtractedUriVariables) {
				t.Errorf("ParseRequestURL() ExtractedUriVariables = %v, want %v", got.ExtractedUriVariables, tt.want.ExtractedUriVariables)
			}
		})
	}
}

func elementsToString(elements []*Element) string {
	var sb strings.Builder
	for i := 0; i < len(elements); i++ {
		sb.WriteString("[")
		sb.WriteString(elements[i].RawVal)
		sb.WriteString("]")
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
