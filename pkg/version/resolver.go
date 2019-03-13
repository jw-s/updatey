package version

import (
	"sort"

	"github.com/Masterminds/semver"
)

// Resolver determines what version to use.
type Resolver interface {
	Resolve(string, []string) string
}

type semVersionResolver struct{}

// NewSemVersionResolver is a resolver which uses the Semantic Version spec.
func NewSemVersionResolver() Resolver {
	return semVersionResolver{}
}

// Resolve determines the version to return based on a Semantic version constraint and a list of versions which may meet the constaint.
// For more information about defining constaints: https://github.com/Masterminds/semver#checking-version-constraints
func (semVersionResolver) Resolve(constraint string, versions []string) string {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return constraint
	}

	var compatibles []*semver.Version

	for _, version := range versions {
		v, err := semver.NewVersion(version)
		if err != nil {
			continue
		}

		if c.Check(v) {
			compatibles = append(compatibles, v)
		}
	}

	sort.Sort(sort.Reverse(semver.Collection(compatibles)))
	if len(compatibles) == 0 {
		return constraint
	}

	return compatibles[0].Original()
}
