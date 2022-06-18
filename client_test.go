package letswatch

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	config := &ClientConfig{}
	c, err := NewClient(config)
	require.NoError(t, err)
	require.NotNil(t, c)
}
