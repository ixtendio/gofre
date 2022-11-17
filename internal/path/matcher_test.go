package path

import (
	"log"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestMatcher_AddPattern(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     string
		wantErr  bool
	}{
		{
			name:     "/ path matcher",
			patterns: []string{"/"},
			want:     "R",
		},
		{
			name:     "double / path matcher, returns error",
			patterns: []string{"/", "/"},
			wantErr:  true,
		},
		{
			name:     "2 literal patterns with one segment, that are the same",
			patterns: []string{"/a", "/a"},
			wantErr:  true,
			want:     "R=>(a:1)",
		},
		{
			name:     "1 literal pattern",
			patterns: []string{"/a"},
			want:     "R=>(a:1)",
		},
		{
			name:     "2 literal patterns with one segment, that are different",
			patterns: []string{"/a", "/b"},
			want:     "R=>(a:1 b:1)",
		},
		{
			name:     "2 literal patterns with multiple segments, that are the same",
			patterns: []string{"/a/b/c/d/e", "/a/b/c/d/e"},
			wantErr:  true,
		},
		{
			name:     "2 literal patterns with multiple segments, that are different",
			patterns: []string{"/a/b/c/d/e", "/a/b/c/d/f"},
			want:     "R=>(a:5=>(b:4=>(c:3=>(d:2=>(e:1 f:1)))))",
		},
		{
			name:     "2 literal patterns with multiple segments, all segments the same",
			patterns: []string{"/a/a/a/a/a"},
			want:     "R=>(a:5=>(a:4=>(a:3=>(a:2=>(a:1)))))",
		},
		{
			name:     "4 literal patterns with multiple segments, 3 of them have common prefix",
			patterns: []string{"/a/b/c", "/a/b/d", "/a/b/d/e", "/b/d"},
			want:     "R=>(b:2=>(d:1) a:4=>(b:3=>(c:1 d:2=>(e:1))))",
		},
		{
			name:     "1 literal, 1 var capture, all of them have common prefix",
			patterns: []string{"/a/{b}", "/a"},
			want:     "R=>(a:2=>({b}:1))",
		},
		{
			name:     "1 literal, 1 var capture and 1 var capture with constraint patterns with multiple segments, all of them have common prefix",
			patterns: []string{"/a/b/c", "/a/{b}/c", "/a/{b:[a-z]+}/d/e"},
			want:     "R=>(a:4=>(b:2=>(c:1) {b:[a-z]+}:3=>(d:2=>(e:1)) {b}:2=>(c:1)))",
		},
		{
			name:     "1 literal, 1 regex 1 var capture and 1 greedy patterns with multiple segments, all of them have common prefix",
			patterns: []string{"/a/**/c/e", "/a/b/c", "/a/a?c*/c/f"},
			want:     "R=>(a:255=>(b:2=>(c:1) a?c*:3=>(c:2=>(f:1)) **:255=>(c:255=>(e:255))))",
		},
		{
			name:     "2 greedy patterns with multiple segments, all of them have common prefix",
			patterns: []string{"/a/**/c", "/a/**/b"},
			want:     "R=>(a:255=>(**:255=>(b:255 c:255)))",
		},
	}
	for _, tt := range tests {
		mc := NewMatcher()
		t.Run(tt.name, func(t *testing.T) {
			var err error
			for _, ps := range tt.patterns {
				p := mustParsePattern(ps)
				if err = mc.AddPattern(p); err != nil {
					break
				}
			}
			if tt.wantErr {
				if err == nil {
					t.Errorf("Matcher.AddPattern() got no error but want error")
				}
			} else {
				if err != nil {
					t.Errorf("Matcher.AddPattern() got error: %v", err)
					return
				}
				got := matcherString(mc)
				if got != tt.want {
					t.Errorf("Matcher.AddPattern() got: %v, want: %v", got, tt.want)
				}
			}
		})
	}
}

func TestMatcher_Match(t *testing.T) {
	type want struct {
		matchedPattern       string
		urlSegmentsMatchType int
		captureVars          []CaptureVar
	}
	tests := []struct {
		name     string
		patterns []string
		args     string
		want     want
	}{
		{
			name:     "/ match",
			patterns: []string{"/"},
			args:     "/",
			want: want{
				matchedPattern:       "/",
				urlSegmentsMatchType: 0,
			},
		},
		{
			name:     "1 literal pattern that match",
			patterns: []string{"/a/b/c"},
			args:     "/a/b/c",
			want: want{
				matchedPattern:       "/a/b/c",
				urlSegmentsMatchType: 111,
			},
		},
		{
			name:     "1 literal pattern that not match",
			patterns: []string{"/a/b/c"},
			args:     "/a/b/c/d",
			want: want{
				matchedPattern: "",
			},
		},
		{
			name:     "2 literal patterns, same segments length, second pattern matches",
			patterns: []string{"/a/b/c", "/a/b/d"},
			args:     "/a/b/d",
			want: want{
				matchedPattern:       "/a/b/d",
				urlSegmentsMatchType: 111,
			},
		},
		{
			name:     "3 literal patterns, different segments length, third matches",
			patterns: []string{"/a/b/c/d/e", "/a/b/c/e/e", "/a/b/c/d/d"},
			args:     "/a/b/c/d/d",
			want: want{
				matchedPattern:       "/a/b/c/d/d",
				urlSegmentsMatchType: 11111,
			},
		},
		{
			name:     "1 literal and 1 capture var, the capture var matches",
			patterns: []string{"/a/b/c/", "/a/{b}/e"},
			args:     "/a/b/e",
			want: want{
				matchedPattern:       "/a/{b}/e",
				urlSegmentsMatchType: 131,
				captureVars:          []CaptureVar{{Name: "b", Value: "b"}},
			},
		},
		{
			name:     "1 literal and 2 capture var, the second capture var matches",
			patterns: []string{"/a/b/f/d", "/a/{b:[a-z]+}/e", "/a/{b}/f"},
			args:     "/a/b/f",
			want: want{
				matchedPattern:       "/a/{b}/f",
				urlSegmentsMatchType: 131,
				captureVars:          []CaptureVar{{Name: "b", Value: "b"}},
			},
		},
		{
			name:     "1 literal 2 capture vars and 1 pattern with single path match, the pattern with single path match matches",
			patterns: []string{"/a/b/*/d", "/a/b/f/d", "/a/{b:[a-z]+}/e", "/a/{b}/f/d"},
			args:     "/a/b/g/d",
			want: want{
				matchedPattern:       "/a/b/*/d",
				urlSegmentsMatchType: 1151,
			},
		},
		{
			name:     "1 literal 2 capture vars 1 pattern with single path match and 1 regex, the pattern with regex matches",
			patterns: []string{"/a/b/?/d", "/a/b/*/d", "/a/b/f/d", "/a/{b:[a-z]+}/e", "/a/{b}/f/d"},
			args:     "/a/b/g/d",
			want: want{
				matchedPattern:       "/a/b/?/d",
				urlSegmentsMatchType: 1141,
			},
		},
		{
			name:     "1 literal 1 regex and 1 greedy pattern, the greedy pattern matches",
			patterns: []string{"/a/b", "/a/*/*/d", "/a/**/e"},
			args:     "/a/b/c/d/e",
			want: want{
				matchedPattern:       "/a/**/e",
				urlSegmentsMatchType: 16661,
			},
		},
		{
			name:     "3 greedy patterns, the most specific matches",
			patterns: []string{"/a/**", "/a/**/d/**/e", "/a/**/e"},
			args:     "/a/b/c/d/e/e",
			want: want{
				matchedPattern:       "/a/**/d/**/e",
				urlSegmentsMatchType: 166161,
			},
		},
		{
			name:     "1 greedy pattern with prefix and suffix, should match",
			patterns: []string{"/bla/**/bla"},
			args:     "/bla/testing/testing/bla/bla/bla",
			want: want{
				matchedPattern:       "/bla/**/bla",
				urlSegmentsMatchType: 166661,
			},
		},
		{
			name:     "1 greedy pattern with prefix and suffix, should not match",
			patterns: []string{"/bla/**/bla"},
			args:     "/bla/testing/testing/bla/bla/blue",
			want: want{
				matchedPattern: "",
			},
		},
		{
			name:     "1 pattern with regex and greedy, should match",
			patterns: []string{"/*bla*/**/bla/**"},
			args:     "/XXXblaXXXX/testing/testing/bla/testing/testing",
			want: want{
				matchedPattern:       "/*bla*/**/bla/**",
				urlSegmentsMatchType: 466166,
			},
		},
		{
			name:     "1 pattern with regex and greedy, should not match",
			patterns: []string{"/*bla*/**/bla/**"},
			args:     "/XXXblaXXXX/testing/testing/blue/testing/testing",
			want: want{
				matchedPattern: "",
			},
		},
		{
			name:     "4 patterns greedy and capture vars, should not match",
			patterns: []string{"/a/{b}/{c}/{d}/f/g", "/a/{c}/{b}/**/f/{d}", "/a/{c}/{b}/**/f/g", "/a/{c}/{b}/*/{d}/f/g"},
			args:     "/a/b/c/d/e/f/g",
			want: want{
				matchedPattern:       "/a/{c}/{b}/*/{d}/f/g",
				urlSegmentsMatchType: 1335311,
				captureVars:          []CaptureVar{{Name: "c", Value: "b"}, {Name: "b", Value: "c"}, {Name: "d", Value: "e"}},
			},
		},
	}
	for _, tt := range tests {
		m := NewMatcher()
		t.Run(tt.name, func(t *testing.T) {
			reqUrl, err := url.Parse("https://www.domain.com" + tt.args)
			if err != nil {
				t.Fatalf("Match() got error: %v at url parsing", err)
			}

			for _, ps := range tt.patterns {
				if err := m.AddPattern(mustParsePattern(ps)); err != nil {
					t.Fatalf("Match() got error: %v at pattern registration", err)
				}
			}
			var got want
			mc := ParseURLPath(reqUrl)
			if p := m.Match(&mc); p != nil {
				got.matchedPattern = p.RawValue
				got.captureVars = mc.captureVars
			}
			if len(tt.want.matchedPattern) > 0 {
				var urlSegmentsMatchType int
				for _, t := range mc.pathSegments {
					urlSegmentsMatchType = urlSegmentsMatchType*10 + int(t.matchType)
				}
				got.urlSegmentsMatchType = urlSegmentsMatchType
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Match() got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func matcherString(mc *Matcher) string {
	var sb strings.Builder
	sb.WriteString("R")
	trieToString(mc, mc.trieRoot.children, &sb)
	return strings.TrimSpace(sb.String())
}

func trieToString(mc *Matcher, children []*node, sb *strings.Builder) {
	childrenLen := len(children)
	if childrenLen == 0 {
		return
	}
	sb.WriteString("=>(")
	for i, child := range children {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(child.val)
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(int(child.maxMatchableSegments)))
		trieToString(mc, child.children, sb)
	}
	sb.WriteString(")")
}

func mustParsePattern(pattern string) *Pattern {
	p, err := ParsePattern(pattern, false)
	if err != nil {
		log.Fatalf("Matcher.AddPattern() parse pattern error, err: %v", err)
	}
	return p
}
