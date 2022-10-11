package router

import (
	"fmt"
	handler2 "github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/internal/path"
	"sort"
	"strings"
)

type trieNode struct {
	parent          *trieNode
	children        []*trieNode
	pathElement     *path.Element
	captureVarNames []string
	handler         handler2.Handler
}

func (n *trieNode) addCaptureVarNameIfNotExists(varName string) {
	if varName == "" {
		return
	}
	for i := 0; i < len(n.captureVarNames); i++ {
		if n.captureVarNames[i] == varName {
			return
		}
	}
	n.captureVarNames = append(n.captureVarNames, varName)
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

func (n *trieNode) addChild(newPathElement *path.Element) (*trieNode, error) {
	if n.isLeaf() {
		return nil, fmt.Errorf("a trie leaf can not have children")
	}
	pathPatternId := newPathElement.PathPatternId
	captureVarName := newPathElement.CaptureVarName
	for i := 0; i < len(n.children); i++ {
		child := n.children[i]
		if !child.isLeaf() {
			switch newPathElement.MatchType {
			case path.MatchSeparatorType:
				if child.pathElement.MatchType == path.MatchSeparatorType {
					return child, nil
				}
			case path.MatchLiteralType:
				if child.pathElement.MatchType == path.MatchLiteralType &&
					child.pathElement.RawVal == newPathElement.RawVal {
					return child, nil
				}
			case path.MatchRegexType:
				if child.pathElement.MatchType == path.MatchRegexType &&
					child.pathElement.MatchPattern == newPathElement.MatchPattern {
					return child, nil
				}
			case path.MatchVarCaptureType:
				if child.pathElement.MatchType == path.MatchVarCaptureType {
					child.addCaptureVarNameIfNotExists(pathPatternId + "/" + captureVarName)
					return child, nil
				}
			case path.MatchVarCaptureWithConstraintType:
				if child.pathElement.MatchType == path.MatchVarCaptureWithConstraintType &&
					child.pathElement.MatchPattern == newPathElement.MatchPattern {
					child.addCaptureVarNameIfNotExists(pathPatternId + "/" + captureVarName)
					return child, nil
				}
			}
		}
	}
	child := &trieNode{
		parent:      n,
		pathElement: newPathElement,
	}
	if captureVarName != "" {
		child.addCaptureVarNameIfNotExists(pathPatternId + "/" + captureVarName)
	}
	n.children = append(n.children, child)
	n.sortChildren()
	return child, nil
}

func (n *trieNode) addLeaf(data handler2.Handler) error {
	if n.isLeaf() {
		return fmt.Errorf("a trie leaf can not have children (leaf)")
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

func (m *matcher) addEndpoint(method string, pathPattern string, caseInsensitivePathMatch bool, handler handler2.Handler) error {
	pathElementsRoot, err := path.ParsePattern(pathPattern, caseInsensitivePathMatch)
	if err != nil {
		return fmt.Errorf("failed parsing pathPattern: %s, err: %writer", pathPattern, err)
	}
	method = strings.ToUpper(method)

	rootNode, found := m.trieRoots[method]
	if !found {
		rootNode = &trieNode{pathElement: pathElementsRoot}
		m.trieRoots[method] = rootNode
	}

	for nextElement := pathElementsRoot.Next; nextElement != nil; nextElement = nextElement.Next {
		n, err := rootNode.addChild(nextElement)
		if err != nil {
			return err
		}
		rootNode = n
	}
	return rootNode.addLeaf(handler)
}

func (m *matcher) match(method string, mc *path.MatchingContext) (handler2.Handler, map[string]string) {
	allCapturedVars := make(map[string]string)
	method = strings.ToUpper(method)
	pathLen := len(mc.PathElements)
	var matcherFunc func(int, *trieNode) *trieNode
	matcherFunc = func(pathSegmentIndex int, root *trieNode) *trieNode {
		if root == nil {
			return nil
		}

		if root.isLeaf() {
			if pathSegmentIndex == pathLen {
				return root
			}
			return nil
		}

		pathSegment := mc.PathElements[pathSegmentIndex]
		captureVarNames := root.captureVarNames
		matched, varValue := root.pathElement.MatchPathSegment(pathSegment)
		if matched {
			if varValue != "" {
				for _, captureVarName := range captureVarNames {
					allCapturedVars[captureVarName] = varValue
				}
			}
			for i := 0; i < len(root.children); {
				childNode := root.children[i]
				leaf := matcherFunc(pathSegmentIndex+1, childNode)
				if leaf != nil {
					return leaf
				}

				if !childNode.isLeaf() && childNode.pathElement.MatchType == path.MatchMultiplePathsType {
					if pathSegmentIndex == pathLen-1 {
						return nil
					}
					pathSegmentIndex++
				} else {
					i++
				}
			}
		}
		return nil
	}
	leaf := matcherFunc(0, m.trieRoots[method])
	if leaf == nil {
		return nil, nil
	}
	capturedVars := make(map[string]string)
	for node := leaf.parent.pathElement; node != nil; node = node.Previous {
		if node.CaptureVarName != "" {
			captureVal := allCapturedVars[node.PathPatternId+"/"+node.CaptureVarName]
			capturedVars[node.CaptureVarName] = captureVal
		}
	}
	return leaf.handler, capturedVars
}

func newMatcher() *matcher {
	return &matcher{trieRoots: make(map[string]*trieNode)}
}
