package path

import (
	"errors"
	"fmt"
	"net/url"
)

func ParseURL(requestUrl *url.URL) *MatchingContext {
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
		if el != "" {
			elements = addPathSegment(elements, el)
		}
	}
	return &MatchingContext{
		OriginalPath: requestPath,
		PathElements: elements,
	}
}

func ParsePattern(pathPattern string, caseInsensitive bool) (*Element, error) {
	if pathPattern[0] != '/' {
		return nil, fmt.Errorf("the path pattern should start with /")
	}
	pathPatternId := generateUniqueId(6)
	pathPatternLen := len(pathPattern)
	root := separatorElement(pathPatternId)
	head := root
	pos := 1
	elementStartPos := 1
	for pos < pathPatternLen {
		ch := pathPattern[pos]
		switch ch {
		case '/':
			if elementStartPos != -1 {
				nextElement, err := parseElement(pathPatternId, pathPattern[elementStartPos:pos], caseInsensitive)
				if err != nil {
					return nil, err
				}
				head = head.linkNext(nextElement)
			}
			head = addSeparatorElement(pathPatternId, head)
			elementStartPos = pos + 1
		case '%':
			if pos+2 < pathPatternLen &&
				pathPattern[pos+1] == '2' &&
				(pathPattern[pos+2] == 'f' || pathPattern[pos+2] == 'F') {
				nextElement, err := parseElement(pathPatternId, pathPattern[elementStartPos:pos], caseInsensitive)
				if err != nil {
					return nil, err
				}
				head = head.linkNext(nextElement)
				head = addSeparatorElement(pathPatternId, head)
				pos = pos + 2
			}
			elementStartPos = pos + 1
		}
		pos++
	}
	if elementStartPos != -1 && elementStartPos < pathPatternLen {
		nextElement, err := parseElement(pathPatternId, pathPattern[elementStartPos:pathPatternLen], caseInsensitive)
		if err != nil {
			return nil, err
		}
		head.linkNext(nextElement)
	}
	return root, nil
}

func addSeparatorElement(pathPatternId string, head *Element) *Element {
	if head.MatchType == MatchSeparatorType {
		return head
	}
	return head.linkNext(separatorElement(pathPatternId))
}

func parseElement(pathPatternId string, element string, caseInsensitive bool) (*Element, error) {
	elementLen := len(element)
	if elementLen == 0 {
		return nil, nil
	}
	if element[0] == '{' && element[elementLen-1] == '}' {
		if elementLen == 2 {
			return nil, errors.New("empty capture var element: {} not allowed in the path pattern")
		}
		return captureVarElement(pathPatternId, element[0:elementLen], caseInsensitive)
	}
	return nonCaptureVarElement(pathPatternId, element, caseInsensitive)
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
