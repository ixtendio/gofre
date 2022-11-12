package path

import (
	"log"
	"net/url"
	"reflect"
	"sort"
	"testing"
)

var (
	staticPaths = []string{"/",
		"/cmd.html",
		"/code.html",
		"/contrib.html",
		"/contribute.html",
		"/debugging_with_gdb.html",
		"/docs.html",
		"/effective_go.html",
		"/files.log",
		"/gccgo_contribute.html",
		"/gccgo_install.html",
		"/go-logo-black.png",
		"/go-logo-blue.png",
		"/go-logo-white.png",
		"/go1.1.html",
		"/go1.2.html",
		"/go1.html",
		"/go1compat.html",
		"/go_faq.html",
		"/go_mem.html",
		"/go_spec.html",
		"/help.html",
		"/ie.css",
		"/install-source.html",
		"/install.html",
		"/logo-153x55.png",
		"/Makefile",
		"/root.html",
		"/share.png",
		"/sieve.gif",
		"/tos.html",
		"/articles/",
		"/articles/go_command.html",
		"/articles/index.html",
		"/articles/wiki/",
		"/articles/wiki/edit.html",
		"/articles/wiki/final-noclosure.go",
		"/articles/wiki/final-noerror.go",
		"/articles/wiki/final-parsetemplate.go",
		"/articles/wiki/final-template.go",
		"/articles/wiki/final.go",
		"/articles/wiki/get.go",
		"/articles/wiki/http-sample.go",
		"/articles/wiki/index.html",
		"/articles/wiki/Makefile",
		"/articles/wiki/notemplate.go",
		"/articles/wiki/part1-noerror.go",
		"/articles/wiki/part1.go",
		"/articles/wiki/part2.go",
		"/articles/wiki/part3-errorhandling.go",
		"/articles/wiki/part3.go",
		"/articles/wiki/test.bash",
		"/articles/wiki/test_edit.good",
		"/articles/wiki/test_Test.txt.good",
		"/articles/wiki/test_view.good",
		"/articles/wiki/view.html",
		"/codewalk/",
		"/codewalk/codewalk.css",
		"/codewalk/codewalk.js",
		"/codewalk/codewalk.xml",
		"/codewalk/functions.xml",
		"/codewalk/markov.go",
		"/codewalk/markov.xml",
		"/codewalk/pig.go",
		"/codewalk/popout.png",
		"/codewalk/run",
		"/codewalk/sharemem.xml",
		"/codewalk/urlpoll.go",
		"/devel/",
		"/devel/release.html",
		"/devel/weekly.html",
		"/gopher/",
		"/gopher/appenginegopher.jpg",
		"/gopher/appenginegophercolor.jpg",
		"/gopher/appenginelogo.gif",
		"/gopher/bumper.png",
		"/gopher/bumper192x108.png",
		"/gopher/bumper320x180.png",
		"/gopher/bumper480x270.png",
		"/gopher/bumper640x360.png",
		"/gopher/doc.png",
		"/gopher/frontpage.png",
		"/gopher/gopherbw.png",
		"/gopher/gophercolor.png",
		"/gopher/gophercolor16x16.png",
		"/gopher/help.png",
		"/gopher/pkg.png",
		"/gopher/project.png",
		"/gopher/ref.png",
		"/gopher/run.png",
		"/gopher/talks.png",
		"/gopher/pencil/",
		"/gopher/pencil/gopherhat.jpg",
		"/gopher/:pencil/gopherhat.jpg",
		"/gopher/pencil/gopherhelmet.jpg",
		"/gopher/pencil/gophermega.jpg",
		"/gopher/pencil/gopherrunning.jpg",
		"/gopher/pencil/gopherswim.jpg",
		"/gopher/pencil/gopherswrench.jpg",
		"/play/",
		"/play/fib.go",
		"/play/hello.go",
		"/play/life.go",
		"/play/peano.go",
		"/play/pi.go",
		"/play/sieve.go",
		"/play/solitaire.go",
		"/play/tree.go",
		"/progs/",
		"/progs/cgo1.go",
		"/progs/cgo2.go",
		"/progs/cgo3.go",
		"/progs/cgo4.go",
		"/progs/defer.go",
		"/progs/defer.out",
		"/progs/defer2.go",
		"/progs/defer2.out",
		"/progs/eff_bytesize.go",
		"/progs/eff_bytesize.out",
		"/progs/eff_qr.go",
		"/progs/eff_sequence.go",
		"/progs/eff_sequence.out",
		"/progs/eff_unused1.go",
		"/progs/eff_unused2.go",
		"/progs/error.go",
		"/progs/error2.go",
		"/progs/error3.go",
		"/progs/error4.go",
		"/progs/go1.go",
		"/progs/gobs1.go",
		"/progs/gobs2.go",
		"/progs/image_draw.go",
		"/progs/image_package1.go",
		"/progs/image_package1.out",
		"/progs/image_package2.go",
		"/progs/image_package2.out",
		"/progs/image_package3.go",
		"/progs/image_package3.out",
		"/progs/image_package4.go",
		"/progs/image_package4.out",
		"/progs/image_package5.go",
		"/progs/image_package5.out",
		"/progs/image_package6.go",
		"/progs/image_package6.out",
		"/progs/interface.go",
		"/progs/interface2.go",
		"/progs/interface2.out",
		"/progs/json1.go",
		"/progs/json2.go",
		"/progs/json2.out",
		"/progs/json3.go",
		"/progs/json4.go",
		"/progs/json5.go",
		"/progs/run",
		"/progs/slices.go",
		"/progs/timeout1.go",
		"/progs/timeout2.go",
		"/progs/update.bash"}
)

func Benchmark_MatchPattern(b *testing.B) {
	mc := ParseURLPath(mustParseURL("https://www.domain.com/gopher/pencil/gopherhelmet.jpg"))
	var patterns []Pattern
	for _, ps := range staticPaths {
		p, err := ParsePatternImproved(ps, false)
		if err != nil {
			b.Fatalf("MatchPatterns() pattern: [%s] parse error: %v", ps, err)
		}
		patterns = append(patterns, p)
	}
	sort.SliceStable(patterns, func(i, j int) bool {
		return patterns[i].HighPriorityThan(patterns[j])
	})

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if p, found := mc.MatchPatterns(patterns); found {
			useFoundPatter(p)
		}
	}
}

func useFoundPatter(p Pattern) {

}

func TestParseURLPath(t *testing.T) {
	tests := []struct {
		name       string
		requestUrl *url.URL
		want       MatchingContext
	}{
		{
			name:       "parse empty",
			requestUrl: mustParseURL("https://example.com"),
			want: MatchingContext{
				originalPath: "",
				pathSegments: nil,
				pathEncode:   newUrlPathEncode(0),
			},
		},
		{
			name:       "/",
			requestUrl: mustParseURL("https://example.com/?q=morefoo%25bar"),
			want: MatchingContext{
				originalPath: "/",
				pathSegments: nil,
				pathEncode:   newUrlPathEncode(0),
			},
		},
		{
			name:       "/abc",
			requestUrl: mustParseURL("https://example.com/abc?q=morefoo%25bar"),
			want: MatchingContext{
				originalPath: "/abc",
				pathSegments: []string{"abc"},
				pathEncode:   newUrlPathEncode(1),
			},
		},
		{
			name:       "//",
			requestUrl: mustParseURL("https://example.com//?q=morefoo%25bar"),
			want: MatchingContext{
				originalPath: "//",
				pathSegments: nil,
				pathEncode:   newUrlPathEncode(0),
			},
		},
		{
			name:       "//abc",
			requestUrl: mustParseURL("https://example.com//abc?q=morefoo%25bar"),
			want: MatchingContext{
				originalPath: "//abc",
				pathSegments: []string{"abc"},
				pathEncode:   newUrlPathEncode(1),
			},
		},
		{
			name:       "/abc/",
			requestUrl: mustParseURL("https://example.com/abc/?q=morefoo%25bar"),
			want: MatchingContext{
				originalPath: "/abc/",
				pathSegments: []string{"abc"},
				pathEncode:   newUrlPathEncode(1),
			},
		},
		{
			name:       "/abc//def",
			requestUrl: mustParseURL("https://example.com/abc//def?q=morefoo%25bar"),
			want: MatchingContext{
				originalPath: "/abc//def",
				pathSegments: []string{"abc", "def"},
				pathEncode:   newUrlPathEncode(2),
			},
		},
		{
			name:       "/foo%25fbar",
			requestUrl: mustParseURL("https://example.com/foo%25fbar?q=morefoo%25bar"),
			want: MatchingContext{
				originalPath: "/foo%fbar",
				pathSegments: []string{"foo%fbar"},
				pathEncode:   newUrlPathEncode(1),
			},
		},
		{
			name:       "/foo%2fbar",
			requestUrl: mustParseURL("https://example.com/foo%2fbar?q=morefoo%25bar"),
			want: MatchingContext{
				originalPath: "/foo/bar",
				pathSegments: []string{"foo", "bar"},
				pathEncode:   newUrlPathEncode(2),
			},
		},
		{
			name:       "/path/to/new/../file",
			requestUrl: mustParseURL("https://example.com/path/to/new/../file"),
			want: MatchingContext{
				originalPath: "/path/to/new/../file",
				pathSegments: []string{"path", "to", "file"},
				pathEncode:   newUrlPathEncode(3),
			},
		},
		{
			name:       "/foo/../../bar",
			requestUrl: mustParseURL("https://example.com/foo/../../bar"),
			want: MatchingContext{
				originalPath: "/foo/../../bar",
				pathSegments: []string{"bar"},
				pathEncode:   newUrlPathEncode(1),
			},
		},
		{
			name:       "/foo/../..",
			requestUrl: mustParseURL("https://example.com/foo/../.."),
			want: MatchingContext{
				originalPath: "/foo/../..",
				pathSegments: nil,
				pathEncode:   newUrlPathEncode(0),
			},
		},
		{
			name:       "/foo/.%2e",
			requestUrl: mustParseURL("https://example.com/foo/.%2e"),
			want: MatchingContext{
				originalPath: "/foo/..",
				pathSegments: nil,
				pathEncode:   newUrlPathEncode(0),
			},
		},
		{
			name:       "/foo/%2e%2e",
			requestUrl: mustParseURL("https://example.com/foo/%2e%2e"),
			want: MatchingContext{
				originalPath: "/foo/..",
				pathSegments: nil,
				pathEncode:   newUrlPathEncode(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseURLPath(tt.requestUrl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseURLPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchingContextImproved_MatchPattern(t *testing.T) {
	type want struct {
		pattern       string
		urlPathEncode encode
		captureVars   map[string]string
		found         bool
	}
	tests := []struct {
		name     string
		urlPath  string
		patterns []string
		want     want
	}{
		{
			name:     "no root pattern, matches /",
			urlPath:  "/",
			patterns: []string{"/a"},
			want: want{
				pattern:       "",
				urlPathEncode: encode{val: 0, len: 0},
				found:         false,
			},
		},
		{
			name:     "root pattern, matches /",
			urlPath:  "/",
			patterns: []string{"/"},
			want: want{
				pattern:       "/",
				urlPathEncode: encode{val: 0, len: 0},
				found:         true,
			},
		},
		{
			name:     "root pattern, matches / with back segments",
			urlPath:  "/a/../b/../",
			patterns: []string{"/"},
			want: want{
				pattern:       "/",
				urlPathEncode: encode{val: 0, len: 0},
				found:         true,
			},
		},
		{
			name:     "2 literal patterns, second should match",
			urlPath:  "/a/b/",
			patterns: []string{"/b/c", "/a/b"},
			want: want{
				pattern:       "/a/b",
				urlPathEncode: encode{val: 11, len: 2},
				found:         true,
			},
		},
		{
			name:     "2 literal patterns, none should match",
			urlPath:  "/a/r/",
			patterns: []string{"/b/c", "/a/b"},
			want: want{
				pattern:       "",
				urlPathEncode: encode{val: 99, len: 2},
				found:         false,
			},
		},
		{
			name:     "greedy pattern, matches /",
			urlPath:  "/",
			patterns: []string{"/**"},
			want: want{
				pattern:       "/**",
				urlPathEncode: encode{val: 0, len: 0},
				found:         true,
			},
		},
		{
			name:     "greedy pattern, matches / with back segments",
			urlPath:  "/a/../b/../",
			patterns: []string{"/**"},
			want: want{
				pattern:       "/**",
				urlPathEncode: encode{val: 0, len: 0},
				found:         true,
			},
		},
		{
			name:     "1 pattern, multiple greedy segments, different path segments",
			urlPath:  "/bla/testing/testing/bla/bla",
			patterns: []string{"/bla/**/testing/**/bla"},
			want: want{
				pattern:       "/bla/**/testing/**/bla",
				urlPathEncode: encode{val: 11661, len: 5},
				found:         true,
			},
		},
		{
			name:     "1 pattern, multiple greedy segments, different path segments, should not match",
			urlPath:  "/bla/testing/testing/bla/blaa",
			patterns: []string{"/bla/**/testing/**/bla"},
			want: want{
				pattern:       "",
				urlPathEncode: encode{val: 99999, len: 5},
				found:         false,
			},
		},
		{
			name:     "1 pattern, multiple greedy segments, same path segments",
			urlPath:  "/bla/bla/bla/bla/bla/bla",
			patterns: []string{"/bla/**/bla/**/bla"},
			want: want{
				pattern:       "/bla/**/bla/**/bla",
				urlPathEncode: encode{val: 116661, len: 6},
				found:         true,
			},
		},
		{
			name:     "2 pattern, with multiple greedy segments but the first one is more specific",
			urlPath:  "/bla/testing/testing/bla/bla",
			patterns: []string{"/bla/**/testing/**/bla", "/bla/**/bla"},
			want: want{
				pattern:       "/bla/**/testing/**/bla",
				urlPathEncode: encode{val: 11661, len: 5},
				found:         true,
			},
		},
		{
			name:     "3 patterns, the same length greedy pattern should match",
			urlPath:  "/a/b/e",
			patterns: []string{"/a/b/c", "/a/*/d", "/a/**"},
			want: want{
				pattern:       "/a/**",
				urlPathEncode: encode{val: 166, len: 3},
				found:         true,
			},
		},
		{
			name:     "3 patterns, greedy pattern should match",
			urlPath:  "/a/b/c/d/e",
			patterns: []string{"/a/b", "/a/*/*/d", "/a/**/e"},
			want: want{
				pattern:       "/a/**/e",
				urlPathEncode: encode{val: 16661, len: 5},
				found:         true,
			},
		},
		{
			name:     "2 patterns, second should match",
			urlPath:  "/a/ab",
			patterns: []string{"/a/{b}", "/a/b"},
			want: want{
				pattern:       "/a/{b}",
				urlPathEncode: encode{val: 13, len: 2},
				captureVars:   map[string]string{"b": "ab"},
				found:         true,
			},
		},
		{
			name:     "1 pattern, with regex and greedy match, the url is ending with with /",
			urlPath:  "/XXXblaXXXX/testing/testing/bla/testing/testing/",
			patterns: []string{"/*bla*/**/bla/**"},
			want: want{
				pattern:       "/*bla*/**/bla/**",
				urlPathEncode: encode{val: 466166, len: 6},
				found:         true,
			},
		},
		{
			name:     "1 pattern, with regex and greedy match",
			urlPath:  "/XXXblaXXXX/testing/testing/bla/testing/testing.jpg",
			patterns: []string{"/*bla*/**/bla/**"},
			want: want{
				pattern:       "/*bla*/**/bla/**",
				urlPathEncode: encode{val: 466166, len: 6},
				found:         true,
			},
		},
		{
			name:     "1 pattern, with capture var and greedy match",
			urlPath:  "/a/b/c/d/e/f/g/h",
			patterns: []string{"/a/{b}/{c}/**/g/h"},
			want: want{
				pattern:       "/a/{b}/{c}/**/g/h",
				urlPathEncode: encode{val: 13366611, len: 8},
				captureVars:   map[string]string{"b": "b", "c": "c"},
				found:         true,
			},
		},
		{
			name:     "2 patterns, with capture var and greedy match",
			urlPath:  "/a/b/c/d/e/f/g",
			patterns: []string{"/a/{b}/{c}/*/f/g", "/a/{c}/{b}/**/f/g"},
			want: want{
				pattern:       "/a/{c}/{b}/**/f/g",
				urlPathEncode: encode{val: 1336611, len: 7},
				captureVars:   map[string]string{"b": "c", "c": "b"},
				found:         true,
			},
		},
		{
			name:     "4 patterns, with capture var, regex and greedy match",
			urlPath:  "/a/b/c/d/e/f/g",
			patterns: []string{"/a/{b}/{c}/{d}/f/g", "/a/{c}/{b}/**/f/{d}", "/a/{c}/{b}/**/f/g", "/a/{c}/{b}/*/{d}/f/g"},
			want: want{
				pattern:       "/a/{c}/{b}/*/{d}/f/g",
				urlPathEncode: encode{val: 1335311, len: 7},
				captureVars:   map[string]string{"b": "c", "c": "b", "d": "e"},
				found:         true,
			},
		},
		{
			name:     "4 patterns, with capture var, regex and greedy match should not match",
			urlPath:  "/a/b/c/d/e/f/h",
			patterns: []string{"/a/{b}/{c}/{d}/f/g", "/a/{c}/{b}/**/f/{d}/q", "/a/{c}/{b}/**/f/g", "/a/{c}/{b}/*/{d}/f/g"},
			want: want{
				pattern:       "",
				urlPathEncode: encode{val: 9999999, len: 7},
				found:         false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := ParseURLPath(mustParseURL("https://www.domain.com" + tt.urlPath))
			var patterns []Pattern
			for _, ps := range tt.patterns {
				p, err := ParsePatternImproved(ps, false)
				if err != nil {
					t.Fatalf("MatchPatterns() pattern: [%s] parse error: %v", ps, err)
				}
				patterns = append(patterns, p)
			}
			sort.SliceStable(patterns, func(i, j int) bool {
				return patterns[i].HighPriorityThan(patterns[j])
			})
			p, found := mc.MatchPatterns(patterns)
			got := want{
				pattern:       p.RawValue,
				urlPathEncode: mc.pathEncode,
				captureVars:   mc.CaptureVars,
				found:         found,
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MatchPatterns() got = %v, want %v", got, tt.want)
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
