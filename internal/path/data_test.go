package path

import (
	"testing"
)

func Test_nonCaptureVarElement(t *testing.T) {
	type args struct {
		val             string
		caseInsensitive bool
	}
	type want struct {
		matchingType     int
		maxMatchElements int
		value            string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "matching_type-match_literal_type",
			args: args{
				val:             "text",
				caseInsensitive: false,
			},
			want: want{
				matchingType:     MatchLiteralType,
				maxMatchElements: 1,
				value:            "text",
			},
			wantErr: false,
		},
		{
			name: "matching_type-match_regex_type_(?)",
			args: args{
				val:             "te?t",
				caseInsensitive: false,
			},
			want: want{
				matchingType:     MatchRegexType,
				maxMatchElements: 1,
				value:            "te?t",
			},
			wantErr: false,
		},
		{
			name: "matching_type-match_regex_type_(*)",
			args: args{
				val:             "te*t",
				caseInsensitive: false,
			},
			want: want{
				matchingType:     MatchRegexType,
				maxMatchElements: 1,
				value:            "te*t",
			},
			wantErr: false,
		},
		{
			name: "matching_type_match-multiple_paths_type_(**)",
			args: args{
				val:             "**",
				caseInsensitive: false,
			},
			want: want{
				matchingType:     MatchMultiplePathsType,
				maxMatchElements: -1,
				value:            "**",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nonCaptureVarElement("12345", tt.args.val, tt.args.caseInsensitive)
			if (err != nil) != tt.wantErr {
				t.Errorf("nonCaptureVarElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.MatchType != tt.want.matchingType {
				t.Errorf("nonCaptureVarElement() got MatchType = %v, want %v", got.MatchType, tt.want.matchingType)
			}
			if got.MaxMatchElements != tt.want.maxMatchElements {
				t.Errorf("nonCaptureVarElement() got MaxMatchElements = %v, want %v", got.MaxMatchElements, tt.want.maxMatchElements)
			}
			if got.RawVal != tt.want.value {
				t.Errorf("nonCaptureVarElement() got RawVal = %v, want %v", got.RawVal, tt.want.value)
			}
		})
	}
}

func Test_nonCaptureVarElement_matcherFunc(t *testing.T) {
	type args struct {
		val             string
		caseInsensitive bool
	}
	tests := []struct {
		name string
		args args
		want map[string]bool
	}{
		{
			name: "literal_case_insensitive",
			args: args{
				val:             "text",
				caseInsensitive: true,
			},
			want: map[string]bool{"text": true, "TeXt": true},
		},
		{
			name: "literal_case_sensitive",
			args: args{
				val:             "text",
				caseInsensitive: false,
			},
			want: map[string]bool{"text": true, "TeXt": false},
		},
		{
			name: "literal_match_multiple_paths_type",
			args: args{
				val:             "**",
				caseInsensitive: false,
			},
			want: map[string]bool{"a": true, "Abc": true},
		},
		{
			name: "literal_match_regex_type_case_sensitive",
			args: args{
				val:             "abc*hij",
				caseInsensitive: false,
			},
			want: map[string]bool{"abcdefghij": true, "abchij": true, "abcij": false, "Abchij": false},
		},
		{
			name: "literal_match_regex_type_case_insensitive",
			args: args{
				val:             "abc*hij",
				caseInsensitive: true,
			},
			want: map[string]bool{"abcdefghij": true, "abchij": true, "abcij": false, "Abchij": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nonCaptureVarElement("12345", tt.args.val, tt.args.caseInsensitive)
			if err != nil {
				t.Errorf("nonCaptureVarElement() error = %v", err)
				return
			}
			mc := &MatchingContext{}
			for key := range tt.want {
				mc.PathElements = append(mc.PathElements, key)
			}
			for i := 0; i < len(mc.PathElements); i++ {
				key := mc.PathElements[i]
				expectedResult := tt.want[key]
				gotResult, _ := got.MatchPathSegment(key)
				if gotResult != expectedResult {
					t.Errorf("nonCaptureVarElement().MatchPathSegment(%d, '%s') got = %v, want = %v", i, key, gotResult, expectedResult)
				}
			}
		})
	}
}

func Test_captureVarElement_matcherFunc(t *testing.T) {
	type args struct {
		val             string
		caseInsensitive bool
	}
	tests := []struct {
		name string
		args args
		want map[string]bool
	}{
		{
			name: "no constraints",
			args: args{
				val:             "{text}",
				caseInsensitive: true,
			},
			want: map[string]bool{"text": true, "TeXt": true},
		},
		{
			name: "UUID constraints",
			args: args{
				val:             "{text:^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$}",
				caseInsensitive: true,
			},
			want: map[string]bool{"zyw3040f-0f1c-4e98-b71c-d3cd61213f90": false, "fbd3040f-0f1c-4e98-b71c-d3cd61213f90": true},
		},
		{
			name: "digit constraints",
			args: args{
				val:             "{text:^\\d.*$}",
				caseInsensitive: true,
			},
			want: map[string]bool{"a12345": false, "09836": true},
		},
		{
			name: "3 digit constraints",
			args: args{
				val:             "{text:^[0-9]{3}$}",
				caseInsensitive: true,
			},
			want: map[string]bool{"1234": false, "123": true, "12": false, "1": false, "": false},
		},
		{
			name: "min 1 and max 3 digit constraints",
			args: args{
				val:             "{text:^[0-9]{1,3}$}",
				caseInsensitive: true,
			},
			want: map[string]bool{"1234": false, "123": true, "12": true, "1": true, "": false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := captureVarElement("12345", tt.args.val, tt.args.caseInsensitive)
			if err != nil {
				t.Errorf("captureVarElement() error = %v", err)
				return
			}
			mc := &MatchingContext{}
			for key := range tt.want {
				mc.PathElements = append(mc.PathElements, key)
			}
			for i := 0; i < len(mc.PathElements); i++ {
				key := mc.PathElements[i]
				expectedResult := tt.want[key]
				gotResult, _ := got.MatchPathSegment(key)
				if gotResult != expectedResult {
					t.Errorf("captureVarElement().MatchPathSegment(%d, '%s') got = %v, want = %v", i, key, gotResult, expectedResult)
				}
			}
		})
	}
}
