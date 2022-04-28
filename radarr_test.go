package letswatch

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRadarrMoviesWithFile(t *testing.T) {
	movies, err := ParseRadarrMoviesWithFile("testdata/top250.json")
	require.NoError(t, err)
	require.Equal(t, len(movies), 250)
}
