package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemVerResolver(t *testing.T) {
	tests := []struct {
		constraint       string
		possibleVersions []string
		expected         string
	}{
		{
			constraint: "^0.5.0",
			possibleVersions: []string{
				"hello",
				"some_version",
				"0.1",
				"0.1-beta",
				"0.5-beta",
			},
			expected: "^0.5.0",
		},
		{
			constraint: "^0.5.0",
			possibleVersions: []string{
				"hello",
				"some_version",
				"0.1",
				"0.1-beta",
				"0.5-beta",
				"0.6",
			},
			expected: "0.6",
		},
		{
			constraint: "~0.4.0",
			possibleVersions: []string{
				"hello",
				"some_version",
				"0.4.0",
				"0.4.0-beta",
				"0.4.1",
				"0.5-beta",
			},
			expected: "0.4.1",
		},
		{
			constraint: "0.4.0",
			possibleVersions: []string{
				"hello",
				"some_version",
				"0.4.0",
				"0.4.0-beta",
				"0.4.1",
				"0.5-beta",
			},
			expected: "0.4.0",
		},
		{
			constraint:       "@",
			possibleVersions: []string{},
			expected:         "@",
		},
	}

	resolver := NewSemVersionResolver()

	for _, test := range tests {
		result := resolver.Resolve(test.constraint, test.possibleVersions)

		assert.Equal(t, test.expected, result)
	}
}
