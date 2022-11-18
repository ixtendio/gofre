package path

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
)

const greedyPatternMaxMatchableSegments = math.MaxUint8

type Segment struct {
	val               string
	matchType         MatchType
	captureVarName    string
	captureVarPattern *regexp.Regexp
}

func (s *Segment) matchUrlPathSegment(urlPath string, urlSegment *UrlSegment, caseInsensitive bool) MatchType {
	matchType := s.matchType
	if matchType == MatchTypeLiteral {
		urlSegmentVal := urlPath[urlSegment.startIndex:urlSegment.endIndex]
		if len(urlSegmentVal) == len(s.val) {
			if caseInsensitive {
				if strings.EqualFold(urlSegmentVal, s.val) {
					return MatchTypeLiteral
				}
			} else {
				if urlSegmentVal == s.val {
					return MatchTypeLiteral
				}
			}
		}
		return MatchTypeUnknown
	} else if matchType == MatchTypeSingleSegment ||
		matchType == MatchTypeCaptureVar ||
		matchType == MatchTypeMultipleSegments {
		return matchType
	} else if matchType == MatchTypeConstraintCaptureVar {
		urlSegmentVal := urlPath[urlSegment.startIndex:urlSegment.endIndex]
		if s.captureVarPattern.MatchString(urlSegmentVal) {
			return MatchTypeConstraintCaptureVar
		}
		return MatchTypeUnknown
	} else if matchType == MatchTypeRegex {
		urlSegmentVal := urlPath[urlSegment.startIndex:urlSegment.endIndex]
		if regexSegmentMatch(urlSegmentVal, s.val, caseInsensitive) {
			return MatchTypeRegex
		}
		return MatchTypeUnknown
	}
	return MatchTypeUnknown
}

func (s *Segment) String() string {
	return s.val
}

type Pattern struct {
	caseInsensitive      bool
	captureVarsLen       uint8
	maxMatchableSegments uint8
	priority             uint64
	segments             []*Segment
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
	return p.maxMatchableSegments == greedyPatternMaxMatchableSegments
}

func (p *Pattern) String() string {
	return p.RawValue
}

func ParsePattern(pathPattern string, caseInsensitive bool) (*Pattern, error) {
	if len(pathPattern) == 0 || pathPattern[0] != '/' {
		return nil, fmt.Errorf("the path pattern should start with /")
	}

	if pathPattern == "/" {
		return &Pattern{
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
	var pathSegmentsCount uint8
	var captureVarsLen uint8
	var lastSegmentMatchType MatchType
	var maxSegmentMatchType MatchType
	segments := make([]*Segment, maxSegmentsSize)
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
				return nil, fmt.Errorf("invalid path pattern: [%s], failed path Segment validation: [%s], err: %w", pathPattern, segmentVal, err)
			}

			pathSegmentStart = pos
			segmentMatchType := determineMatchTypeForSegment(segmentVal)
			if lastSegmentMatchType == MatchTypeMultipleSegments && segmentMatchType == MatchTypeMultipleSegments {
				return nil, fmt.Errorf("invalid path pattern: [%s], not allowed to have consecutive path segments with **: [%s]", pathPattern, segmentVal)
			}
			lastSegmentMatchType = segmentMatchType
			if segmentMatchType > maxSegmentMatchType {
				maxSegmentMatchType = segmentMatchType
			}

			segment := &Segment{
				val:               segmentVal,
				matchType:         segmentMatchType,
				captureVarName:    "",
				captureVarPattern: nil,
			}
			if segmentMatchType == MatchTypeCaptureVar {
				captureVarsLen++
				segment.captureVarName = segmentVal[1 : len(segmentVal)-1]
			} else if segmentMatchType == MatchTypeConstraintCaptureVar {
				captureVarsLen++
				colonStartIndex := strings.IndexRune(segmentVal, ':')
				regexPattern := segmentVal[colonStartIndex+1 : len(segmentVal)-1]
				if caseInsensitive {
					regexPattern = "(?i)" + regexPattern
				}
				regex, err := regexp.Compile(regexPattern)
				if err != nil {
					return nil, fmt.Errorf("invalid path pattern: [%s], failed to compile regex: [%s], err: %w", pathPattern, regexPattern, err)
				}
				segment.captureVarName = segmentVal[1:colonStartIndex]
				segment.captureVarPattern = regex
			}
			segments[pathSegmentsCount] = segment
			pathSegmentsCount++
		}
	}

	maxMatchableSegments := pathSegmentsCount
	if maxSegmentMatchType == MatchTypeMultipleSegments {
		maxMatchableSegments = greedyPatternMaxMatchableSegments
	}
	segments = segments[0:pathSegmentsCount]
	return &Pattern{
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
		return MatchTypeSingleSegment
	} else if pathSegment == "**" {
		return MatchTypeMultipleSegments
	}
	if pathSegment[0] == '{' && pathSegment[len(pathSegment)-1] == '}' {
		for pos := 0; pos < len(pathSegment); pos++ {
			ch := pathSegment[pos]
			if ch == ':' {
				return MatchTypeConstraintCaptureVar
			}
		}
		return MatchTypeCaptureVar
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
		return errors.New("empty path Segment")
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
