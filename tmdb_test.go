package letswatch

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestGetMovieWithIMDBID(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	find_res, err := ioutil.ReadFile("testdata/find_tmdb.json")
	require.NoError(t, err)

	movie_details, err := ioutil.ReadFile("testdata/movie_details.json")
	require.NoError(t, err)
	httpmock.RegisterResponder("GET", "https://api.themoviedb.org/3/find/tt0111161",
		httpmock.NewStringResponder(200, string(find_res)))
	httpmock.RegisterResponder("GET", "https://api.themoviedb.org/3/movie/290098",
		httpmock.NewStringResponder(200, string(movie_details)))
	os.Setenv("TMDB_KEY", "fake-key")
	movie, err := GetMovieWithIMDBID("tt0111161")
	require.NoError(t, err)
	require.NotNil(t, movie)
	require.Equal(t, movie.Title, "The Handmaiden")
}
