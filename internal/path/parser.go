package path

import (
	"fmt"
	"net/url"
)

func ParseRequestURL(requestUrl *url.URL) *MatchingContext {
	requestPath := requestUrl.Path
	pathLen := len(requestPath)
	var elements []string
	elementStartPos := -1
	for pos := 0; pos < pathLen; pos++ {
		ch := requestPath[pos]
		if ch == '/' {
			if elementStartPos != -1 {
				elements = addPathSegment(elements, requestPath[elementStartPos:pos])
			}
			elements = addPathSeparator(elements)
			elementStartPos = pos + 1
		}
	}
	if elementStartPos != -1 && elementStartPos < pathLen {
		el := requestPath[elementStartPos:pathLen]
		if el == Separator {
			elements = addPathSeparator(elements)
		} else if el != "" {
			elements = addPathSegment(elements, el)
		}
	}
	return &MatchingContext{
		originalPath:          requestPath,
		elements:              elements,
		ExtractedUriVariables: make(map[string]string),
	}
}

func Parse(pathPattern string, caseInsensitive bool) ([]*Element, error) {
	if pathPattern[0] != '/' {
		return nil, fmt.Errorf("the path pattern should start with /")
	}
	var elements []*Element
	pathPatternLen := len(pathPattern)
	pos := 0
	elementStartPos := -1
	for pos < pathPatternLen {
		ch := pathPattern[pos]
		switch ch {
		case '/':
			if elementStartPos != -1 {
				nextElement, err := parseElement(pathPattern[elementStartPos:pos], caseInsensitive)
				if err != nil {
					return nil, err
				}
				if nextElement != nil {
					elements = append(elements, nextElement)
				}
			}
			elements = addSeparatorElement(elements)
			elementStartPos = pos + 1
		case '%':
			if pos+2 < pathPatternLen &&
				pathPattern[pos+1] == '2' &&
				(pathPattern[pos+2] == 'f' || pathPattern[pos+2] == 'F') {
				nextElement, err := parseElement(pathPattern[elementStartPos:pos], caseInsensitive)
				if err != nil {
					return nil, err
				}
				if nextElement != nil {
					elements = append(elements, nextElement)
				}
				elements = addSeparatorElement(elements)
				pos = pos + 2
			}
			elementStartPos = pos + 1
		}
		pos++
	}
	if elementStartPos != -1 && elementStartPos < pathPatternLen {
		nextElement, err := parseElement(pathPattern[elementStartPos:pathPatternLen], caseInsensitive)
		if err != nil {
			return nil, err
		}
		if nextElement != nil {
			elements = append(elements, nextElement)
		}
	}
	return elements, nil
}

func addSeparatorElement(elements []*Element) []*Element {
	elLen := len(elements)
	if elLen > 0 && elements[elLen-1].MatchType == MatchSeparatorType {
		return elements
	}
	return append(elements, separatorElement())
}

func parseElement(element string, caseInsensitive bool) (*Element, error) {
	elementLen := len(element)
	if elementLen == 0 {
		return nil, nil
	}
	if element[0] == '{' && element[elementLen-1] == '}' {
		if elementLen == 2 {
			return nil, nil
		}
		return captureVarElement(element[0:elementLen], caseInsensitive)
	}
	return nonCaptureVarElement(element, caseInsensitive)
}

func addPathSegment(elements []string, segment string) []string {
	if segment == "" {
		return elements
	}
	if segment == ".." {
		length := len(elements)
		switch length {
		case 0, 1:
			return elements
		case 2:
			return elements[0:1]
		default:
			if elements[length-1] == Separator {
				return elements[0 : length-2]
			}
			return elements[0 : length-1]
		}
	} else {
		return append(elements, segment)
	}
}

func addPathSeparator(elements []string) []string {
	elementsLen := len(elements)
	if elementsLen == 0 || elements[elementsLen-1] != Separator {
		return append(elements, Separator)
	}
	return elements
}
