package letswatch

import (
	"os"
	"testing"

	"github.com/drewstinnett/go-letterboxd"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	os.Clearenv()
}

func shutdown() {
}

func TestNewClient(t *testing.T) {
	c, err := NewClient(ClientConfig{
		UseCache:  false,
		TMDBKey:   "foo",
		PlexURL:   "https://plex.example.com",
		PlexToken: "plex-token",
		LetterboxdConfig: &letterboxd.ClientConfig{
			DisableCache: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestNewClientWithViper(t *testing.T) {
	v := viper.New()
	v.Set("tmdb_key", "foo")
	v.Set("plex_url", "https://plex.example.com")
	v.Set("plex_token", "token")
	v.Set("redis-host", "http://localhost:8888")
	got, err := NewClientWithViper(*v)
	require.NoError(t, err)
	require.NotNil(t, got)
}
