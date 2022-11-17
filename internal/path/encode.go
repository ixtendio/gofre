package path

type MatchType uint8

const (
	maxPathSegments               = 19
	MatchTypeUnknown              = MatchType(0)
	MatchTypeLiteral              = MatchType(1)
	MatchTypeConstraintCaptureVar = MatchType(2)
	MatchTypeCaptureVar           = MatchType(3)
	MatchTypeRegex                = MatchType(4)
	MatchTypeSingleSegment        = MatchType(5)
	MatchTypeMultipleSegments     = MatchType(6)
)

func computePriority(segments []*segment) uint64 {
	var priority uint64
	var multiplePathsMatcherSegmentsCount int
	digitsCount := len(segments)
	for i := 0; i < digitsCount; i++ {
		mt := segments[i].matchType
		if mt == MatchTypeMultipleSegments {
			multiplePathsMatcherSegmentsCount++
		}
		priority = priority*10 + uint64(mt)
	}

	if multiplePathsMatcherSegmentsCount == 0 {
		return priority * getDecimalDivider(maxPathSegments-digitsCount)
	} else {
		digitsPerSegmentToAdd := (maxPathSegments - digitsCount) / multiplePathsMatcherSegmentsCount
		digitsPerSegmentReminderToAdd := (maxPathSegments - digitsCount) % multiplePathsMatcherSegmentsCount

		var getSplitIndex = func(nr uint64, dividerDigits int) int {
			var splitIndex int
			m := getDecimalDivider(dividerDigits)
			for n := nr; n > 0; n = n / m {
				if MatchType(n%10) == MatchTypeMultipleSegments {
					break
				}
				splitIndex++
			}
			return splitIndex + dividerDigits
		}

		dividerDigits := 1
		for i := 0; i < multiplePathsMatcherSegmentsCount; i++ {
			splitIndex := getSplitIndex(priority, dividerDigits)
			m := getDecimalDivider(splitIndex)
			n := priority / m
			r := priority % m
			digitsToAdd := digitsPerSegmentToAdd
			if i == multiplePathsMatcherSegmentsCount-1 {
				digitsToAdd += digitsPerSegmentReminderToAdd
			}

			for j := 0; j < digitsToAdd; j++ {
				n = n*10 + uint64(MatchTypeMultipleSegments)
			}
			priority = n*m + r
			dividerDigits = splitIndex + digitsToAdd
		}

		return priority
	}
}

func getDecimalDivider(zeroCount int) uint64 {
	switch zeroCount {
	case 18:
		return 1000000000000000000
	case 17:
		return 100000000000000000
	case 16:
		return 10000000000000000
	case 15:
		return 1000000000000000
	case 14:
		return 100000000000000
	case 13:
		return 10000000000000
	case 12:
		return 1000000000000
	case 11:
		return 100000000000
	case 10:
		return 10000000000
	case 9:
		return 1000000000
	case 8:
		return 100000000
	case 7:
		return 10000000
	case 6:
		return 1000000
	case 5:
		return 100000
	case 4:
		return 10000
	case 3:
		return 1000
	case 2:
		return 100
	case 1:
		return 10
	default:
		return 1
	}
}
