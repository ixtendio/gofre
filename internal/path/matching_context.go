package path

import "net/url"

type MatchingContext struct {
	// The original request path
	originalPath string
	// The non-empty path segments where the double slashes were removed
	pathSegments []segment
	CaptureVars  map[string]string
}

func (mc *MatchingContext) MatchPatterns(patterns []*Pattern) (Pattern, bool) {
	patternsLen := len(patterns)
	urlLen := len(mc.pathSegments)
	var matchFuncRecursive func(patternIndex int, urlSegmentIndex int, patternSegmentIndex int) (Pattern, bool)

	matchFuncRecursive = func(patternIndex int, urlSegmentIndex int, patternSegmentIndex int) (Pattern, bool) {
		var foundPattern Pattern
		var found bool
		for pi := patternIndex; pi < patternsLen; pi++ {
			pattern := patterns[pi]
			patternSegmentsLen := len(pattern.segments)
			if urlLen == 0 {
				if patternSegmentsLen == 0 || (pattern.isGreedy() && patternSegmentsLen == 1) {
					return *pattern, true
				}
			} else if (urlLen == pattern.maxMatchableSegments || pattern.isGreedy()) && urlSegmentIndex < urlLen {
				urlSegment := &mc.pathSegments[urlSegmentIndex]
				segmentMatchType := pattern.determinePathSegmentMatchType(urlSegment.val, patternSegmentIndex)
				if segmentMatchType == MatchTypeMultiplePaths {
					for usi := urlSegmentIndex; usi < urlLen; usi++ {
						urlSegment := &mc.pathSegments[usi]
						if nextSmt := pattern.determinePathSegmentMatchType(urlSegment.val, patternSegmentIndex+1); nextSmt != MatchTypeUnknown {
							foundPattern, found = matchFuncRecursive(pi, usi, patternSegmentIndex+1)
						} else {
							if usi == urlLen-1 && patternSegmentIndex == patternSegmentsLen-1 {
								urlSegment.matchType = MatchTypeMultiplePaths
								foundPattern = *pattern
								found = true
							}
						}
						if found {
							break
						}
						urlSegment.matchType = MatchTypeMultiplePaths
					}
				} else if segmentMatchType != MatchTypeUnknown {
					urlSegment.matchType = segmentMatchType
					if urlSegmentIndex < urlLen && patternSegmentIndex < patternSegmentsLen {
						if urlSegmentIndex == urlLen-1 && patternSegmentIndex == patternSegmentsLen-1 {
							foundPattern, found = *pattern, true
						} else {
							foundPattern, found = matchFuncRecursive(pi, urlSegmentIndex+1, patternSegmentIndex+1)
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

	matchFuncIter := func() (Pattern, bool) {
		var patternIndex int
		var patternSegmentIndex int
		var urlSegmentIndex int
		var doReturn bool
		lastGreedyPatternSegmentIndex := -1
		lastGreedyUrlSegmentIndex := -1
		for patternIndex < patternsLen {
			pattern := patterns[patternIndex]
			patternSegmentsLen := len(pattern.segments)
			if urlLen == 0 {
				if patternSegmentsLen == 0 || (pattern.isGreedy() && patternSegmentsLen == 1) {
					return *pattern, true
				}
				patternIndex++
				continue
			}

			if !pattern.isGreedy() && urlLen != pattern.maxMatchableSegments {
				patternIndex++
				continue
			}

			if urlSegmentIndex < urlLen && patternSegmentIndex < patternSegmentsLen {
				urlSegment := &mc.pathSegments[urlSegmentIndex]
				segmentMatchType := pattern.determinePathSegmentMatchType(urlSegment.val, patternSegmentIndex)
				if segmentMatchType == MatchTypeMultiplePaths {
					lastGreedyPatternSegmentIndex = patternSegmentIndex
					lastGreedyUrlSegmentIndex = urlSegmentIndex
					if patternSegmentIndex+1 < patternSegmentsLen && pattern.determinePathSegmentMatchType(urlSegment.val, patternSegmentIndex+1) != MatchTypeUnknown {
						patternSegmentIndex++
					} else {
						if urlSegmentIndex == urlLen-1 && patternSegmentIndex == patternSegmentsLen-1 {
							urlSegment.matchType = MatchTypeMultiplePaths
							return *pattern, true
						}
						urlSegmentIndex++
						urlSegment.matchType = MatchTypeMultiplePaths
					}
				} else if segmentMatchType != MatchTypeUnknown {
					urlSegment.matchType = segmentMatchType
					if urlSegmentIndex == urlLen-1 && patternSegmentIndex == patternSegmentsLen-1 {
						return *pattern, true
					} else {
						urlSegmentIndex++
						patternSegmentIndex++
					}
				} else {
					doReturn = true
				}
			} else {
				doReturn = true
			}

			if doReturn {
				doReturn = false
				if lastGreedyPatternSegmentIndex != -1 {
					urlSegment := &mc.pathSegments[lastGreedyUrlSegmentIndex]
					urlSegment.matchType = MatchTypeMultiplePaths
					patternSegmentIndex = lastGreedyPatternSegmentIndex
					urlSegmentIndex = lastGreedyUrlSegmentIndex + 1
					lastGreedyPatternSegmentIndex = -1
					lastGreedyUrlSegmentIndex = -1
				} else {
					patternIndex++
				}
			}
		}
		return Pattern{}, false
	}
	_ = matchFuncIter

	if p, found := matchFuncRecursive(0, 0, 0); found {
		//if p, found := matchFuncIter(); found {
		if p.captureVarsLen > 0 {
			mc.CaptureVars = make(map[string]string, p.captureVarsLen)

			patternSegmentsLen := len(p.segments)
			var psi int
			for i := 0; i < len(mc.pathSegments); i++ {
				urlSegment := &mc.pathSegments[i]
				if urlSegment.matchType == MatchTypeWithCaptureVars ||
					urlSegment.matchType == MatchTypeWithConstraintCaptureVars {
					for ; psi < patternSegmentsLen; psi++ {
						patternSegment := &p.segments[psi]
						if patternSegment.matchType == MatchTypeWithCaptureVars ||
							patternSegment.matchType == MatchTypeWithConstraintCaptureVars {
							mc.CaptureVars[patternSegment.captureVarName] = urlSegment.val
							psi++
							break
						}
					}
				}
			}
		}
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
	segments := make([]segment, maxSegmentsSize)
	segmentStartPos := -1
	addSegmentFunc := func(seg string) {
		if len(seg) > 0 {
			if seg == ".." {
				segmentsIndex--
				if segmentsIndex < 0 {
					segmentsIndex = 0
				}
			} else {
				segments[segmentsIndex] = segment{val: seg}
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
	}
}
