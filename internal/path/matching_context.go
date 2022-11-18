package path

import (
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

type MatchingContext struct {
	PathSegments []UrlSegment
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
