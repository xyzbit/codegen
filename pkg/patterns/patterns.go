package patterns

import (
	"path/filepath"
	"strings"

	"github.com/xyzbit/codegen/pkg/set"
)

// Pattern is a set of patterns.
type Pattern []string

func (p Pattern) Match(list ...string) []string {
	matchTableSet := set.From()
	for _, s := range list {
		for _, v := range p {
			match, _ := filepath.Match(v, filepath.Base(s))
			if match {
				matchTableSet.Add(s)
			}
		}
	}
	return matchTableSet.String()
}

func New(patterns ...string) Pattern {
	patternSet := set.From()
	if len(patterns) == 0 {
		patternSet.Add("*")
		return patternSet.String()
	}

	for _, v := range patterns {
		fields := strings.FieldsFunc(v, func(r rune) bool {
			return r == ','
		})
		for _, f := range fields {
			patternSet.Add(f)
		}
	}

	return patternSet.String()
}
