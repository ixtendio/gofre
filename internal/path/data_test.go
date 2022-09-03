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
		matchingType int
		value        string
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
				matchingType: MatchLiteralType,
				value:        "text",
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
				matchingType: MatchRegexType,
				value:        "te?t",
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
				matchingType: MatchRegexType,
				value:        "te*t",
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
				matchingType: MatchMultiplePathsType,
				value:        "**",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nonCaptureVarElement(tt.args.val, tt.args.caseInsensitive)
			if (err != nil) != tt.wantErr {
				t.Errorf("nonCaptureVarElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.MatchType != tt.want.matchingType {
				t.Errorf("nonCaptureVarElement() got MatchType = %v, want %v", got.MatchType, tt.want.matchingType)
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
			name: "matching_func-literal_case_insensitive",
			args: args{
				val:             "text",
				caseInsensitive: true,
			},
			want: map[string]bool{"text": true, "TeXt": true},
		},
		{
			name: "matching_func-literal_case_sensitive",
			args: args{
				val:             "text",
				caseInsensitive: false,
			},
			want: map[string]bool{"text": true, "TeXt": false},
		},
		{
			name: "matching_func-literal_match_multiple_paths_type",
			args: args{
				val:             "**",
				caseInsensitive: false,
			},
			want: map[string]bool{"a": true, "Abc": true},
		},
		{
			name: "matching_func-literal_match_regex_type_case_sensitive",
			args: args{
				val:             "abc*hij",
				caseInsensitive: false,
			},
			want: map[string]bool{"abcdefghij": true, "abchij": true, "abcij": false, "Abchij": false},
		},
		{
			name: "matching_func-literal_match_regex_type_case_insensitive",
			args: args{
				val:             "abc*hij",
				caseInsensitive: true,
			},
			want: map[string]bool{"abcdefghij": true, "abchij": true, "abcij": false, "Abchij": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nonCaptureVarElement(tt.args.val, tt.args.caseInsensitive)
			if err != nil {
				t.Errorf("nonCaptureVarElement() error = %v", err)
				return
			}
			var mc MatchingContext
			for key := range tt.want {
				mc.elements = append(mc.elements, key)
			}
			for i := 0; i < len(mc.elements); i++ {
				key := mc.elements[i]
				expectedResult := tt.want[key]
				if got.matcherFunc(i, mc) != expectedResult {
					t.Errorf("nonCaptureVarElement().matcherFunc(%d, '%s') = %v, want %v", i, key, !expectedResult, expectedResult)
				}
			}
		})
	}
}
