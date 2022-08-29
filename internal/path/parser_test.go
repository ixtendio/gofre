package path

import (
	"log"
	"net/url"
	"reflect"
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
			if got.String() != tt.want {
				t.Errorf("Parse() got = %v, want %v", got.String(), tt.want)
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
				originalPath:          "/",
				elements:              []string{"/"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/abc",
			args: args{requestUrl: mustParseURL("https://example.com/abc?q=morefoo%25bar")},
			want: &MatchingContext{
				originalPath:          "/abc",
				elements:              []string{"/", "abc"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "//",
			args: args{requestUrl: mustParseURL("https://example.com//?q=morefoo%25bar")},
			want: &MatchingContext{
				originalPath:          "//",
				elements:              []string{"/"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "//abc",
			args: args{requestUrl: mustParseURL("https://example.com//abc?q=morefoo%25bar")},
			want: &MatchingContext{
				originalPath:          "//abc",
				elements:              []string{"/", "abc"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/abc/",
			args: args{requestUrl: mustParseURL("https://example.com/abc/?q=morefoo%25bar")},
			want: &MatchingContext{
				originalPath:          "/abc/",
				elements:              []string{"/", "abc", "/"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/abc//def",
			args: args{requestUrl: mustParseURL("https://example.com/abc//def?q=morefoo%25bar")},
			want: &MatchingContext{
				originalPath:          "/abc//def",
				elements:              []string{"/", "abc", "/", "def"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/foo%25fbar",
			args: args{requestUrl: mustParseURL("https://example.com/foo%25fbar?q=morefoo%25bar")},
			want: &MatchingContext{
				originalPath:          "/foo%fbar",
				elements:              []string{"/", "foo%fbar"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/foo%2fbar",
			args: args{requestUrl: mustParseURL("https://example.com/foo%2fbar?q=morefoo%25bar")},
			want: &MatchingContext{
				originalPath:          "/foo/bar",
				elements:              []string{"/", "foo", "/", "bar"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/path/to/new/../file",
			args: args{requestUrl: mustParseURL("https://example.com/path/to/new/../file")},
			want: &MatchingContext{
				originalPath:          "/path/to/new/../file",
				elements:              []string{"/", "path", "/", "to", "/", "file"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/foo/../../bar",
			args: args{requestUrl: mustParseURL("https://example.com/foo/../../bar")},
			want: &MatchingContext{
				originalPath:          "/foo/../../bar",
				elements:              []string{"/", "bar"},
				ExtractedUriVariables: map[string]string{},
			},
		},
		{
			name: "/foo/../..",
			args: args{requestUrl: mustParseURL("https://example.com/foo/../..")},
			want: &MatchingContext{
				originalPath:          "/foo/../..",
				elements:              []string{"/"},
				ExtractedUriVariables: map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRequestURL(tt.args.requestUrl)
			if got.originalPath != tt.want.originalPath {
				t.Errorf("ParseRequestURL() originalPath = %v, want %v", got.originalPath, tt.want.originalPath)
			}
			if !reflect.DeepEqual(got.elements, tt.want.elements) {
				t.Errorf("ParseRequestURL() elements = %v, want %v", got.elements, tt.want.elements)
			}
			if !reflect.DeepEqual(got.ExtractedUriVariables, tt.want.ExtractedUriVariables) {
				t.Errorf("ParseRequestURL() ExtractedUriVariables = %v, want %v", got.ExtractedUriVariables, tt.want.ExtractedUriVariables)
			}
		})
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Failed parsing the url: %s, err:%v", rawURL, err)
	}
	return u
}
