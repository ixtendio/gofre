package path

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
)

type captureVar struct {
	name         string
	segmentIndex int
	pattern      *regexp.Regexp
}

type Pattern struct {
	segmentsCount             int
	caseInsensitive           bool
	maxMatchableSegmentsCount int
	pathEncode                encode
	RawValue                  string
	captureVars               []captureVar
	Attachment                any
}

func (p *Pattern) HighPriorityThan(pattern Pattern) bool {
	return p.pathEncode.val < pattern.pathEncode.val
}

func (p *Pattern) isGreedy() bool {
	return p.maxMatchableSegmentsCount == math.MaxInt64
}

// determinePathSegmentMatchType returns the MatchType for the current URL path segment and the next pattern segment index
// If the current URL path dose not math the pattern, then MatchTypeUnknown, 0 is returned
// If the segmentIndex is the last segment pattern then -1 will be returned as the next pattern segment index
func (p *Pattern) determinePathSegmentMatchType(urlPathSegment string, segmentIndex int) MatchType {
	if segmentIndex < 0 || segmentIndex >= p.segmentsCount {
		return MatchTypeUnknown
	}
	if !p.isGreedy() && segmentIndex >= p.maxMatchableSegmentsCount {
		return MatchTypeUnknown
	}
	var match bool
	matchType := p.getSegmentMatchType(segmentIndex)
	switch matchType {
	case MatchTypeLiteral:
		patternSegment := p.getSegment(segmentIndex)
		if p.caseInsensitive {
			match = strings.EqualFold(urlPathSegment, patternSegment)
		} else {
			match = urlPathSegment == patternSegment
		}
	case MatchTypeSinglePath, MatchTypeWithCaptureVars, MatchTypeMultiplePaths:
		match = true
	case MatchTypeWithConstraintCaptureVars:
		for i := 0; i < len(p.captureVars); i++ {
			cv := p.captureVars[i]
			if cv.segmentIndex == segmentIndex && cv.pattern != nil {
				match = cv.pattern.MatchString(urlPathSegment)
				break
			}
		}
	case MatchTypeRegex:
		patternSegment := p.getSegment(segmentIndex)
		match = regexSegmentMatch(urlPathSegment, patternSegment, p.caseInsensitive)
	}
	if match {
		return matchType
	}
	return MatchTypeUnknown
}

func (p *Pattern) getSegmentMatchType(lookupSegmentIndex int) MatchType {
	var segmentIndex int
	prevMatchType := MatchTypeUnknown
	n := p.pathEncode.val
	for i := maxPathSegments - 1; i >= 0; i-- {
		spliter := getDecimalDivider(i)
		currentMatchType := MatchType(n / spliter)
		n = n % spliter
		if prevMatchType == MatchTypeMultiplePaths && currentMatchType == MatchTypeMultiplePaths {
			continue
		}
		if segmentIndex == lookupSegmentIndex {
			return currentMatchType
		}
		prevMatchType = currentMatchType
		segmentIndex++
	}
	return MatchTypeUnknown
}

func (p *Pattern) getSegment(segmentIndex int) string {
	var pathSegmentStart int
	var pathSegmentsCount int
	pathPattern := p.RawValue
	pathPatternLen := len(pathPattern)
	for pos := 1; pos < pathPatternLen; pos++ {
		if pathPattern[pos] == '/' || pos == pathPatternLen-1 {
			if pathSegmentsCount == segmentIndex {
				if pathPattern[pos] == '/' {
					return pathPattern[pathSegmentStart+1 : pos]
				} else {
					return pathPattern[pathSegmentStart+1:]
				}
			}
			pathSegmentsCount++
			pathSegmentStart = pos
		}
	}
	return ""
}

func (p *Pattern) String() string {
	return p.RawValue
}

func ParsePatternImproved(pathPattern string, caseInsensitive bool) (Pattern, error) {
	if len(pathPattern) == 0 || pathPattern[0] != '/' {
		return Pattern{}, fmt.Errorf("the path pattern should start with /")
	}

	pathPatternLen := len(pathPattern)
	captureVars := createCaptureVarsContainers(pathPattern)
	var captureVarsIndex int
	var pathSegmentStart int
	var pathEncode encode
	//var pathSegmentsMatchTypeEncoding uint64
	var pathSegmentsCount int
	var lastSegmentMatchType MatchType
	var maxSegmentMatchType MatchType

	for pos := 1; pos < pathPatternLen; pos++ {
		if pathPattern[pos] == '/' || pos == pathPatternLen-1 {
			var pathSegment string
			if pathPattern[pos] == '/' {
				pathSegment = pathPattern[pathSegmentStart+1 : pos]
			} else {
				pathSegment = pathPattern[pathSegmentStart+1:]
			}
			if err := validatePathSegment(pathSegment); err != nil {
				return Pattern{}, fmt.Errorf("invalid path pattern: [%s], failed path segment validation: [%s], err: %w", pathPattern, pathSegment, err)
			}

			pathSegmentStart = pos
			pathSegmentsCount++
			currentSegmentMatchType := determineMatchTypeForSegment(pathSegment)
			if lastSegmentMatchType == MatchTypeMultiplePaths && currentSegmentMatchType == MatchTypeMultiplePaths {
				return Pattern{}, fmt.Errorf("invalid path pattern: [%s], not allowed to have consecutive path segments with **: [%s]", pathPattern, pathSegment)
			}
			lastSegmentMatchType = currentSegmentMatchType
			if currentSegmentMatchType > maxSegmentMatchType {
				maxSegmentMatchType = currentSegmentMatchType
			}

			pathEncode = pathEncode.append(currentSegmentMatchType)
			if currentSegmentMatchType == MatchTypeWithCaptureVars {
				captureVars[captureVarsIndex] = captureVar{
					segmentIndex: pathSegmentsCount - 1,
					name:         pathSegment[1 : len(pathSegment)-1],
				}
				captureVarsIndex++
			} else if currentSegmentMatchType == MatchTypeWithConstraintCaptureVars {
				colonStartIndex := strings.IndexRune(pathSegment, ':')
				regexPattern := pathSegment[colonStartIndex+1 : len(pathSegment)-1]
				if caseInsensitive {
					regexPattern = "(?i)" + regexPattern
				}
				regex, err := regexp.Compile(regexPattern)
				if err != nil {
					return Pattern{}, fmt.Errorf("invalid path pattern: [%s], failed to compile regex: [%s], err: %w", pathPattern, regexPattern, err)
				}
				captureVars[captureVarsIndex] = captureVar{
					segmentIndex: pathSegmentsCount - 1,
					name:         pathSegment[1:colonStartIndex],
					pattern:      regex,
				}
				captureVarsIndex++
			}
		}
	}
	maxMatchableSegments := pathSegmentsCount
	if maxSegmentMatchType == MatchTypeMultiplePaths {
		maxMatchableSegments = math.MaxInt
	}
	return Pattern{
		segmentsCount:             pathSegmentsCount,
		caseInsensitive:           caseInsensitive,
		maxMatchableSegmentsCount: maxMatchableSegments,
		pathEncode:                pathEncode.doPadding(),
		RawValue:                  pathPattern,
		captureVars:               captureVars,
	}, nil
}

func regexSegmentMatch(urlPathSegment string, patternSegment string, caseInsensitive bool) bool {
	var hasStar bool
	patternSegmentLen := len(patternSegment)
	urlPathSegmentLen := len(urlPathSegment)
	for i := 0; i < patternSegmentLen; i++ {
		if patternSegment[i] == '*' {
			hasStar = true
			break
		}
	}
	if !hasStar && urlPathSegmentLen != patternSegmentLen {
		return false
	}

	runesMatches := func(uc uint8, pc uint8) bool {
		if pc == '?' {
			return true
		}
		if caseInsensitive {
			return unicode.ToUpper(rune(uc)) == unicode.ToUpper(rune(pc))
		} else {
			return uc == pc
		}
	}

	var pIndex int
	var uIndex int
	var checkPointRestore bool
	pIndexCheckPoint := -1
	uIndexCheckPoint := -1
	for {
		if pIndex == patternSegmentLen && uIndex == urlPathSegmentLen {
			return true
		}
		if pIndex == patternSegmentLen || uIndex == urlPathSegmentLen {
			checkPointRestore = true
		}

		if checkPointRestore {
			// no checkpoint, meaning that no greedy character found
			if pIndexCheckPoint == -1 || uIndexCheckPoint == -1 {
				return false
			}
			pIndex = pIndexCheckPoint
			uIndex = uIndexCheckPoint + 1
		}

		checkPointRestore = false
		uc := urlPathSegment[uIndex]
		pc := patternSegment[pIndex]
		if pc == '?' {
			pIndex++
			uIndex++
		} else if pc == '*' {
			//skip the next * if exists (example ab****)
			for i := pIndex + 1; i < patternSegmentLen; i++ {
				if patternSegment[i] != '*' {
					pIndex = i - 1
					break
				}
			}

			if pIndex == patternSegmentLen-1 {
				//the * is the last character in the pattern and matches the rest of the text
				return true
			}

			pIndexCheckPoint = -1
			uIndexCheckPoint = -1
			pc = patternSegment[pIndex+1]
			for ; uIndex < urlPathSegmentLen; uIndex++ {
				uc = urlPathSegment[uIndex]
				if runesMatches(uc, pc) {
					// we save the checkpoint and we continue with the next matching
					pIndexCheckPoint = pIndex
					uIndexCheckPoint = uIndex
					pIndex++
					break
				}
			}
		} else {
			if runesMatches(uc, pc) {
				pIndex++
				uIndex++
			} else {
				checkPointRestore = true
			}
		}
	}

}

func determineMatchTypeForSegment(pathSegment string) MatchType {
	if pathSegment == "*" {
		return MatchTypeSinglePath
	} else if pathSegment == "**" {
		return MatchTypeMultiplePaths
	}
	if pathSegment[0] == '{' && pathSegment[len(pathSegment)-1] == '}' {
		for pos := 0; pos < len(pathSegment); pos++ {
			ch := pathSegment[pos]
			if ch == ':' {
				return MatchTypeWithConstraintCaptureVars
			}
		}
		return MatchTypeWithCaptureVars
	}
	for pos := 0; pos < len(pathSegment); pos++ {
		ch := pathSegment[pos]
		if ch == '?' || ch == '*' {
			return MatchTypeRegex
		}
	}
	return MatchTypeLiteral
}

func validatePathSegment(pathSegment string) error {
	pathSegmentLen := len(pathSegment)
	if pathSegmentLen == 0 {
		return errors.New("empty path segment")
	}

	if pathSegment[0] == '{' && pathSegment[pathSegmentLen-1] != '}' {
		return errors.New("opened bracket without being closed")
	}

	if pathSegment[pathSegmentLen-1] == '}' && pathSegment[0] != '{' {
		return errors.New("closed bracket without being opened")
	}

	if pathSegment[0] == '{' && pathSegment[pathSegmentLen-1] == '}' {
		if pathSegmentLen == 2 {
			return errors.New("empty capture variable name")
		}
		for pos := 1; pos < pathSegmentLen-1; pos++ {
			if pathSegment[pos] == ':' {
				if pos == 1 {
					return errors.New("empty capture variable name")
				}
				if pos == pathSegmentLen-2 {
					return errors.New("empty capture regex constraint")
				}
			}
		}
	}

	if pathSegment == "*" || pathSegment == "**" {
		return nil
	}

	var consecutiveAsterixCount int
	for pos := 0; pos < pathSegmentLen; pos++ {
		if pathSegment[pos] == '*' {
			consecutiveAsterixCount++
		} else {
			consecutiveAsterixCount = 0
		}
		if consecutiveAsterixCount >= 2 {
			return errors.New("not allowed two or more consecutive asterix together with other characters")
		}
	}

	return nil
}

func createCaptureVarsContainers(pathPattern string) []captureVar {
	var captureVarsCount int
	for pos := 1; pos < len(pathPattern); pos++ {
		ch := pathPattern[pos]
		switch ch {
		case '{':
			if pathPattern[pos-1] == '/' {
				captureVarsCount++
			}
		}
	}
	if captureVarsCount == 0 {
		return nil
	}
	return make([]captureVar, captureVarsCount)
}
