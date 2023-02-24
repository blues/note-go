package note

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWordsFromString(t *testing.T) {
	cases := map[string]string{
		"dev:foobar":    "near-eat-read",
		"dev:123456778": "farm-quiet-dumb",
		"dev:qwerty":    "flour-water-stock",
	}

	for k, v := range cases {
		require.Equal(t, v, WordsFromString(k))
	}
}
