package path

import (
	"reflect"
	"regexp"
	"testing"
)

func Test_ParsePattern(t *testing.T) {
	type want struct {
		rawValue             string
		caseInsensitive      bool
		captureVarsLen       uint8
		maxMatchableSegments uint8
		priority             uint64
		segments             []segment
	}
	type patterns struct {
		pathPattern     string
		caseInsensitive bool
	}
	tests := []struct {
		name     string
		patterns patterns
		want     want
		wantErr  bool
	}{
		{
			name:     "path only with greedy matchType",
			patterns: patterns{pathPattern: "/**/**"},
			want:     want{},
			wantErr:  true,
		},
		{
			name:     "path with consecutive greedy matchType segments",
			patterns: patterns{pathPattern: "/**/a/**/**/b"},
			want:     want{},
			wantErr:  true,
		},
		{
			name:     "path that not starts with slash",
			patterns: patterns{pathPattern: "abc/"},
			want:     want{},
			wantErr:  true,
		},
		{
			name:     "path with capture variable without regex",
			patterns: patterns{pathPattern: "/abc/{id}"},
			want: want{
				rawValue:             "/abc/{id}",
				caseInsensitive:      false,
				captureVarsLen:       1,
				maxMatchableSegments: 2,
				priority:             1300000000000000000,
				segments: []segment{{
					val:       "abc",
					matchType: 1,
				}, {
					val:            "{id}",
					matchType:      3,
					captureVarName: "id",
				}},
			},
			wantErr: false,
		},
		{
			name:     "path with capture variable and regex",
			patterns: patterns{pathPattern: "/abc/{id:\\d}"},
			want: want{
				rawValue:             "/abc/{id:\\d}",
				caseInsensitive:      false,
				captureVarsLen:       1,
				maxMatchableSegments: 2,
				priority:             1200000000000000000,
				segments: []segment{{
					val:       "abc",
					matchType: 1,
				}, {
					val:            "{id:\\d}",
					matchType:      2,
					captureVarName: "id",
				}},
			},
			wantErr: false,
		},
		{
			name:     "root path",
			patterns: patterns{pathPattern: "/"},
			want: want{
				rawValue:             "/",
				caseInsensitive:      false,
				captureVarsLen:       0,
				maxMatchableSegments: 0,
				priority:             0,
			},
			wantErr: false,
		},
		{
			name:     "root path with double slash",
			patterns: patterns{pathPattern: "//"},
			want:     want{},
			wantErr:  true,
		},
		{
			name:     "path with many slash",
			patterns: patterns{pathPattern: "/abc///cde////"},
			want:     want{},
			wantErr:  true,
		},
		{
			name:     "path with single asterix",
			patterns: patterns{pathPattern: "/a/*"},
			want: want{
				rawValue:             "/a/*",
				caseInsensitive:      false,
				captureVarsLen:       0,
				maxMatchableSegments: 2,
				priority:             1500000000000000000,
				segments: []segment{{
					val:       "a",
					matchType: 1,
				}, {
					val:       "*",
					matchType: 5,
				}},
			},
			wantErr: false,
		}, {
			name:     "path with max segments",
			patterns: patterns{pathPattern: "/a/*/b/{q}/{y:[a-z]+}/?as*/d/e/f/g/{t}/i/j/k/l/m/n/o/{w:[a-z]+}"},
			want: want{
				rawValue:             "/a/*/b/{q}/{y:[a-z]+}/?as*/d/e/f/g/{t}/i/j/k/l/m/n/o/{w:[a-z]+}",
				caseInsensitive:      false,
				captureVarsLen:       4,
				maxMatchableSegments: 19,
				priority:             1513241111311111112,
				segments: []segment{
					{val: "a", matchType: 1},
					{val: "*", matchType: 5},
					{val: "b", matchType: 1},
					{val: "{q}", matchType: 3, captureVarName: "q"},
					{val: "{y:[a-z]+}", matchType: 2, captureVarName: "y"},
					{val: "?as*", matchType: 4},
					{val: "d", matchType: 1},
					{val: "e", matchType: 1},
					{val: "f", matchType: 1},
					{val: "g", matchType: 1},
					{val: "{t}", matchType: 3, captureVarName: "t"},
					{val: "i", matchType: 1},
					{val: "j", matchType: 1},
					{val: "k", matchType: 1},
					{val: "l", matchType: 1},
					{val: "m", matchType: 1},
					{val: "n", matchType: 1},
					{val: "o", matchType: 1},
					{val: "{w:[a-z]+}", matchType: 2, captureVarName: "w"},
				},
			},
			wantErr: false,
		},
		{
			name:     "path with double asterix at start",
			patterns: patterns{pathPattern: "/**/a"},
			want: want{
				rawValue:             "/**/a",
				caseInsensitive:      false,
				captureVarsLen:       0,
				maxMatchableSegments: 255,
				priority:             6666666666666666661,
				segments: []segment{
					{val: "**", matchType: 6},
					{val: "a", matchType: 1},
				},
			},
			wantErr: false,
		},
		{
			name:     "path with double asterix at the end",
			patterns: patterns{pathPattern: "/a/**"},
			want: want{
				rawValue:             "/a/**",
				caseInsensitive:      false,
				captureVarsLen:       0,
				maxMatchableSegments: 255,
				priority:             1666666666666666666,
				segments: []segment{
					{val: "a", matchType: 1},
					{val: "**", matchType: 6},
				},
			},
			wantErr: false,
		},
		{
			name:     "path with double asterix in the middle",
			patterns: patterns{pathPattern: "/a/**/b"},
			want: want{
				rawValue:             "/a/**/b",
				caseInsensitive:      false,
				captureVarsLen:       0,
				maxMatchableSegments: 255,
				priority:             1666666666666666661,
				segments: []segment{
					{val: "a", matchType: 1},
					{val: "**", matchType: 6},
					{val: "b", matchType: 1},
				},
			},
			wantErr: false,
		},
		{
			name:     "path with multiple double asterix segments",
			patterns: patterns{pathPattern: "/a/**/b/**/c/**/d/**/e/**/f/g/h"},
			want: want{
				rawValue:             "/a/**/b/**/c/**/d/**/e/**/f/g/h",
				caseInsensitive:      false,
				captureVarsLen:       0,
				maxMatchableSegments: 255,
				priority:             1666166166166166111,
				segments: []segment{
					{val: "a", matchType: 1},
					{val: "**", matchType: 6},
					{val: "b", matchType: 1},
					{val: "**", matchType: 6},
					{val: "c", matchType: 1},
					{val: "**", matchType: 6},
					{val: "d", matchType: 1},
					{val: "**", matchType: 6},
					{val: "e", matchType: 1},
					{val: "**", matchType: 6},
					{val: "f", matchType: 1},
					{val: "g", matchType: 1},
					{val: "h", matchType: 1},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePattern(tt.patterns.pathPattern, tt.patterns.caseInsensitive)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePattern() got nil error, wantErr %v", tt.wantErr)
				}
			} else {
				var segments []segment
				if p.segments != nil {
					for _, s := range p.segments {
						segments = append(segments, segment{
							val:            s.val,
							matchType:      s.matchType,
							captureVarName: s.captureVarName,
						})
					}
				}
				got := want{
					rawValue:             p.RawValue,
					caseInsensitive:      p.caseInsensitive,
					captureVarsLen:       p.captureVarsLen,
					maxMatchableSegments: p.maxMatchableSegments,
					priority:             p.priority,
					segments:             segments,
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParsePattern() \ngot: %v, \nwant: %v", got, tt.want)
				}
			}
		})
	}
}

func Test_validatePathSegment(t *testing.T) {
	tests := []struct {
		name        string
		pathSegment string
		wantErr     bool
	}{
		{
			name:        "empty pattern",
			pathSegment: "",
			wantErr:     true,
		},
		{
			name:        "literal pattern",
			pathSegment: "a",
			wantErr:     false,
		},
		{
			name:        "capture var pattern without open bracket",
			pathSegment: "a}",
			wantErr:     true,
		},
		{
			name:        "capture var pattern without closed bracket",
			pathSegment: "{c",
			wantErr:     true,
		},
		{
			name:        "capture var pattern without name",
			pathSegment: "{}",
			wantErr:     true,
		},
		{
			name:        "capture var pattern with constraint but without name",
			pathSegment: "{:[a-z]+}",
			wantErr:     true,
		},
		{
			name:        "capture var pattern with constraint regex",
			pathSegment: "{a:}",
			wantErr:     true,
		},
		{
			name:        "capture var pattern with constraint regex and without name",
			pathSegment: "{:}",
			wantErr:     true,
		},
		{
			name:        "path segment with triple asterix",
			pathSegment: "***",
			wantErr:     true,
		},
		{
			name:        "path segment with double asterix and another text at start",
			pathSegment: "**abc",
			wantErr:     true,
		},
		{
			name:        "path segment with double asterix and another text",
			pathSegment: "abc**def",
			wantErr:     true,
		},
		{
			name:        "path segment with double asterix and another text at the end",
			pathSegment: "bla**",
			wantErr:     true,
		},
		{
			name:        "valid capture var pattern without constraint",
			pathSegment: "{abc}",
			wantErr:     false,
		},
		{
			name:        "valid capture var pattern with constraint",
			pathSegment: "{abc:[a-z]+}",
			wantErr:     false,
		},
		{
			name:        "valid capture var pattern with constraint and nested brackets",
			pathSegment: "{abc:[a-z]{3}}",
			wantErr:     false,
		},
		{
			name:        "valid path segment with regex ?",
			pathSegment: "?asd",
			wantErr:     false,
		},
		{
			name:        "valid path segment with regex * at beginning",
			pathSegment: "*asd",
			wantErr:     false,
		},
		{
			name:        "valid path segment with regex *",
			pathSegment: "a*sd",
			wantErr:     false,
		}, {
			name:        "valid path segment with regex * at the end",
			pathSegment: "asd*",
			wantErr:     false,
		},
		{
			name:        "valid path segment with one asterix",
			pathSegment: "*",
			wantErr:     false,
		},
		{
			name:        "valid path segment with double asterix",
			pathSegment: "**",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePathSegment(tt.pathSegment); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_determineMatchTypeForSegment(t *testing.T) {
	tests := []struct {
		name        string
		pathSegment string
		want        MatchType
	}{
		{
			name:        "MatchTypeSingleSegment",
			pathSegment: "*",
			want:        MatchTypeSingleSegment,
		},
		{
			name:        "MatchTypeMultipleSegments",
			pathSegment: "**",
			want:        MatchTypeMultipleSegments,
		},
		{
			name:        "MatchTypeConstraintCaptureVar",
			pathSegment: "{abc:[a-z]+}",
			want:        MatchTypeConstraintCaptureVar,
		},
		{
			name:        "MatchTypeCaptureVar",
			pathSegment: "{abc}",
			want:        MatchTypeCaptureVar,
		},
		{
			name:        "MatchTypeRegex ?",
			pathSegment: "abc?asd",
			want:        MatchTypeRegex,
		},
		{
			name:        "MatchTypeRegex *",
			pathSegment: "abc*asd",
			want:        MatchTypeRegex,
		},
		{
			name:        "MatchTypeLiteral",
			pathSegment: "abcasd",
			want:        MatchTypeLiteral,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineMatchTypeForSegment(tt.pathSegment); got != tt.want {
				t.Errorf("determineMatchTypeForSegment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPattern_matchUrlPathSegment(t *testing.T) {
	regex, _ := regexp.Compile("[a-c]{4}")
	type given struct {
		val               string
		matchType         MatchType
		captureVarPattern *regexp.Regexp
	}
	type args struct {
		urlPath         string
		urlSegment      UrlSegment
		caseInsensitive bool
	}
	tests := []struct {
		name  string
		given given
		args  args
		want  MatchType
	}{
		{
			name: "MatchTypeLiteral case-insensitive OK",
			given: given{
				val:       "a",
				matchType: MatchTypeLiteral,
			},
			args: args{
				urlPath: "/a/b/c",
				urlSegment: UrlSegment{
					startIndex: 1,
					endIndex:   2,
				},
				caseInsensitive: false,
			},
			want: MatchTypeLiteral,
		},
		{
			name: "MatchTypeLiteral case-insensitive should FAIL",
			given: given{
				val:       "A",
				matchType: MatchTypeLiteral,
			},
			args: args{
				urlPath: "/a/b/c",
				urlSegment: UrlSegment{
					startIndex: 1,
					endIndex:   2,
				},
				caseInsensitive: false,
			},
			want: MatchTypeUnknown,
		},
		{
			name: "MatchTypeLiteral case-sensitive OK",
			given: given{
				val:       "A",
				matchType: MatchTypeLiteral,
			},
			args: args{
				urlPath: "/a/b/c",
				urlSegment: UrlSegment{
					startIndex: 1,
					endIndex:   2,
				},
				caseInsensitive: true,
			},
			want: MatchTypeLiteral,
		},
		{
			name: "MatchTypeSingleSegment",
			given: given{
				matchType: 5,
			},
			want: MatchTypeSingleSegment,
		},
		{
			name: "MatchTypeCaptureVar",
			given: given{
				matchType: MatchTypeCaptureVar,
			},
			want: MatchTypeCaptureVar,
		},
		{
			name: "MatchTypeMultipleSegments",
			given: given{
				matchType: MatchTypeMultipleSegments,
			},
			want: MatchTypeMultipleSegments,
		},
		{
			name: "MatchTypeConstraintCaptureVar",
			given: given{
				val:               "a",
				matchType:         MatchTypeConstraintCaptureVar,
				captureVarPattern: regex,
			},
			args: args{
				urlPath: "abcca/b/c",
				urlSegment: UrlSegment{
					startIndex: 1,
					endIndex:   5,
				},
				caseInsensitive: false,
			},
			want: MatchTypeConstraintCaptureVar,
		},
		{
			name: "MatchTypeRegex case-sensitive",
			given: given{
				val:       "a?c*fg",
				matchType: MatchTypeRegex,
			},
			args: args{
				urlPath: "/abcdefg/b",
				urlSegment: UrlSegment{
					startIndex: 1,
					endIndex:   8,
				},
				caseInsensitive: false,
			},
			want: MatchTypeRegex,
		}, {
			name: "MatchTypeRegex case-sensitive FAILS",
			given: given{
				val:       "a?c*fG",
				matchType: MatchTypeRegex,
			},
			args: args{
				urlPath: "/abcdefg/b",
				urlSegment: UrlSegment{
					startIndex: 1,
					endIndex:   8,
				},
				caseInsensitive: false,
			},
			want: MatchTypeUnknown,
		}, {
			name: "MatchTypeRegex case-insensitive",
			given: given{
				val:       "a?c*fG",
				matchType: MatchTypeRegex,
			},
			args: args{
				urlPath: "/abcdefg/b",
				urlSegment: UrlSegment{
					startIndex: 1,
					endIndex:   8,
				},
				caseInsensitive: true,
			},
			want: MatchTypeRegex,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := segment{
				val:               tt.given.val,
				matchType:         tt.given.matchType,
				captureVarPattern: tt.given.captureVarPattern,
			}
			got := s.matchUrlPathSegment(tt.args.urlPath, &tt.args.urlSegment, tt.args.caseInsensitive)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchUrlPathSegment() got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func Test_regexSegmentsMatch(t *testing.T) {
	type patterns struct {
		urlPathSegment  string
		patternSegment  string
		caseInsensitive bool
	}
	tests := []struct {
		name     string
		patterns patterns
		want     bool
	}{
		{
			name: "single char matchType and multiple",
			patterns: patterns{
				urlPathSegment:  "aabdd",
				patternSegment:  "a?b*d",
				caseInsensitive: false,
			},
			want: true,
		},
		{
			name: "single char matchType and multiple char matchType",
			patterns: patterns{
				urlPathSegment:  "aabddcddce",
				patternSegment:  "a?b*ddce",
				caseInsensitive: false,
			},
			want: true,
		},
		{
			name: "single char matchType and many consecutive *",
			patterns: patterns{
				urlPathSegment:  "aabddcddce",
				patternSegment:  "a?b*****ddce",
				caseInsensitive: false,
			},
			want: true,
		},
		{
			name: "multiple *",
			patterns: patterns{
				urlPathSegment:  "aabwwqq",
				patternSegment:  "a?b*w*q",
				caseInsensitive: false,
			},
			want: true,
		},
		{
			name: "end with *",
			patterns: patterns{
				urlPathSegment:  "aabwwqq",
				patternSegment:  "a?b*",
				caseInsensitive: false,
			},
			want: true,
		},
		{
			name: "starts with *",
			patterns: patterns{
				urlPathSegment:  "awwq",
				patternSegment:  "*w?q",
				caseInsensitive: false,
			},
			want: true,
		},
		{
			name: "single char matchType and multiple char matchType returns false",
			patterns: patterns{
				urlPathSegment:  "aabddcddc",
				patternSegment:  "a?b*ddce",
				caseInsensitive: false,
			},
			want: false,
		},
		{
			name: "non greedy matchType: same length returns true",
			patterns: patterns{
				urlPathSegment:  "aabddcddc",
				patternSegment:  "a??d?cdd?",
				caseInsensitive: false,
			},
			want: true,
		}, {
			name: "non greedy matchType: same length returns false",
			patterns: patterns{
				urlPathSegment:  "aabddcdcc",
				patternSegment:  "a??d?dd?",
				caseInsensitive: false,
			},
			want: false,
		}, {
			name: "non greedy matchType: different length",
			patterns: patterns{
				urlPathSegment:  "aabddcdc",
				patternSegment:  "a??d?dd?",
				caseInsensitive: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := regexSegmentMatch(tt.patterns.urlPathSegment, tt.patterns.patternSegment, tt.patterns.caseInsensitive); got != tt.want {
				t.Errorf("regexSegmentMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
