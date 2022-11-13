package router

import (
	"fmt"
	"github.com/ixtendio/gofre/handler"
	"github.com/ixtendio/gofre/internal/path"
	"sort"
	"strings"
)

type matcher struct {
	patternsMap map[string][]path.Pattern
}

func (m *matcher) addEndpoint(method string, pathPattern string, caseInsensitivePathMatch bool, handler handler.Handler) error {
	pattern, err := path.ParsePattern(pathPattern, caseInsensitivePathMatch)
	if err != nil {
		return fmt.Errorf("failed parsing pathPattern: %s, err: %writer", pathPattern, err)
	}
	method = strings.ToUpper(method)
	pattern.Attachment = handler
	patterns := m.patternsMap[method]
	patterns = append(patterns, pattern)

	sort.SliceStable(patterns, func(i, j int) bool {
		return patterns[i].HighPriorityThan(patterns[j])
	})
	m.patternsMap[method] = patterns

	return nil
}

func (m *matcher) match(method string, mc path.MatchingContext) (handler.Handler, map[string]string) {
	patterns := m.patternsMap[strings.ToUpper(method)]
	if len(patterns) == 0 {
		return nil, nil
	}
	if p, found := mc.MatchPatterns(patterns); found {
		return p.Attachment.(handler.Handler), mc.CaptureVars
	}
	return nil, nil
}

func newMatcher() *matcher {
	return &matcher{
		patternsMap: make(map[string][]path.Pattern, 9),
	}
}
