package letswatch

import (
	"testing"

	"github.com/drewstinnett/go-letterboxd"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	c, err := NewClient(&ClientConfig{
		UseCache: false,
		LetterboxdConfig: &letterboxd.ClientConfig{
			DisableCache: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, c)
}
