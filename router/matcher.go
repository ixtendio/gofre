package router

import (
	"fmt"
	"github.com/ixtendio/gow/internal/path"
	"sort"
	"strings"
)

type trieNode struct {
	parent      *trieNode
	children    []*trieNode
	pathElement *path.Element
	handler     Handler
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
	return n.handler != nil
}

func (n *trieNode) addChild(pathElement *path.Element) (*trieNode, error) {
	if n.isLeaf() {
		return nil, fmt.Errorf("a trie leaf can not have a child")
	}
	for i := 0; i < len(n.children); i++ {
		child := n.children[i]
		if !child.isLeaf() {
			switch pathElement.MatchType {
			case path.MatchSeparatorType:
				if child.pathElement.MatchType == path.MatchSeparatorType {
					return child, nil
				}
			case path.MatchLiteralType:
				if child.pathElement.MatchType == path.MatchLiteralType &&
					child.pathElement.RawVal == pathElement.RawVal {
					return child, nil
				}
			case path.MatchRegexType:
				if child.pathElement.MatchType == path.MatchRegexType &&
					child.pathElement.MatchPattern == pathElement.MatchPattern {
					return child, nil
				}
			case path.MatchVarCaptureType:
				if child.pathElement.MatchType == path.MatchVarCaptureType {
					return child, nil
				}
			case path.MatchVarCaptureWithConstraintType:
				if child.pathElement.MatchType == path.MatchVarCaptureWithConstraintType &&
					child.pathElement.MatchPattern == pathElement.MatchPattern {
					return child, nil
				}
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
		if n.children[i].isLeaf() {
			return fmt.Errorf("the node '%s' contains already a leaf", n.pathElement.RawVal)
		}
	}
	n.children = append(n.children, &trieNode{
		parent:  n,
		handler: data,
	})
	n.sortChildren()
	return nil
}

type matcher struct {
	trieRoots map[string]*trieNode
}

func (m *matcher) addEndpoint(method string, pathPattern string, caseInsensitivePathMatch bool, handler Handler) error {
	pathElements, err := path.Parse(pathPattern, caseInsensitivePathMatch)
	if err != nil {
		return fmt.Errorf("failed parsing pathPattern: %s, err: %w", pathPattern, err)
	}
	method = strings.ToUpper(method)

	rootNode, found := m.trieRoots[method]
	if !found {
		rootNode = &trieNode{pathElement: pathElements[0]}
		m.trieRoots[method] = rootNode
	}
	for i := 1; i < len(pathElements); i++ {
		n, err := rootNode.addChild(pathElements[i])
		if err != nil {
			return err
		}
		rootNode = n
	}
	return rootNode.addLeaf(handler)
}

func (m *matcher) match(method string, mc *path.MatchingContext) Handler {
	method = strings.ToUpper(method)
	pathLen := len(mc.PathElements)
	var matcher func(int, *trieNode) *trieNode
	matcher = func(pathIndex int, root *trieNode) *trieNode {
		if root == nil {
			return nil
		}

		if root.isLeaf() {
			if pathIndex == pathLen {
				return root
			}
			return nil
		}

		if root.pathElement.MatchFunc(pathIndex, mc) {
			for i := 0; i < len(root.children); {
				node := root.children[i]
				h := matcher(pathIndex+1, node)
				if h != nil {
					return h
				}

				if !node.isLeaf() && node.pathElement.MatchType == path.MatchMultiplePathsType {
					if pathIndex == pathLen-1 {
						return nil
					}
					pathIndex++
				} else {
					i++
				}
			}
		}
		return nil
	}
	leaf := matcher(0, m.trieRoots[method])
	if leaf != nil {
		return leaf.handler
	}
	return nil
}

func newMatcher() *matcher {
	return &matcher{trieRoots: make(map[string]*trieNode)}
}
