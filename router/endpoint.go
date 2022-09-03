package router

import (
	"fmt"
	"github.com/ixtendio/gow/internal/path"
	"sort"
	"strings"
)

type endpoint struct {
	method   string
	rootPath *path.Element
	handler  Handler
}

type trieNode struct {
	parent      *trieNode
	children    []*trieNode
	pathElement *path.Element
	data        Handler
}

func (n *trieNode) String() string {
	return n.pathElement.String()
}

func (n *trieNode) sortChildren() {
	sort.SliceStable(n.children, func(i, j int) bool {
		ci := n.children[i]
		cj := n.children[j]
		if ci.isLeaf() {
			return true
		}
		if cj.isLeaf() {
			return false
		}
		return ci.pathElement.MatchType < cj.pathElement.MatchType
	})
}

func (n *trieNode) isLeaf() bool {
	return n.data != nil
}

func (n *trieNode) addChild(pathElement *path.Element) (*trieNode, error) {
	if n.isLeaf() {
		return nil, fmt.Errorf("a trie leaf can not have a child")
	}
	switch pathElement.MatchType {
	case path.MatchLiteralType:
		for i := 0; i < len(n.children); i++ {
			child := n.children[i]
			if !child.isLeaf() &&
				child.pathElement.MatchType == path.MatchLiteralType &&
				child.pathElement.RawVal == pathElement.RawVal {
				return child, nil
			}
		}
	case path.MatchRegexType:
		for i := 0; i < len(n.children); i++ {
			child := n.children[i]
			if !child.isLeaf() &&
				child.pathElement.MatchType == path.MatchRegexType &&
				child.pathElement.MatchPattern == pathElement.MatchPattern {
				return child, nil
			}
		}
	case path.MatchVarCaptureType:
		for i := 0; i < len(n.children); i++ {
			child := n.children[i]
			if !child.isLeaf() &&
				child.pathElement.MatchType == path.MatchVarCaptureType {
				return child, nil
			}
		}
	case path.MatchVarCaptureWithConstraintType:
		for i := 0; i < len(n.children); i++ {
			child := n.children[i]
			if !child.isLeaf() &&
				child.pathElement.MatchType == path.MatchVarCaptureWithConstraintType &&
				child.pathElement.MatchPattern == pathElement.MatchPattern {
				return child, nil
			}
		}
	}
	child := &trieNode{
		parent:      n,
		pathElement: pathElement,
	}
	n.children = append(n.children, child)
	n.sortChildren()
	return child, nil
}

func (n *trieNode) addLeaf(data Handler) error {
	if n.isLeaf() {
		return fmt.Errorf("a trie leaf can not have a child (leaf)")
	}
	for i := 0; i < len(n.children); i++ {
		n := n.children[i]
		if n.isLeaf() {
			return fmt.Errorf("the node %s contains already a leaf", n.String())
		}
	}
	n.children = append(n.children, &trieNode{
		parent: n,
		data:   data,
	})
	n.sortChildren()
	return nil
}

type endpointMatcher struct {
	trieRoots map[string]*trieNode
}

func (m *endpointMatcher) addEndpoint(endpoint *endpoint) error {
	method := strings.ToUpper(endpoint.method)
	node, found := m.trieRoots[method]
	if !found {
		node = &trieNode{}
		m.trieRoots[method] = node
	}
	pathElementNode := endpoint.rootPath
	for pathElementNode != nil {
		if pathElementNode.MatchType != path.MatchSeparatorType {
			n, err := node.addChild(pathElementNode)
			if err != nil {
				return err
			}
			node = n
		}
		pathElementNode = pathElementNode.Next
	}
	return node.addLeaf(endpoint.handler)
}

func (m *endpointMatcher) getClosestEndpoint(mc *path.MatchingContext) *endpoint {
	return nil
}

func newEndpointMatcher() *endpointMatcher {
	return &endpointMatcher{trieRoots: make(map[string]*trieNode)}
}
