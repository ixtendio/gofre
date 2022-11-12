package path

import "net/url"

type MatchingContext struct {
	// The original request path
	originalPath string
	// The non-empty path segments where the double slashes were removed
	pathSegments []string
	CaptureVars  map[string]string
	// the path segment encode
	pathEncode encode
}

func (mc *MatchingContext) MatchPatterns(patterns []Pattern) (Pattern, bool) {
	patternsLen := len(patterns)
	urlLen := len(mc.pathSegments)
	urlPathEncode := mc.pathEncode
	var matchFunc func(patternIndex int, urlSegmentIndex int, patternSegmentIndex int) (Pattern, bool)

	matchFunc = func(patternIndex int, urlSegmentIndex int, patternSegmentIndex int) (Pattern, bool) {
		var foundPattern Pattern
		var found bool
		for pi := patternIndex; pi < patternsLen; pi++ {
			pattern := patterns[pi]
			if urlLen == 0 {
				if pattern.segmentsCount == 0 || (pattern.isGreedy() && pattern.segmentsCount == 1) {
					return pattern, true
				}
			} else if urlSegmentIndex < urlLen && urlLen <= pattern.maxMatchableSegmentsCount {
				segment := mc.pathSegments[urlSegmentIndex]
				segmentMatchType := pattern.determinePathSegmentMatchType(segment, patternSegmentIndex)
				if segmentMatchType == MatchTypeMultiplePaths {
					for usi := urlSegmentIndex; usi < urlLen; usi++ {
						segment := mc.pathSegments[usi]
						if nextSmt := pattern.determinePathSegmentMatchType(segment, patternSegmentIndex+1); nextSmt != MatchTypeUnknown {
							foundPattern, found = matchFunc(pi, usi, patternSegmentIndex+1)
						} else {
							if usi == urlLen-1 && patternSegmentIndex == pattern.segmentsCount-1 {
								urlPathEncode = urlPathEncode.set(usi, MatchTypeMultiplePaths)
								foundPattern = pattern
								found = true
							}
						}
						if found {
							break
						}
						urlPathEncode = urlPathEncode.set(usi, MatchTypeMultiplePaths)
					}
				} else if segmentMatchType != MatchTypeUnknown {
					urlPathEncode = urlPathEncode.set(urlSegmentIndex, segmentMatchType)
					if urlSegmentIndex < urlLen && patternSegmentIndex < pattern.segmentsCount {
						if urlSegmentIndex == urlLen-1 && patternSegmentIndex == pattern.segmentsCount-1 {
							foundPattern, found = pattern, true
						} else {
							foundPattern, found = matchFunc(pi, urlSegmentIndex+1, patternSegmentIndex+1)
						}
					}
				}
			}

			if found {
				return foundPattern, true
			}
		}
		return foundPattern, found
	}

	if p, found := matchFunc(0, 0, 0); found {
		captureVarsLen := len(p.captureVars)
		if captureVarsLen != 0 {
			mc.CaptureVars = make(map[string]string, captureVarsLen)
			var usi int
			var cvi int
			l, r := urlPathEncode.split(0)
			for l.len > 0 {
				if MatchType(l.val) == MatchTypeWithCaptureVars ||
					MatchType(l.val) == MatchTypeWithConstraintCaptureVars {
					varName := p.captureVars[cvi].name
					varValue := mc.pathSegments[usi]
					mc.CaptureVars[varName] = varValue
					cvi++
				}
				l, r = r.split(0)
				usi++
			}
		}
		mc.pathEncode = urlPathEncode
		return p, true
	}
	return Pattern{}, false
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
	segments := make([]string, maxSegmentsSize)
	segmentStartPos := -1
	addSegmentFunc := func(segment string) {
		if len(segment) > 0 {
			if segment == ".." {
				segmentsIndex--
				if segmentsIndex < 0 {
					segmentsIndex = 0
				}
			} else {
				segments[segmentsIndex] = segment
				segmentsIndex++
			}
		}
	}

	for pos := 0; pos < pathLen; pos++ {
		ch := requestPath[pos]
		if ch == '/' {
			if segmentStartPos != -1 {
				addSegmentFunc(requestPath[segmentStartPos:pos])
			}
			segmentStartPos = pos + 1
		}
	}

	if segmentStartPos != -1 && segmentStartPos < pathLen {
		addSegmentFunc(requestPath[segmentStartPos:pathLen])
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
		pathEncode:   newUrlPathEncode(segmentsIndex),
	}
}
