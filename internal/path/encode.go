package path

import "strconv"

type MatchType int

const (
	maxPathSegments                    = 19
	MatchTypeUnknown                   = MatchType(0)
	MatchTypeLiteral                   = MatchType(1)
	MatchTypeWithConstraintCaptureVars = MatchType(2)
	MatchTypeWithCaptureVars           = MatchType(3)
	MatchTypeRegex                     = MatchType(4)
	MatchTypeSinglePath                = MatchType(5)
	MatchTypeMultiplePaths             = MatchType(6)
	matchTypeUnset                     = MatchType(9)
)

type encode struct {
	val uint64
	len int
}

func (e encode) doPadding() encode {
	val := e.val
	if val == 0 || e.len == maxPathSegments {
		return e
	}

	var multiplePathsMatcherSegmentsCount int
	var digitsCount int
	for n := val; n > 0; n = n / 10 {
		if MatchType(n%10) == MatchTypeMultiplePaths {
			multiplePathsMatcherSegmentsCount++
		}
		digitsCount++
	}

	if multiplePathsMatcherSegmentsCount == 0 {
		return encode{val: val * getDecimalDivider(maxPathSegments-digitsCount), len: maxPathSegments}
	} else {
		digitsPerSegmentToAdd := (maxPathSegments - digitsCount) / multiplePathsMatcherSegmentsCount
		digitsPerSegmentReminderToAdd := (maxPathSegments - digitsCount) % multiplePathsMatcherSegmentsCount

		var getSplitIndex = func(nr uint64, dividerDigits int) int {
			var splitIndex int
			m := getDecimalDivider(dividerDigits)
			for n := nr; n > 0; n = n / m {
				if MatchType(n%10) == MatchTypeMultiplePaths {
					break
				}
				splitIndex++
			}
			return splitIndex + dividerDigits
		}

		dividerDigits := 1
		for i := 0; i < multiplePathsMatcherSegmentsCount; i++ {
			splitIndex := getSplitIndex(val, dividerDigits)
			m := getDecimalDivider(splitIndex)
			n := val / m
			r := val % m
			digitsToAdd := digitsPerSegmentToAdd
			if i == multiplePathsMatcherSegmentsCount-1 {
				digitsToAdd += digitsPerSegmentReminderToAdd
			}

			for j := 0; j < digitsToAdd; j++ {
				n = n*10 + uint64(MatchTypeMultiplePaths)
			}
			val = n*m + r
			dividerDigits = splitIndex + digitsToAdd
		}

		return encode{val: val, len: maxPathSegments}
	}
}

func (e encode) split(index int) (encode, encode) {
	if index < 0 || index >= e.len {
		return e, encode{}
	}
	val := e.val
	rLen := e.len - index - 1
	spliter := getDecimalDivider(rLen)
	return encode{val: val / spliter, len: e.len - rLen}, encode{val: val % spliter, len: rLen}
}

func (e encode) set(index int, setVal MatchType) encode {
	if e.val < 10 {
		return encode{val: uint64(setVal), len: 1}
	}
	l, r := e.split(index)
	if l.val == 0 {
		return e
	}

	val := (l.val/10)*10 + uint64(setVal)
	val = val*getDecimalDivider(r.len) + r.val
	return encode{val: val, len: e.len}
}

func (e encode) append(value MatchType) encode {
	if e.len >= maxPathSegments {
		return e
	}
	return encode{val: e.val*10 + uint64(value), len: e.len + 1}
}

func (e encode) String() string {
	return strconv.Itoa(int(e.val))
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

func newUrlPathEncode(size int) encode {
	switch size {
	case 19:
		return encode{val: 9999999999999999999, len: size}
	case 18:
		return encode{val: 999999999999999999, len: size}
	case 17:
		return encode{val: 99999999999999999, len: size}
	case 16:
		return encode{val: 9999999999999999, len: size}
	case 15:
		return encode{val: 999999999999999, len: size}
	case 14:
		return encode{val: 99999999999999, len: size}
	case 13:
		return encode{val: 9999999999999, len: size}
	case 12:
		return encode{val: 999999999999, len: size}
	case 11:
		return encode{val: 99999999999, len: size}
	case 10:
		return encode{val: 9999999999, len: size}
	case 9:
		return encode{val: 999999999, len: size}
	case 8:
		return encode{val: 99999999, len: size}
	case 7:
		return encode{val: 9999999, len: size}
	case 6:
		return encode{val: 999999, len: size}
	case 5:
		return encode{val: 99999, len: size}
	case 4:
		return encode{val: 9999, len: size}
	case 3:
		return encode{val: 999, len: size}
	case 2:
		return encode{val: 99, len: size}
	case 1:
		return encode{val: 9, len: size}
	default:
		return encode{}
	}
}
