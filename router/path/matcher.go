package path

import (
	"errors"
	"sort"
	"strings"
)

type stackSegment struct {
	currentNodeChildren uint16
	urlSegmentIndex     uint8
}

type node struct {
	maxMatchableSegments uint8
	priority             uint64
	segment              *segment
	pattern              *Pattern
	parent               *node
	children             []*node
}

func (n *node) canMatchPathWithLength(urlPathLen uint8) bool {
	return urlPathLen <= n.maxMatchableSegments
}

func (n *node) isLeaf() bool {
	return n.pattern != nil
}

func (n *node) String() string {
	if n.segment == nil {
		return "/"
	}
	return n.segment.val
}

type Matcher struct {
	caseInsensitive bool
	rootPathMatcher *Pattern
	trieRoot        *node
}

func NewMatcher(caseInsensitive bool) *Matcher {
	return &Matcher{
		caseInsensitive: caseInsensitive,
		trieRoot:        &node{},
	}
}

func (m *Matcher) AddPattern(pattern *Pattern) error {
	if len(pattern.segments) == 0 {
		if m.rootPathMatcher != nil {
			return errors.New("duplicated pattern detected: '/'")
		}
		m.rootPathMatcher = pattern
		return nil
	}

	var inserted bool
	var trieDepth int
	segmentsLength := len(pattern.segments)
	currentNode := m.trieRoot

	for segmentIndex := 0; segmentIndex < segmentsLength; segmentIndex++ {
		segment := pattern.segments[segmentIndex]
		var found bool
		var maxMatchableSegments uint8
		if pattern.isGreedy() {
			maxMatchableSegments = greedyPatternMaxMatchableSegments
		} else {
			maxMatchableSegments = pattern.maxMatchableSegments - uint8(segmentIndex)
		}
		children := currentNode.children
		for i := 0; i < len(children); i++ {
			child := children[i]
			if child.segment.matchType == segment.matchType {
				var hasSameVal bool
				if child.segment.matchType == MatchTypeCaptureVar ||
					child.segment.matchType == MatchTypeSingleSegment ||
					child.segment.matchType == MatchTypeMultipleSegments {
					hasSameVal = true
				} else if child.segment.matchType == MatchTypeConstraintCaptureVar {
					hasSameVal = segment.captureVarPattern.String() == child.segment.val
				} else {
					if m.caseInsensitive {
						hasSameVal = strings.EqualFold(segment.val, child.segment.val)
					} else {
						hasSameVal = segment.val == child.segment.val
					}
				}

				if hasSameVal {
					if child.maxMatchableSegments < maxMatchableSegments {
						child.maxMatchableSegments = maxMatchableSegments
					}
					currentNode = child
					found = true
					if currentNode.parent != nil {
						trieDepth++
					}
					break
				}
			}
		}

		if !found {
			newNode := &node{
				priority: pattern.priority,
				segment:  segment,
				parent:   currentNode,
			}

			// the last node on the branch tree is a leaf
			if segmentIndex == segmentsLength-1 {
				newNode.pattern = pattern
			}
			children = append(children, newNode)
			sort.SliceStable(children, func(i, j int) bool {
				iChild := children[i]
				jChild := children[j]
				if iChild.priority == jChild.priority {
					return strings.Compare(iChild.segment.val, jChild.segment.val) <= 0
				}
				return iChild.priority < jChild.priority
			})
			currentNode.children = children
			currentNode = newNode
			newNode.maxMatchableSegments = maxMatchableSegments
			inserted = true
		} else {
			if segmentIndex == segmentsLength-1 {
				if currentNode.isLeaf() {
					return errors.New("duplicated pattern detected: '" + pattern.String() + "'")
				}
				currentNode.pattern = pattern
				inserted = true
			}
		}
	}

	if !inserted && currentNode.isLeaf() {
		return errors.New("duplicated pattern detected: '" + pattern.String() + "'")
	}
	return nil
}

func (m *Matcher) Match(urlPath string, mc *MatchingContext) *Pattern {
	if len(mc.PathSegments) > MaxPathSegments {
		return nil
	}
	if len(mc.PathSegments) == 0 && m.rootPathMatcher != nil {
		return m.rootPathMatcher
	}
	var treeDepth int
	urlLen := uint8(len(mc.PathSegments))
	nodeStack := make([]stackSegment, MaxPathSegments)
	urlSegmentMatchTypeStack := make([]MatchType, MaxPathSegments)
	currentNode := m.trieRoot
	var urlSegmentIndex uint8
	for urlSegmentIndex < urlLen {
		var matched bool
		urlSegment := &mc.PathSegments[urlSegmentIndex]
		//urlSegmentVal := urlPath[urlSegment.startIndex:urlSegment.endIndex]
		if urlSegmentMatchTypeStack[urlSegmentIndex] != 0 {
			urlSegment.matchType = urlSegmentMatchTypeStack[urlSegmentIndex]
		}

		children := currentNode.children
		childrenLen := uint16(len(children))
		for ci := nodeStack[treeDepth].currentNodeChildren; ci < childrenLen; ci++ {
			childNode := children[ci]
			if !childNode.canMatchPathWithLength(urlLen - urlSegmentIndex) {
				continue
			}
			urlSegmentMatchType := childNode.segment.matchUrlPathSegment(urlPath, urlSegment, m.caseInsensitive)
			if urlSegmentMatchType == MatchTypeMultipleSegments {
				greedyChildren := childNode.children
				greedyChildrenLen := len(greedyChildren)
				if greedyChildrenLen == 0 && childNode.isLeaf() {
					//fill until the end the remained URL segments with MatchTypeMultipleSegments match type
					for i := urlSegmentIndex; i < urlLen; i++ {
						urlSegment := &mc.PathSegments[i]
						urlSegment.matchType = MatchTypeMultipleSegments
					}
					mc.matchedPattern = childNode.pattern
					return childNode.pattern
				}
				for urlSegmentIndex < urlLen {
					urlSegment := &mc.PathSegments[urlSegmentIndex]
					urlSegment.matchType = MatchTypeMultipleSegments
					for gci := 0; gci < greedyChildrenLen; gci++ {
						greedyChildNode := greedyChildren[gci]
						if greedyChildNode.segment.matchUrlPathSegment(urlPath, urlSegment, m.caseInsensitive) != MatchTypeUnknown {
							nodeStack[treeDepth] = stackSegment{
								currentNodeChildren: ci,
								urlSegmentIndex:     urlSegmentIndex + 1,
							}
							urlSegmentMatchTypeStack[urlSegmentIndex] = MatchTypeMultipleSegments
							matched = true
							treeDepth++
							urlSegmentIndex-- //the urlSegmentIndex should not be increased. We have to decrease it here because will be increased at the end of the main loop
							currentNode = childNode
							goto END
						}
					}
					urlSegmentIndex++
				}
				return nil
			} else if urlSegmentMatchType != MatchTypeUnknown {
				urlSegment.matchType = urlSegmentMatchType
				if urlSegmentIndex == urlLen-1 {
					if childNode.isLeaf() {
						mc.matchedPattern = childNode.pattern
						return childNode.pattern
					}
					goto END
				}
				matched = true
				urlSegment.matchType = urlSegmentMatchType
				nodeStack[treeDepth] = stackSegment{
					currentNodeChildren: ci + 1,
					urlSegmentIndex:     urlSegmentIndex,
				}
				treeDepth++
				currentNode = childNode
				goto END
			}
		}

	END:
		if matched {
			if urlSegmentIndex == urlLen-1 && currentNode.isLeaf() {
				mc.matchedPattern = currentNode.pattern
				return currentNode.pattern
			}
			urlSegmentIndex++
		} else {
			if currentNode.parent != nil {
				//stack cleanup
				urlSegmentMatchTypeStack[urlSegmentIndex] = 0
				nodeStack[treeDepth] = stackSegment{}
				treeDepth--
				urlSegmentIndex = nodeStack[treeDepth].urlSegmentIndex
				currentNode = currentNode.parent
			} else {
				break
			}
		}
	}
	return nil
}
