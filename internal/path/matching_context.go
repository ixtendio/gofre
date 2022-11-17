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

type MatchingContext struct {
	// The original request path
	originalPath string
	// The non-empty path segments where the double slashes were removed
	pathSegments []segment
	// The capture vars, if exists
	captureVars []CaptureVar
}

func (mc *MatchingContext) CaptureVars() []CaptureVar {
	return mc.captureVars
}

func (mc *MatchingContext) String() string {
	return mc.originalPath
}

func ParseURLPath(requestUrl *url.URL) MatchingContext {
	requestPath := requestUrl.Path
	if len(requestPath) == 0 || requestPath == "/" {
		return MatchingContext{
			originalPath: requestPath,
			pathSegments: nil,
		}
	}

	pathLen := len(requestPath)
	var maxSegmentsSize int
	for pos := 0; pos < pathLen; pos++ {
		if requestPath[pos] == '/' {
			maxSegmentsSize++
		}
	}

	var segmentsIndex int
	segments := make([]segment, maxSegmentsSize)
	segmentStartPos := -1
	addSegment := func(seg string) {
		if len(seg) > 0 {
			if seg == ".." {
				segmentsIndex--
				if segmentsIndex < 0 {
					segmentsIndex = 0
				}
			} else {
				segments[segmentsIndex] = segment{
					val: seg,
				}
				segmentsIndex++
			}
		}
	}

	for pos := 0; pos < pathLen; pos++ {
		ch := requestPath[pos]
		if ch == '/' {
			if segmentStartPos != -1 {
				addSegment(requestPath[segmentStartPos:pos])
			}
			segmentStartPos = pos + 1
		}
	}

	if segmentStartPos != -1 && segmentStartPos < pathLen {
		addSegment(requestPath[segmentStartPos:pathLen])
	}
	if segmentsIndex == 0 {
		return MatchingContext{
			originalPath: requestPath,
			pathSegments: nil,
		}
	}
	return MatchingContext{
		originalPath: requestPath,
		pathSegments: segments[0:segmentsIndex],
	}
}
