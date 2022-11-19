package path

import (
	"net/http"
	"net/url"
)

type CaptureVar struct {
	Name  string
	Value string
}

func (cv *CaptureVar) String() string {
	return cv.Name + "=" + cv.Value
}

type UrlSegment struct {
	startIndex uint16
	endIndex   uint16
	matchType  MatchType
}

func (s *UrlSegment) Reset() {
	s.startIndex = 0
	s.endIndex = 0
	s.matchType = 0
}

type MatchingContext struct {
	R              *http.Request
	matchedPattern *Pattern
	PathSegments   []UrlSegment
}

func (mc *MatchingContext) Clone() MatchingContext {
	segmentsLen := len(mc.PathSegments)
	segments := make([]UrlSegment, segmentsLen)
	for i := 0; i < segmentsLen; i++ {
		s := &mc.PathSegments[i]
		segments[i] = UrlSegment{
			startIndex: s.startIndex,
			endIndex:   s.endIndex,
			matchType:  s.matchType,
		}
	}
	return MatchingContext{
		R:              mc.R,
		matchedPattern: mc.matchedPattern,
		PathSegments:   segments,
	}
}

func (mc *MatchingContext) PathVar(name string) string {
	p := mc.matchedPattern
	if p == nil || p.captureVarsLen == 0 {
		return ""
	}

	var captureVarsIndex int
	var patternSegment *segment
	patternSegmentsLen := len(p.segments)
	for psi := 0; psi < patternSegmentsLen; psi++ {
		ps := p.segments[psi]
		if ps.matchType == MatchTypeCaptureVar || ps.matchType == MatchTypeConstraintCaptureVar {
			if ps.captureVarName == name {
				patternSegment = ps
				break
			}
			captureVarsIndex++
		}
	}

	if patternSegment == nil {
		return ""
	}

	matchingPath := mc.R.URL.Path
	for i := 0; i < len(mc.PathSegments); i++ {
		urlSegment := &mc.PathSegments[i]
		if urlSegment.matchType == MatchTypeCaptureVar ||
			urlSegment.matchType == MatchTypeConstraintCaptureVar {
			captureVarsIndex--
			if captureVarsIndex == -1 {
				return matchingPath[urlSegment.startIndex:urlSegment.endIndex]
			}
		}
	}
	return ""
}

func ParseURLPath(requestUrl *url.URL, mc *MatchingContext) {
	requestPath := requestUrl.Path
	if len(requestPath) == 0 || requestPath == "/" {
		mc.PathSegments = nil
		return
	}

	pathLen := len(requestPath)
	var segmentsIndex int
	segmentStartPos := -1
	addSegment := func(startIndex int, endIndex int) bool {
		seg := requestPath[segmentStartPos:endIndex]
		if len(seg) > 0 {
			if seg == ".." {
				segmentsIndex--
				if segmentsIndex < 0 {
					segmentsIndex = 0
				}
			} else {
				if segmentsIndex >= MaxPathSegments {
					return false
				}

				mc.PathSegments[segmentsIndex] = UrlSegment{
					startIndex: uint16(startIndex),
					endIndex:   uint16(endIndex),
				}
				segmentsIndex++
			}
		}
		return true
	}

	for pos := 0; pos < pathLen; pos++ {
		ch := requestPath[pos]
		if ch == '/' {
			if segmentStartPos != -1 {
				if !addSegment(segmentStartPos, pos) {
					mc.PathSegments = nil
					return
				}
			}
			segmentStartPos = pos + 1
		}
	}

	if segmentStartPos != -1 && segmentStartPos < pathLen {
		if !addSegment(segmentStartPos, pathLen) {
			mc.PathSegments = nil
			return
		}
	}
	if segmentsIndex == 0 {
		mc.PathSegments = nil
	}
	mc.PathSegments = mc.PathSegments[0:segmentsIndex]
}
