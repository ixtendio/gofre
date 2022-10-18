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

func Test_trieNode_addCaptureVarNameIfNotExists(t *testing.T) {
	type fields struct {
		captureVarNames []string
	}
	tests := []struct {
		name   string
		fields fields
		args   string
		want   []string
	}{
		{
			name:   "captureVarNames is empty",
			fields: fields{},
			args:   "a",
			want:   []string{"a"},
		},
		{
			name:   "captureVarNames is not empty",
			fields: fields{captureVarNames: []string{"a", "b"}},
			args:   "c",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "var name is empty",
			fields: fields{captureVarNames: []string{"a"}},
			args:   "",
			want:   []string{"a"},
		},
		{
			name:   "var name already exists but is uppercase",
			fields: fields{captureVarNames: []string{"A"}},
			args:   "a",
			want:   []string{"A", "a"},
		},
		{
			name:   "var name already exists",
			fields: fields{captureVarNames: []string{"a"}},
			args:   "a",
			want:   []string{"a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &trieNode{
				captureVarNames: tt.fields.captureVarNames,
			}
			n.addCaptureVarNameIfNotExists(tt.args)

			if !reflect.DeepEqual(n.captureVarNames, tt.want) {
				t.Errorf("addCaptureVarNameIfNotExists() got = '%v', want = '%v'", n.captureVarNames, tt.want)
			}
		})
	}
}

func Test_trieNode_sortChildren(t *testing.T) {
	nilHandler := func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return nil, nil
	}
	type fields struct {
		children []*trieNode
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "empty children",
			fields: fields{},
			want:   nil,
		},
		{
			name: "sort 2 leafs",
			fields: fields{
				children: []*trieNode{
					{
						pathElement: &path.Element{PathPatternId: "1"},
						handler:     nilHandler,
					}, {
						pathElement: &path.Element{PathPatternId: "2"},
						handler:     nilHandler,
					}},
			},
			want: []string{"1", "2"},
		},
		{
			name: "sort 1 leaf and a node - uc1",
			fields: fields{
				children: []*trieNode{
					{
						pathElement: &path.Element{PathPatternId: "1"},
					}, {
						pathElement: &path.Element{PathPatternId: "2"},
						handler:     nilHandler,
					}},
			},
			want: []string{"2", "1"},
		},
		{
			name: "sort 1 leaf and a node - uc2",
			fields: fields{
				children: []*trieNode{
					{
						pathElement: &path.Element{PathPatternId: "1"},
						handler:     nilHandler,
					}, {
						pathElement: &path.Element{PathPatternId: "2"},
					}},
			},
			want: []string{"1", "2"},
		},
		{
			name: "sort 2 leafs and a node",
			fields: fields{
				children: []*trieNode{
					{
						pathElement: &path.Element{PathPatternId: "1"},
					},
					{
						pathElement: &path.Element{PathPatternId: "2"},
						handler:     nilHandler,
					},
					{
						pathElement: &path.Element{PathPatternId: "3"},
						handler:     nilHandler,
					}},
			},
			want: []string{"2", "3", "1"},
		},
		{
			name: "sort 2 nodes with the same priority",
			fields: fields{
				children: []*trieNode{
					{
						pathElement: &path.Element{PathPatternId: "1", MatchType: path.MatchLiteralType},
					},
					{
						pathElement: &path.Element{PathPatternId: "2", MatchType: path.MatchLiteralType},
					}},
			},
			want: []string{"1", "2"},
		},
		{
			name: "sort nodes with leaf leafs",
			fields: fields{
				children: []*trieNode{
					{
						pathElement: &path.Element{PathPatternId: "60", MatchType: path.MatchLiteralType},
					},
					{
						pathElement: &path.Element{PathPatternId: "70", MatchType: path.MatchSeparatorType},
					},
					{
						pathElement: &path.Element{PathPatternId: "50", MatchType: path.MatchMultiplePathsType},
					},
					{
						pathElement: &path.Element{PathPatternId: "40", MatchType: path.MatchRegexType},
					},
					{
						pathElement: &path.Element{PathPatternId: "20", MatchType: path.MatchVarCaptureWithConstraintType},
					},
					{
						pathElement: &path.Element{PathPatternId: "10", MatchType: path.MatchVarCaptureType},
					},
					{
						pathElement: &path.Element{PathPatternId: "30"},
						handler:     nilHandler,
					},
				},
			},
			want: []string{"30", "70", "60", "20", "10", "40", "50"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &trieNode{
				children: tt.fields.children,
			}
			n.sortChildren()
			var got []string
			for _, c := range n.children {
				got = append(got, c.pathElement.PathPatternId)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addCaptureVarNameIfNotExists() got = '%v', want = '%v'", got, tt.want)
			}
		})
	}
}

func Test_trieNode_addChild(t *testing.T) {
	nilHandler := func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return nil, nil
	}
	type parentNode struct {
		children []*trieNode
		handler  handler2.Handler
	}
	type want struct {
		returnNodeId              string
		returnNodeCaptureVarNames []string
		parentChildrenId          []string
	}
	tests := []struct {
		name       string
		parentNode parentNode
		args       *path.Element
		want       *want
		wantErr    bool
	}{
		{
			name: "MatchVarCaptureWithConstraintType - add existing child with the same match pattern",
			parentNode: parentNode{children: []*trieNode{{
				captureVarNames: []string{"1/a"},
				pathElement: &path.Element{
					PathPatternId:  "1",
					MatchType:      path.MatchVarCaptureWithConstraintType,
					MatchPattern:   "a.*b",
					CaptureVarName: "b",
				}},
			}},
			args: &path.Element{
				PathPatternId:  "2",
				MatchType:      path.MatchVarCaptureWithConstraintType,
				MatchPattern:   "a.*b",
				CaptureVarName: "b",
			},
			want: &want{
				returnNodeId:              "1",
				parentChildrenId:          []string{"1"},
				returnNodeCaptureVarNames: []string{"1/a", "2/b"},
			},
			wantErr: false,
		},
		{
			name: "MatchVarCaptureWithConstraintType - add existing child with different match pattern",
			parentNode: parentNode{children: []*trieNode{{
				captureVarNames: []string{"1/a"},
				pathElement: &path.Element{
					PathPatternId:  "1",
					MatchType:      path.MatchVarCaptureWithConstraintType,
					MatchPattern:   "a.*b",
					CaptureVarName: "b",
				}},
			}},
			args: &path.Element{
				PathPatternId:  "2",
				MatchType:      path.MatchVarCaptureWithConstraintType,
				MatchPattern:   "a.*c",
				CaptureVarName: "b",
			},
			want: &want{
				returnNodeId:              "2",
				parentChildrenId:          []string{"1", "2"},
				returnNodeCaptureVarNames: []string{"2/b"},
			},
			wantErr: false,
		},
		{
			name: "MatchVarCaptureType - add existing child with the same capture var name",
			parentNode: parentNode{children: []*trieNode{{
				pathElement: &path.Element{
					PathPatternId:  "1",
					MatchType:      path.MatchVarCaptureType,
					CaptureVarName: "b",
				}},
			}},
			args: &path.Element{
				PathPatternId:  "2",
				MatchType:      path.MatchVarCaptureType,
				CaptureVarName: "b",
			},
			want: &want{
				returnNodeId:              "1",
				parentChildrenId:          []string{"1"},
				returnNodeCaptureVarNames: []string{"2/b"},
			},
			wantErr: false,
		},
		{
			name: "MatchVarCaptureType - add existing child with the different capture var name",
			parentNode: parentNode{children: []*trieNode{{
				captureVarNames: []string{"1/a"},
				pathElement: &path.Element{
					PathPatternId:  "1",
					MatchType:      path.MatchVarCaptureType,
					CaptureVarName: "a",
				}},
			}},
			args: &path.Element{
				PathPatternId:  "2",
				MatchType:      path.MatchVarCaptureType,
				CaptureVarName: "b",
			},
			want: &want{
				returnNodeId:              "1",
				parentChildrenId:          []string{"1"},
				returnNodeCaptureVarNames: []string{"1/a", "2/b"},
			},
			wantErr: false,
		},
		{
			name: "MatchRegexType - add existing child with the same raw value",
			parentNode: parentNode{children: []*trieNode{{
				pathElement: &path.Element{
					PathPatternId: "1",
					MatchType:     path.MatchRegexType,
					MatchPattern:  "a.*b",
				}},
			}},
			args: &path.Element{
				PathPatternId: "2",
				MatchType:     path.MatchRegexType,
				MatchPattern:  "a.*b",
			},
			want: &want{
				returnNodeId:     "1",
				parentChildrenId: []string{"1"},
			},
			wantErr: false,
		},
		{
			name: "MatchRegexType - add existing child with different raw value",
			parentNode: parentNode{children: []*trieNode{{
				pathElement: &path.Element{
					PathPatternId: "1",
					MatchType:     path.MatchRegexType,
					MatchPattern:  "a.*b",
				}},
			}},
			args: &path.Element{
				PathPatternId: "2",
				MatchType:     path.MatchRegexType,
				MatchPattern:  "a.*c",
			},
			want: &want{
				returnNodeId:     "2",
				parentChildrenId: []string{"1", "2"},
			},
			wantErr: false,
		},
		{
			name: "MatchLiteralType - add existing child with the same raw value",
			parentNode: parentNode{children: []*trieNode{{
				pathElement: &path.Element{
					PathPatternId: "1",
					MatchType:     path.MatchLiteralType,
					RawVal:        "a",
				}},
			}},
			args: &path.Element{
				PathPatternId: "2",
				MatchType:     path.MatchLiteralType,
				RawVal:        "a",
			},
			want: &want{
				returnNodeId:     "1",
				parentChildrenId: []string{"1"},
			},
			wantErr: false,
		},
		{
			name: "MatchLiteralType - add existing child with different raw value",
			parentNode: parentNode{children: []*trieNode{{
				pathElement: &path.Element{
					PathPatternId: "1",
					MatchType:     path.MatchLiteralType,
					RawVal:        "a",
				}},
			}},
			args: &path.Element{
				PathPatternId: "2",
				MatchType:     path.MatchLiteralType,
				RawVal:        "b",
			},
			want: &want{
				returnNodeId:     "2",
				parentChildrenId: []string{"1", "2"},
			},
			wantErr: false,
		},
		{
			name: "MatchSeparatorType - add existing child",
			parentNode: parentNode{children: []*trieNode{{
				pathElement: &path.Element{
					PathPatternId: "1",
					MatchType:     path.MatchSeparatorType}},
			}},
			args: &path.Element{
				PathPatternId: "2",
				MatchType:     path.MatchSeparatorType},
			want: &want{
				returnNodeId:     "1",
				parentChildrenId: []string{"1"},
			},
			wantErr: false,
		},
		{
			name:       "add child to an empty list",
			parentNode: parentNode{},
			args: &path.Element{
				PathPatternId: "1",
				MatchType:     path.MatchSeparatorType},
			want: &want{
				returnNodeId:     "1",
				parentChildrenId: []string{"1"},
			},
			wantErr: false,
		},
		{
			name:       "add child to a leaf",
			parentNode: parentNode{handler: nilHandler},
			args:       &path.Element{},
			want:       nil,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &trieNode{
				children: tt.parentNode.children,
				handler:  tt.parentNode.handler,
			}
			got, err := parent.addChild(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("addChild() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				var gotChildren []string
				for _, c := range parent.children {
					gotChildren = append(gotChildren, c.pathElement.PathPatternId)
				}
				if !reflect.DeepEqual(gotChildren, tt.want.parentChildrenId) {
					t.Fatalf("addChild() got children = '%v', want children = '%v'", gotChildren, tt.want.parentChildrenId)
				}
				if got.pathElement.PathPatternId != tt.want.returnNodeId {
					t.Fatalf("addChild() got nodeId = '%s', want nodeId = '%s'", got.pathElement.PathPatternId, tt.want.returnNodeId)
				}
				if !reflect.DeepEqual(got.captureVarNames, tt.want.returnNodeCaptureVarNames) {
					t.Fatalf("addChild() got captureVarNames = '%v', want captureVarNames = '%v'", got.captureVarNames, tt.want.returnNodeCaptureVarNames)
				}
			}
		})
	}
}

func Test_trieNode_addLeaf(t *testing.T) {
	nilHandler := func(ctx context.Context, r *request.HttpRequest) (response.HttpResponse, error) {
		return nil, nil
	}
	type parentNodeFields struct {
		children []*trieNode
		handler  handler2.Handler
	}
	tests := []struct {
		name              string
		parentNode        parentNodeFields
		wantChildrenCount int
		wantErr           bool
	}{
		{
			name:              "add leaf when a leaf already exists",
			parentNode:        parentNodeFields{children: []*trieNode{{handler: nilHandler}}},
			wantChildrenCount: 1,
			wantErr:           true,
		},
		{
			name:              "add leaf to empty children",
			parentNode:        parentNodeFields{},
			wantChildrenCount: 1,
			wantErr:           false,
		},
		{
			name:              "add leaf to when parent is leaf",
			parentNode:        parentNodeFields{handler: nilHandler},
			wantChildrenCount: 0,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &trieNode{
				pathElement: &path.Element{RawVal: "a"},
				children:    tt.parentNode.children,
				handler:     tt.parentNode.handler,
			}

			err := parent.addLeaf(nilHandler)
			if (err != nil) != tt.wantErr {
				t.Fatalf("addLeaf() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(parent.children) != tt.wantChildrenCount {
				t.Fatalf("addLeaf() got children = '%v', want children = '%v'", len(parent.children), tt.wantChildrenCount)
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
