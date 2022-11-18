package path

import (
	"errors"
	"sort"
	"strings"
)

type node struct {
	val                  string
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
	return n.val
}

type Matcher struct {
	rootPathMatcher *Pattern
	trieRoot        *node
}

func NewMatcher() *Matcher {
	return &Matcher{trieRoot: &node{val: "/"}}
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
		nodeVal := segment.val
		if segment.matchType == MatchTypeCaptureVar {
			nodeVal = "{}"
		} else if segment.matchType == MatchTypeConstraintCaptureVar {
			nodeVal = segment.captureVarPattern.String()
		}

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
			if child.segment.matchType == segment.matchType && child.val == nodeVal {
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

		if !found {
			newNode := &node{
				val:      segment.val,
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
			//we update the maxMatchableSegments field only for the non leafs nodes
			//if !currentNode.isLeaf() {
			newNode.maxMatchableSegments = maxMatchableSegments
			//}
			inserted = true
		}
	}

	if !inserted && currentNode.pattern != nil {
		return errors.New("duplicated pattern detected: '" + pattern.String() + "'")
	}
	return nil
}

func (m *Matcher) Match(mc *MatchingContext) *Pattern {
	p := m.matchPattern(mc)
	if p != nil && p.captureVarsLen > 0 {
		mc.captureVars = make([]CaptureVar, p.captureVarsLen)
		patternSegmentsLen := len(p.segments)
		var psi int
		var captureVarsIndex int
		for i := 0; i < len(mc.pathSegments); i++ {
			urlSegment := &mc.pathSegments[i]
			if urlSegment.matchType == MatchTypeCaptureVar ||
				urlSegment.matchType == MatchTypeConstraintCaptureVar {
				for ; psi < patternSegmentsLen; psi++ {
					patternSegment := p.segments[psi]
					if patternSegment.matchType == MatchTypeCaptureVar ||
						patternSegment.matchType == MatchTypeConstraintCaptureVar {
						mc.captureVars[captureVarsIndex] = CaptureVar{
							Name:  patternSegment.captureVarName,
							Value: urlSegment.val,
						}
						captureVarsIndex++
						psi++
						break
					}
				}
			}
		}
	}
	return p
}

func (m *Matcher) matchPattern(mc *MatchingContext) *Pattern {
	if len(mc.pathSegments) > maxPathSegments {
		return nil
	}
	if len(mc.pathSegments) == 0 && m.rootPathMatcher != nil {
		return m.rootPathMatcher
	}
	type stackSegment struct {
		currentNodeChildren uint16
		urlSegmentIndex     uint8
	}
	var treeDepth int
	urlLen := uint8(len(mc.pathSegments))
	nodeStack := make([]stackSegment, maxPathSegments)
	urlSegmentMatchTypeStack := make([]MatchType, maxPathSegments)
	currentNode := m.trieRoot
	var urlSegmentIndex uint8
	for urlSegmentIndex < urlLen {
		var matched bool
		urlSegment := &mc.pathSegments[urlSegmentIndex]
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
			urlSegmentMatchType := childNode.segment.matchUrlPathSegment(urlSegment)
			if urlSegmentMatchType == MatchTypeMultipleSegments {
				greedyChildren := childNode.children
				greedyChildrenLen := len(greedyChildren)
				if greedyChildrenLen == 0 && childNode.isLeaf() {
					//fill until the end the remained URL segments with MatchTypeMultipleSegments match type
					for i := urlSegmentIndex; i < urlLen; i++ {
						urlSegment := &mc.pathSegments[i]
						urlSegment.matchType = MatchTypeMultipleSegments
					}
					return childNode.pattern
				}
				for urlSegmentIndex < urlLen {
					urlSegment := &mc.pathSegments[urlSegmentIndex]
					urlSegment.matchType = MatchTypeMultipleSegments
					for gci := 0; gci < greedyChildrenLen; gci++ {
						greedyChildNode := greedyChildren[gci]
						if greedyChildNode.segment.matchUrlPathSegment(urlSegment) != MatchTypeUnknown {
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
