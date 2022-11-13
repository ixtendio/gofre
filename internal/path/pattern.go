package path

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
)

type segment struct {
	val               string
	matchType         MatchType
	captureVarName    string
	captureVarPattern *regexp.Regexp
}

func (s segment) String() string {
	return s.val
}

type Pattern struct {
	caseInsensitive      bool
	captureVarsLen       int
	maxMatchableSegments int
	priority             uint64
	segments             []segment
	RawValue             string
	Attachment           any
}

func (p *Pattern) HighPriorityThan(other *Pattern) bool {
	if p.priority == other.priority {
		return strings.Compare(p.RawValue, other.RawValue) < 0
	}
	return p.priority < other.priority
}

func (p *Pattern) isGreedy() bool {
	return p.maxMatchableSegments == math.MaxInt64
}

// determinePathSegmentMatchType returns the MatchType for the current URL path segment and the next pattern segment index
// If the current URL path dose not math the pattern, then MatchTypeUnknown, 0 is returned
// If the segmentIndex is the last segment pattern then -1 will be returned as the next pattern segment index
func (p *Pattern) determinePathSegmentMatchType(urlPathSegment string, segmentIndex int) MatchType {
	if !p.isGreedy() && segmentIndex >= p.maxMatchableSegments {
		return MatchTypeUnknown
	}

	segment := &p.segments[segmentIndex]
	matchType := segment.matchType
	if matchType == MatchTypeLiteral {
		if len(urlPathSegment) == len(segment.val) {
			if p.caseInsensitive {
				if strings.EqualFold(urlPathSegment, segment.val) {
					return MatchTypeLiteral
				}
			} else {
				if urlPathSegment == segment.val {
					return MatchTypeLiteral
				}
			}
		}
		return MatchTypeUnknown
	} else if matchType == MatchTypeSinglePath ||
		matchType == MatchTypeWithCaptureVars ||
		matchType == MatchTypeMultiplePaths {
		return matchType
	} else if matchType == MatchTypeWithConstraintCaptureVars {
		if segment.captureVarPattern.MatchString(urlPathSegment) {
			return MatchTypeWithConstraintCaptureVars
		}
		return MatchTypeUnknown
	} else if matchType == MatchTypeRegex {
		if regexSegmentMatch(urlPathSegment, segment.val, p.caseInsensitive) {
			return MatchTypeRegex
		}
		return MatchTypeUnknown
	}

	//matchType := segment.matchType
	//switch matchType {
	//case MatchTypeLiteral:
	//	patternSegment := segment.val
	//	if len(urlPathSegment) == len(patternSegment) {
	//		if p.caseInsensitive {
	//			match = strings.EqualFold(urlPathSegment, patternSegment)
	//		} else {
	//			match = urlPathSegment == patternSegment
	//		}
	//	}
	//case MatchTypeSinglePath, MatchTypeWithCaptureVars, MatchTypeMultiplePaths:
	//	match = true
	//case MatchTypeWithConstraintCaptureVars:
	//	match = segment.captureVarPattern.MatchString(urlPathSegment)
	//case MatchTypeRegex:
	//	patternSegment := segment.val
	//	match = regexSegmentMatch(urlPathSegment, patternSegment, p.caseInsensitive)
	//}
	//if match {
	//	return matchType
	//}
	return MatchTypeUnknown
}

func (p *Pattern) String() string {
	return p.RawValue
}

func ParsePattern(pathPattern string, caseInsensitive bool) (Pattern, error) {
	if len(pathPattern) == 0 || pathPattern[0] != '/' {
		return Pattern{}, fmt.Errorf("the path pattern should start with /")
	}

	if pathPattern == "/" {
		return Pattern{
			caseInsensitive: caseInsensitive,
			RawValue:        "/",
		}, nil
	}

	pathLen := len(pathPattern)
	var maxSegmentsSize int
	for pos := 0; pos < pathLen; pos++ {
		if pathPattern[pos] == '/' {
			maxSegmentsSize++
		}
	}

	var pathSegmentStart int
	var pathSegmentsCount int
	var captureVarsLen int
	var lastSegmentMatchType MatchType
	var maxSegmentMatchType MatchType
	segments := make([]segment, maxSegmentsSize)
	pathPatternLen := len(pathPattern)

	for pos := 1; pos < pathPatternLen; pos++ {
		if pathPattern[pos] == '/' || pos == pathPatternLen-1 {
			var segmentVal string
			if pathPattern[pos] == '/' {
				segmentVal = pathPattern[pathSegmentStart+1 : pos]
			} else {
				segmentVal = pathPattern[pathSegmentStart+1:]
			}
			if err := validatePathSegment(segmentVal); err != nil {
				return Pattern{}, fmt.Errorf("invalid path pattern: [%s], failed path segment validation: [%s], err: %w", pathPattern, segmentVal, err)
			}

			pathSegmentStart = pos
			segmentMatchType := determineMatchTypeForSegment(segmentVal)
			if lastSegmentMatchType == MatchTypeMultiplePaths && segmentMatchType == MatchTypeMultiplePaths {
				return Pattern{}, fmt.Errorf("invalid path pattern: [%s], not allowed to have consecutive path segments with **: [%s]", pathPattern, segmentVal)
			}
			lastSegmentMatchType = segmentMatchType
			if segmentMatchType > maxSegmentMatchType {
				maxSegmentMatchType = segmentMatchType
			}

			segment := segment{
				val:               segmentVal,
				matchType:         segmentMatchType,
				captureVarName:    "",
				captureVarPattern: nil,
			}
			if segmentMatchType == MatchTypeWithCaptureVars {
				captureVarsLen++
				segment.captureVarName = segmentVal[1 : len(segmentVal)-1]
			} else if segmentMatchType == MatchTypeWithConstraintCaptureVars {
				captureVarsLen++
				colonStartIndex := strings.IndexRune(segmentVal, ':')
				regexPattern := segmentVal[colonStartIndex+1 : len(segmentVal)-1]
				if caseInsensitive {
					regexPattern = "(?i)" + regexPattern
				}
				regex, err := regexp.Compile(regexPattern)
				if err != nil {
					return Pattern{}, fmt.Errorf("invalid path pattern: [%s], failed to compile regex: [%s], err: %w", pathPattern, regexPattern, err)
				}
				segment.captureVarName = segmentVal[1:colonStartIndex]
				segment.captureVarPattern = regex
			}
			segments[pathSegmentsCount] = segment
			pathSegmentsCount++
		}
	}

	maxMatchableSegments := pathSegmentsCount
	if maxSegmentMatchType == MatchTypeMultiplePaths {
		maxMatchableSegments = math.MaxInt
	}
	segments = segments[0:pathSegmentsCount]
	return Pattern{
		caseInsensitive:      caseInsensitive,
		captureVarsLen:       captureVarsLen,
		maxMatchableSegments: maxMatchableSegments,
		priority:             computePriority(segments),
		segments:             segments,
		RawValue:             pathPattern,
		Attachment:           nil,
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
