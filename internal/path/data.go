package path

import (
	"fmt"
	"regexp"
	"strings"
)

const pathSeparator = "/"

const (
	MatchSeparatorType = iota + 1
	MatchLiteralType
	MatchVarCaptureWithConstraintType
	MatchVarCaptureType
	MatchRegexType
	MatchMultiplePathsType
)

type MatchingContext struct {
	originalPath          string
	elements              []string
	ExtractedUriVariables map[string]string
}

type Element struct {
	// The next path element in the chain
	MatchingType int
	next         *Element
	matcherFunc  func(pathIndex int, mc MatchingContext) bool
	value        string
}

func (e *Element) Matches(pathIndex int, mc MatchingContext) bool {
	if e.matcherFunc(pathIndex, mc) {
		if e.next != nil {
			return e.next.Matches(pathIndex+1, mc)
		}
		return true
	}
	return false
}

func (p *Element) String() string {
	var toString strings.Builder
	c := p
	for c != nil {
		toString.WriteRune('[')
		toString.WriteString(string(c.value))
		toString.WriteRune(']')
		c = c.next
	}
	return toString.String()
}

func separatorElement() *Element {
	return &Element{
		MatchingType: MatchSeparatorType,
		matcherFunc: func(pathIndex int, mc MatchingContext) bool {
			return pathIndex < len(mc.elements) && mc.elements[pathIndex] == pathSeparator
		},
		value: pathSeparator,
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
		MatchingType: kind,
		matcherFunc: func(pathIndex int, mc MatchingContext) bool {
			if pathIndex >= len(mc.elements) {
				return false
			}
			valToCompare := mc.elements[pathIndex]
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
		value: val,
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
	var constraintPattern *regexp.Regexp
	sepIndex := strings.IndexRune(val, ':')
	if sepIndex > 0 {
		varName = val[0:sepIndex]
		if sepIndex+1 < len(val) {
			regexPattern := val[sepIndex+1 : len(val)-1]
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
		MatchingType: kind,
		matcherFunc: func(pathIndex int, mc MatchingContext) bool {
			if pathIndex >= len(mc.elements) {
				return false
			}
			valToMatch := mc.elements[pathIndex]
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
		value: val,
	}, nil
}
