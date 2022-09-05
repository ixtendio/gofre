package path

import (
	"fmt"
	"regexp"
	"strings"
)

const Separator = "/"

const (
	MatchSeparatorType = iota + 1
	MatchLiteralType
	MatchVarCaptureWithConstraintType
	MatchVarCaptureType
	MatchRegexType
	MatchMultiplePathsType
)

type MatchingContext struct {
	// The original request path
	OriginalPath string
	// The path elements where the double slashes were removed
	PathElements []string
	// A map where the var captured are stored
	ExtractedUriVariables map[string]string
}

type Element struct {
	// The path segment type
	MatchType int
	// The string value of the path pattern segment
	RawVal string
	// The match pattern of the path segment, if exists
	MatchPattern string
	// A function that matches a request path segment with the current pattern segment.
	// If the pattern segment supports variables, then these will be published to the MatchingContext
	MatchFunc func(pathIndex int, mc *MatchingContext) bool
}

func separatorElement() *Element {
	return &Element{
		MatchType: MatchSeparatorType,
		MatchFunc: func(pathIndex int, mc *MatchingContext) bool {
			return pathIndex < len(mc.PathElements) && mc.PathElements[pathIndex] == Separator
		},
		RawVal: Separator,
	}
}

func nonCaptureVarElement(val string, caseInsensitive bool) (*Element, error) {
	var matchPattern *regexp.Regexp
	kind := MatchLiteralType
	if val == "**" {
		kind = MatchMultiplePathsType
	} else {
		for i := 0; i < len(val); i++ {
			if val[i] == '*' || val[i] == '?' {
				kind = MatchRegexType
				break
			}
		}
		if kind == MatchRegexType {
			var err error
			var sb strings.Builder
			if caseInsensitive {
				sb.WriteString("(?i)")
			}
			for i := 0; i < len(val); i++ {
				if val[i] == '*' || val[i] == '?' {
					sb.WriteRune('.')
				}
				sb.WriteByte(val[i])
			}
			matchPattern, err = regexp.Compile(sb.String())
			if err != nil {
				return nil, fmt.Errorf("failed compiling regex pattern: %s, err: %w", sb.String(), err)
			}
		}
	}
	return &Element{
		MatchType: kind,
		MatchFunc: func(pathIndex int, mc *MatchingContext) bool {
			if pathIndex >= len(mc.PathElements) {
				return false
			}
			valToCompare := mc.PathElements[pathIndex]
			switch kind {
			case MatchLiteralType:
				if len(val) != len(valToCompare) {
					return false
				}
				if caseInsensitive {
					return strings.EqualFold(val, valToCompare)
				} else {
					return val == valToCompare
				}
			case MatchRegexType:
				return matchPattern.MatchString(valToCompare)
			case MatchMultiplePathsType:
				return true
			default:
				return false
			}
		},
		RawVal:       val,
		MatchPattern: val,
	}, nil
}

func captureVarElement(val string, caseInsensitive bool) (*Element, error) {
	if len(val) < 3 {
		return nil, fmt.Errorf("the capture var path should have at least 3 chars, should start with '{' and end with '}'")
	}
	if val[0] != '{' || val[len(val)-1] != '}' {
		return nil, fmt.Errorf("the capture var path should start with '{' and end with '}'")
	}

	kind := MatchVarCaptureType
	var varName string
	var regexPattern string
	var constraintPattern *regexp.Regexp
	sepIndex := strings.IndexRune(val, ':')
	if sepIndex > 0 {
		varName = val[1:sepIndex]
		if sepIndex+1 < len(val) {
			regexPattern = val[sepIndex+1 : len(val)-1]
			if regexPattern != "*" && regexPattern != ".*" {
				if caseInsensitive {
					regexPattern = "(?i)" + regexPattern
				}
				var err error
				constraintPattern, err = regexp.Compile(regexPattern)
				if err != nil {
					return nil, fmt.Errorf("failed to parse constraintPattern: %s, err: %w", regexPattern, err)
				}
				kind = MatchVarCaptureWithConstraintType
			}
		}
	} else {
		varName = val
	}

	return &Element{
		MatchType: kind,
		MatchFunc: func(pathIndex int, mc *MatchingContext) bool {
			if pathIndex >= len(mc.PathElements) {
				return false
			}
			valToMatch := mc.PathElements[pathIndex]
			if constraintPattern == nil {
				mc.ExtractedUriVariables[varName] = valToMatch
				return true
			} else {
				if constraintPattern.MatchString(valToMatch) {
					mc.ExtractedUriVariables[varName] = valToMatch
					return true
				}
				return false
			}
		},
		RawVal:       val,
		MatchPattern: regexPattern,
	}, nil
}
