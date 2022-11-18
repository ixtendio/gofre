package path

//
//import (
//	"math"
//	"reflect"
//	"testing"
//)
//
//func Benchmark_Map(b *testing.B) {
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		m := make(map[string]string, 4)
//		m["key1"] = "val1"
//		m["key2"] = "val2"
//		m["key3"] = "val3"
//		m["key4"] = "val4"
//	}
//}
//
//func Benchmark_Array(b *testing.B) {
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		m := make([]string, 8)
//		m[0] = "key1"
//		m[1] = "val1"
//		m[2] = "key2"
//		m[3] = "val2"
//		m[4] = "key3"
//		m[5] = "val3"
//		m[6] = "key4"
//		m[7] = "val4"
//	}
//}
//
//func Benchmark_uint64(b *testing.B) {
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		_ = 1234
//	}
//}
//
//func Benchmark_ParsePatternImproved(b *testing.B) {
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		_, err := ParsePattern("/gopher/pencil/pencil1/pencil1/pencil1/pencil1/pencil1/gopherhat.jpg", false)
//		//_, err := ParsePattern("/gopher/{pencil}/{pencil1}/{pencil1}/{pencil1}/{pencil1}/{pencil1}/gopherhat.jpg", false)
//		if err != nil {
//			b.Errorf("%v", err)
//		}
//	}
//}
//
//func Test_ParsePatternImproved(t *testing.T) {
//	type captureVar struct {
//		segmentIndex int
//		name         string
//		hasPattern   bool
//	}
//	type want struct {
//		segmentsCount                  int
//		maxMatchableSegment            int
//		pathSegmentsMatchTypesEncoding encode
//		rawValue                       string
//		captureVars                    []captureVar
//	}
//	type patterns struct {
//		pathPattern     string
//		caseInsensitive bool
//	}
//	tests := []struct {
//		name    string
//		patterns    patterns
//		want    want
//		wantErr bool
//	}{
//		{
//			name:    "path only with greedy matchType",
//			patterns:    patterns{pathPattern: "/**/**"},
//			want:    want{},
//			wantErr: true,
//		},
//		{
//			name:    "path with consecutive greedy matchType segments",
//			patterns:    patterns{pathPattern: "/**/a/**/**/b"},
//			want:    want{},
//			wantErr: true,
//		},
//		{
//			name:    "path that not starts with slash",
//			patterns:    patterns{pathPattern: "abc/"},
//			want:    want{},
//			wantErr: true,
//		},
//		{
//			name: "path with capture variable without regex",
//			patterns: patterns{pathPattern: "/abc/{id}"},
//			want: want{
//				segmentsCount:                  2,
//				maxMatchableSegment:            2,
//				pathSegmentsMatchTypesEncoding: encode{val: 1300000000000000000, len: MaxPathSegments},
//				rawValue:                       "/abc/{id}",
//				captureVars: []captureVar{{
//					segmentIndex: 1,
//					name:         "id",
//					hasPattern:   false,
//				}},
//			},
//			wantErr: false,
//		},
//		{
//			name: "path with capture variable and regex",
//			patterns: patterns{pathPattern: "/abc/{id:\\d}"},
//			want: want{
//				segmentsCount:                  2,
//				maxMatchableSegment:            2,
//				pathSegmentsMatchTypesEncoding: encode{val: 1200000000000000000, len: MaxPathSegments},
//				rawValue:                       "/abc/{id:\\d}",
//				captureVars: []captureVar{{
//					segmentIndex: 1,
//					name:         "id",
//					hasPattern:   true,
//				}},
//			},
//			wantErr: false,
//		},
//		{
//			name: "root path",
//			patterns: patterns{pathPattern: "/"},
//			want: want{
//				segmentsCount:                  0,
//				maxMatchableSegment:            0,
//				pathSegmentsMatchTypesEncoding: encode{},
//				rawValue:                       "/",
//				captureVars:                    nil,
//			},
//			wantErr: false,
//		},
//		{
//			name:    "root path with double slash",
//			patterns:    patterns{pathPattern: "//"},
//			want:    want{},
//			wantErr: true,
//		},
//		{
//			name:    "path with many slash",
//			patterns:    patterns{pathPattern: "/abc///cde////"},
//			want:    want{},
//			wantErr: true,
//		},
//		{
//			name: "path with single asterix",
//			patterns: patterns{pathPattern: "/a/*"},
//			want: want{
//				segmentsCount:                  2,
//				maxMatchableSegment:            2,
//				pathSegmentsMatchTypesEncoding: encode{val: 1500000000000000000, len: MaxPathSegments},
//				rawValue:                       "/a/*",
//				captureVars:                    nil,
//			},
//			wantErr: false,
//		}, {
//			name: "path with max segments",
//			patterns: patterns{pathPattern: "/a/*/b/{q}/{y:[a-z]+}/?as*/d/e/f/g/{t}/i/j/k/l/m/n/o/{w:[a-z]+}"},
//			want: want{
//				segmentsCount:                  19,
//				maxMatchableSegment:            19,
//				pathSegmentsMatchTypesEncoding: encode{val: 1513241111311111112, len: MaxPathSegments},
//				rawValue:                       "/a/*/b/{q}/{y:[a-z]+}/?as*/d/e/f/g/{t}/i/j/k/l/m/n/o/{w:[a-z]+}",
//				captureVars: []captureVar{{
//					segmentIndex: 3,
//					name:         "q",
//				}, {
//					segmentIndex: 4,
//					name:         "y",
//					hasPattern:   true,
//				}, {
//					segmentIndex: 10,
//					name:         "t",
//				}, {
//					segmentIndex: 18,
//					name:         "w",
//					hasPattern:   true,
//				}},
//			},
//			wantErr: false,
//		},
//		{
//			name: "path with double asterix at start",
//			patterns: patterns{pathPattern: "/**/a"},
//			want: want{
//				segmentsCount:                  2,
//				maxMatchableSegment:            math.MaxInt,
//				pathSegmentsMatchTypesEncoding: encode{val: 6666666666666666661, len: MaxPathSegments},
//				rawValue:                       "/**/a",
//				captureVars:                    nil,
//			},
//			wantErr: false,
//		},
//		{
//			name: "path with double asterix at the end",
//			patterns: patterns{pathPattern: "/a/**"},
//			want: want{
//				segmentsCount:                  2,
//				maxMatchableSegment:            math.MaxInt,
//				pathSegmentsMatchTypesEncoding: encode{val: 1666666666666666666, len: MaxPathSegments},
//				rawValue:                       "/a/**",
//				captureVars:                    nil,
//			},
//			wantErr: false,
//		},
//		{
//			name: "path with double asterix in the middle",
//			patterns: patterns{pathPattern: "/a/**/b"},
//			want: want{
//				segmentsCount:                  3,
//				maxMatchableSegment:            math.MaxInt,
//				pathSegmentsMatchTypesEncoding: encode{val: 1666666666666666661, len: MaxPathSegments},
//				rawValue:                       "/a/**/b",
//				captureVars:                    nil,
//			},
//			wantErr: false,
//		},
//		{
//			name: "path with multiple double asterix segments",
//			patterns: patterns{pathPattern: "/a/**/b/**/c/**/d/**/e/**/f/g/h"},
//			want: want{
//				segmentsCount:                  13,
//				maxMatchableSegment:            math.MaxInt,
//				pathSegmentsMatchTypesEncoding: encode{val: 1666166166166166111, len: MaxPathSegments},
//				rawValue:                       "/a/**/b/**/c/**/d/**/e/**/f/g/h",
//				captureVars:                    nil,
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p, err := ParsePattern(tt.patterns.pathPattern, tt.patterns.caseInsensitive)
//			var captureVars []captureVar
//			for _, cp := range p.captureVars {
//				var hasPattern bool
//				if cp.pattern != nil {
//					hasPattern = true
//				}
//				captureVars = append(captureVars, captureVar{
//					segmentIndex: cp.segmentIndex,
//					name:         cp.name,
//					hasPattern:   hasPattern,
//				})
//			}
//			got := want{
//				segmentsCount:                  p.segmentsCount,
//				maxMatchableSegment:            p.maxMatchableSegments,
//				pathSegmentsMatchTypesEncoding: p.pathEncode,
//				rawValue:                       p.RawValue,
//				captureVars:                    captureVars,
//			}
//			if tt.wantErr {
//				if err == nil {
//					t.Errorf("ParsePattern() got nil error, wantErr %v", tt.wantErr)
//				}
//			} else {
//				if !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("ParsePattern() got: %v, want: %v", got, tt.want)
//				}
//			}
//		})
//	}
//}
//
//func Test_validatePathSegment(t *testing.T) {
//	tests := []struct {
//		name        string
//		pathSegment string
//		wantErr     bool
//	}{
//		{
//			name:        "empty pattern",
//			pathSegment: "",
//			wantErr:     true,
//		},
//		{
//			name:        "literal pattern",
//			pathSegment: "a",
//			wantErr:     false,
//		},
//		{
//			name:        "capture var pattern without open bracket",
//			pathSegment: "a}",
//			wantErr:     true,
//		},
//		{
//			name:        "capture var pattern without closed bracket",
//			pathSegment: "{c",
//			wantErr:     true,
//		},
//		{
//			name:        "capture var pattern without name",
//			pathSegment: "{}",
//			wantErr:     true,
//		},
//		{
//			name:        "capture var pattern with constraint but without name",
//			pathSegment: "{:[a-z]+}",
//			wantErr:     true,
//		},
//		{
//			name:        "capture var pattern with constraint regex",
//			pathSegment: "{a:}",
//			wantErr:     true,
//		},
//		{
//			name:        "capture var pattern with constraint regex and without name",
//			pathSegment: "{:}",
//			wantErr:     true,
//		},
//		{
//			name:        "path Segment with triple asterix",
//			pathSegment: "***",
//			wantErr:     true,
//		},
//		{
//			name:        "path Segment with double asterix and another text at start",
//			pathSegment: "**abc",
//			wantErr:     true,
//		},
//		{
//			name:        "path Segment with double asterix and another text",
//			pathSegment: "abc**def",
//			wantErr:     true,
//		},
//		{
//			name:        "path Segment with double asterix and another text at the end",
//			pathSegment: "bla**",
//			wantErr:     true,
//		},
//		{
//			name:        "valid capture var pattern without constraint",
//			pathSegment: "{abc}",
//			wantErr:     false,
//		},
//		{
//			name:        "valid capture var pattern with constraint",
//			pathSegment: "{abc:[a-z]+}",
//			wantErr:     false,
//		},
//		{
//			name:        "valid capture var pattern with constraint and nested brackets",
//			pathSegment: "{abc:[a-z]{3}}",
//			wantErr:     false,
//		},
//		{
//			name:        "valid path Segment with regex ?",
//			pathSegment: "?asd",
//			wantErr:     false,
//		},
//		{
//			name:        "valid path Segment with regex * at beginning",
//			pathSegment: "*asd",
//			wantErr:     false,
//		},
//		{
//			name:        "valid path Segment with regex *",
//			pathSegment: "a*sd",
//			wantErr:     false,
//		}, {
//			name:        "valid path Segment with regex * at the end",
//			pathSegment: "asd*",
//			wantErr:     false,
//		},
//		{
//			name:        "valid path Segment with one asterix",
//			pathSegment: "*",
//			wantErr:     false,
//		},
//		{
//			name:        "valid path Segment with double asterix",
//			pathSegment: "**",
//			wantErr:     false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := validatePathSegment(tt.pathSegment); (err != nil) != tt.wantErr {
//				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func Test_createCaptureVarsContainers(t *testing.T) {
//	tests := []struct {
//		name        string
//		pathPattern string
//		want        int
//	}{
//		{
//			name:        "no capture vars",
//			pathPattern: "/a/b/c",
//			want:        0,
//		},
//		{
//			name:        "one capture vars",
//			pathPattern: "/a/{b}/c",
//			want:        1,
//		},
//		{
//			name:        "two capture vars",
//			pathPattern: "/a/{b}/{c}",
//			want:        2,
//		},
//		{
//			name:        "single capture var",
//			pathPattern: "/{c}",
//			want:        1,
//		},
//		{
//			name:        "single capture var with regex",
//			pathPattern: "/{c:\\d}",
//			want:        1,
//		},
//		{
//			name:        "two capture var with regex",
//			pathPattern: "/{c:\\d}/{d:\\w}",
//			want:        2,
//		}, {
//			name:        "two capture var from which one with regex",
//			pathPattern: "/{c:\\d}/{d}",
//			want:        2,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got := createCaptureVarsContainers(tt.pathPattern)
//			if len(got) != tt.want {
//				t.Errorf("createCaptureVarsContainers().captureVarsLen = %v, want %v", len(got), tt.want)
//			}
//		})
//	}
//}
//
//func Test_determineMatchTypeForSegment(t *testing.T) {
//	tests := []struct {
//		name        string
//		pathSegment string
//		want        MatchType
//	}{
//		{
//			name:        "MatchTypeSingleSegment",
//			pathSegment: "*",
//			want:        MatchTypeSingleSegment,
//		},
//		{
//			name:        "MatchTypeMultipleSegments",
//			pathSegment: "**",
//			want:        MatchTypeMultipleSegments,
//		},
//		{
//			name:        "MatchTypeConstraintCaptureVar",
//			pathSegment: "{abc:[a-z]+}",
//			want:        MatchTypeConstraintCaptureVar,
//		},
//		{
//			name:        "MatchTypeCaptureVar",
//			pathSegment: "{abc}",
//			want:        MatchTypeCaptureVar,
//		},
//		{
//			name:        "MatchTypeRegex ?",
//			pathSegment: "abc?asd",
//			want:        MatchTypeRegex,
//		},
//		{
//			name:        "MatchTypeRegex *",
//			pathSegment: "abc*asd",
//			want:        MatchTypeRegex,
//		},
//		{
//			name:        "MatchTypeLiteral",
//			pathSegment: "abcasd",
//			want:        MatchTypeLiteral,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := determineMatchTypeForSegment(tt.pathSegment); got != tt.want {
//				t.Errorf("determineMatchTypeForSegment() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestPattern_getSegmentMatchType(t *testing.T) {
//	tests := []struct {
//		name         string
//		pattern      string
//		segmentIndex int
//		want         MatchType
//	}{
//		{
//			name:         "index out of bounds",
//			pattern:      "/",
//			segmentIndex: 0,
//			want:         MatchTypeUnknown,
//		},
//		{
//			name:         "index 0",
//			pattern:      "/a/*/b/{q}/{y:[a-z]+}/?as*/d/e/f/g/{t}/i/j/k/l/m/n/o/{w:[a-z]+}",
//			segmentIndex: 0,
//			want:         MatchTypeLiteral,
//		},
//		{
//			name:         "index 18",
//			pattern:      "/a/*/b/{q}/{y:[a-z]+}/?as*/d/e/f/g/{t}/i/j/k/l/m/n/o/{w:[a-z]+}",
//			segmentIndex: 18,
//			want:         MatchTypeConstraintCaptureVar,
//		},
//		{
//			name:         "index 3",
//			pattern:      "/a/*/b/{q}/{y:[a-z]+}/?as*/d/e/f/g/{t}/i/j/k/l/m/n/o/{w:[a-z]+}",
//			segmentIndex: 3,
//			want:         MatchTypeCaptureVar,
//		},
//		{
//			name:         "index 4",
//			pattern:      "/a/**/b/**/{c}/**/d/**/e/**/f/g/h",
//			segmentIndex: 4,
//			want:         MatchTypeCaptureVar,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p, err := ParsePattern(tt.pattern, false)
//			if err != nil {
//				t.Fatalf("getSegmentMatchType() got err: %v", err)
//			}
//			if got := p.getSegmentMatchType(tt.segmentIndex); got != tt.want {
//				t.Errorf("getSegmentMatchType() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestPattern_DeterminePathSegmentMatchType(t *testing.T) {
//	type patterns struct {
//		pathPattern     string
//		caseInsensitive bool
//		urlPathSegment  string
//		segmentIndex    int
//	}
//	tests := []struct {
//		name string
//		patterns patterns
//		want MatchType
//	}{
//		{
//			name: "single path matchType:",
//			patterns: patterns{
//				pathPattern:    "/a/*/b",
//				urlPathSegment: "agdg",
//				segmentIndex:   1,
//			},
//			want: MatchTypeSingleSegment,
//		}, {
//			name: "capture var path matchType:",
//			patterns: patterns{
//				pathPattern:    "/a/{a}/b",
//				urlPathSegment: "agdg",
//				segmentIndex:   1,
//			},
//			want: MatchTypeCaptureVar,
//		},
//		{
//			name: "literal matchType: first Segment",
//			patterns: patterns{
//				pathPattern:    "/a/c/b",
//				urlPathSegment: "a",
//				segmentIndex:   0,
//			},
//			want: MatchTypeLiteral,
//		},
//		{
//			name: "literal matchType: last Segment",
//			patterns: patterns{
//				pathPattern:    "/a/c/b",
//				urlPathSegment: "b",
//				segmentIndex:   2,
//			},
//			want: MatchTypeLiteral,
//		},
//		{
//			name: "literal matchType: case-sensitive",
//			patterns: patterns{
//				pathPattern:     "/a/c/b",
//				caseInsensitive: true,
//				urlPathSegment:  "C",
//				segmentIndex:    1,
//			},
//			want: MatchTypeLiteral,
//		},
//		{
//			name: "literal matchType: case-sensitive fails",
//			patterns: patterns{
//				pathPattern:    "/a/c/b",
//				urlPathSegment: "C",
//				segmentIndex:   1,
//			},
//			want: MatchTypeUnknown,
//		},
//		{
//			name: "constraint capture var matchType: first Segment",
//			patterns: patterns{
//				pathPattern:    "/{a:\\d+}/c/b",
//				urlPathSegment: "123",
//				segmentIndex:   0,
//			},
//			want: MatchTypeConstraintCaptureVar,
//		},
//		{
//			name: "constraint capture var matchType: last Segment",
//			patterns: patterns{
//				pathPattern:    "/a/c/{b:\\d+}",
//				urlPathSegment: "012",
//				segmentIndex:   2,
//			},
//			want: MatchTypeConstraintCaptureVar,
//		},
//		{
//			name: "constraint capture var matchType: case-sensitive",
//			patterns: patterns{
//				pathPattern:     "/a/{c:[a-c]{2}}/b",
//				caseInsensitive: true,
//				urlPathSegment:  "AC",
//				segmentIndex:    1,
//			},
//			want: MatchTypeConstraintCaptureVar,
//		},
//		{
//			name: "constraint capture var matchType: case-sensitive fails",
//			patterns: patterns{
//				pathPattern:    "/a/{c:[a-c]{2}}/b",
//				urlPathSegment: "AC",
//				segmentIndex:   1,
//			},
//			want: MatchTypeUnknown,
//		},
//		{
//			name: "constraint capture var matchType: does not matchType",
//			patterns: patterns{
//				pathPattern:    "/a/{c:[a-c]{2}}/b",
//				urlPathSegment: "xy",
//				segmentIndex:   1,
//			},
//			want: MatchTypeUnknown,
//		},
//		{
//			name: "greedy matchType: simple case",
//			patterns: patterns{
//				pathPattern:    "/a/**",
//				urlPathSegment: "g",
//				segmentIndex:   1,
//			},
//			want: MatchTypeMultipleSegments,
//		},
//		{
//			name: "greedy matchType: when next Segment matches the current one",
//			patterns: patterns{
//				pathPattern:    "/a/**/g",
//				urlPathSegment: "g",
//				segmentIndex:   1,
//			},
//			want: MatchTypeMultipleSegments,
//		},
//		{
//			name: "regex matchType: simple case",
//			patterns: patterns{
//				pathPattern:    "/a/a*d/c",
//				urlPathSegment: "abcd",
//				segmentIndex:   1,
//			},
//			want: MatchTypeRegex,
//		},
//		{
//			name: "regex matchType: last Segment",
//			patterns: patterns{
//				pathPattern:    "/a/a*d",
//				urlPathSegment: "abcd",
//				segmentIndex:   1,
//			},
//			want: MatchTypeRegex,
//		},
//		{
//			name: "regex matchType: does not matchType",
//			patterns: patterns{
//				pathPattern:    "/a/a?d",
//				urlPathSegment: "abcd",
//				segmentIndex:   1,
//			},
//			want: MatchTypeUnknown,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p, err := ParsePattern(tt.patterns.pathPattern, tt.patterns.caseInsensitive)
//			if err != nil {
//				t.Fatalf("determinePathSegmentMatchType() got err: %v", err)
//			}
//			got := p.determinePathSegmentMatchType(tt.patterns.urlPathSegment, tt.patterns.segmentIndex)
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("determinePathSegmentMatchType() got: %v, want: %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_regexSegmentsMatch(t *testing.T) {
//	type patterns struct {
//		urlPathSegment  string
//		patternSegment  string
//		caseInsensitive bool
//	}
//	tests := []struct {
//		name string
//		patterns patterns
//		want bool
//	}{
//		{
//			name: "single char matchType and multiple",
//			patterns: patterns{
//				urlPathSegment:  "aabdd",
//				patternSegment:  "a?b*d",
//				caseInsensitive: false,
//			},
//			want: true,
//		},
//		{
//			name: "single char matchType and multiple char matchType",
//			patterns: patterns{
//				urlPathSegment:  "aabddcddce",
//				patternSegment:  "a?b*ddce",
//				caseInsensitive: false,
//			},
//			want: true,
//		},
//		{
//			name: "single char matchType and many consecutive *",
//			patterns: patterns{
//				urlPathSegment:  "aabddcddce",
//				patternSegment:  "a?b*****ddce",
//				caseInsensitive: false,
//			},
//			want: true,
//		},
//		{
//			name: "multiple *",
//			patterns: patterns{
//				urlPathSegment:  "aabwwqq",
//				patternSegment:  "a?b*w*q",
//				caseInsensitive: false,
//			},
//			want: true,
//		},
//		{
//			name: "end with *",
//			patterns: patterns{
//				urlPathSegment:  "aabwwqq",
//				patternSegment:  "a?b*",
//				caseInsensitive: false,
//			},
//			want: true,
//		},
//		{
//			name: "starts with *",
//			patterns: patterns{
//				urlPathSegment:  "awwq",
//				patternSegment:  "*w?q",
//				caseInsensitive: false,
//			},
//			want: true,
//		},
//		{
//			name: "single char matchType and multiple char matchType returns false",
//			patterns: patterns{
//				urlPathSegment:  "aabddcddc",
//				patternSegment:  "a?b*ddce",
//				caseInsensitive: false,
//			},
//			want: false,
//		},
//		{
//			name: "non greedy matchType: same length returns true",
//			patterns: patterns{
//				urlPathSegment:  "aabddcddc",
//				patternSegment:  "a??d?cdd?",
//				caseInsensitive: false,
//			},
//			want: true,
//		}, {
//			name: "non greedy matchType: same length returns false",
//			patterns: patterns{
//				urlPathSegment:  "aabddcdcc",
//				patternSegment:  "a??d?dd?",
//				caseInsensitive: false,
//			},
//			want: false,
//		}, {
//			name: "non greedy matchType: different length",
//			patterns: patterns{
//				urlPathSegment:  "aabddcdc",
//				patternSegment:  "a??d?dd?",
//				caseInsensitive: false,
//			},
//			want: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := regexSegmentMatch(tt.patterns.urlPathSegment, tt.patterns.patternSegment, tt.patterns.caseInsensitive); got != tt.want {
//				t.Errorf("regexSegmentMatch() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
