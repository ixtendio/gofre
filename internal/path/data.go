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
}

type Element struct {
	// A unique identifier of the pattern from where this element was extracted
	PathPatternId string
	// The path segment type
	MatchType int
	// The next element in path or nil
	Next *Element
	// The previous element in path or nil
	Previous *Element
	// The string value of the path pattern segment
	RawVal string
	// The match pattern of the path segment, if exists
	MatchPattern string
	// The name of the capture variable or empty
	CaptureVarName string
	// A function that matches a request path segment with the current pattern.
	// If the pattern segment supports variable capture then, the value will be returned
	MatchPathSegment func(pathSegment string) (bool, string)
}

func (e *Element) linkNext(next *Element) *Element {
	if next == nil {
		return e
	}
	e.Next = next
	next.Previous = e
	return next
}

func separatorElement(pathPatternId string) *Element {
	return &Element{
		PathPatternId: pathPatternId,
		MatchType:     MatchSeparatorType,
		MatchPathSegment: func(pathSegment string) (bool, string) {
			return pathSegment == Separator, ""
		},
		RawVal: Separator,
	}
}

func nonCaptureVarElement(pathPatternId string, val string, caseInsensitive bool) (*Element, error) {
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
		PathPatternId: pathPatternId,
		MatchType:     kind,
		MatchPathSegment: func(pathSegment string) (bool, string) {
			switch kind {
			case MatchLiteralType:
				if len(val) != len(pathSegment) {
					return false, ""
				}
				if caseInsensitive {
					return strings.EqualFold(val, pathSegment), ""
				} else {
					return val == pathSegment, ""
				}
			case MatchRegexType:
				return matchPattern.MatchString(pathSegment), ""
			case MatchMultiplePathsType:
				return true, ""
			default:
				return false, ""
			}
		},
		RawVal:       val,
		MatchPattern: val,
	}, nil
}

func captureVarElement(pathPatternId string, val string, caseInsensitive bool) (*Element, error) {
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
		varName = val[1 : len(val)-1]
	}

	return &Element{
		PathPatternId: pathPatternId,
		MatchType:     kind,
		MatchPathSegment: func(pathSegment string) (bool, string) {
			if constraintPattern == nil || constraintPattern.MatchString(pathSegment) {
				return true, pathSegment
			}
			return false, ""
		},
		CaptureVarName: varName,
		RawVal:         val,
		MatchPattern:   regexPattern,
	}, nil
}
