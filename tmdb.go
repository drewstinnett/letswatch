package letswatch

import (
	"errors"
	"os"

	"github.com/apex/log"
	tmdb "github.com/cyruzin/golang-tmdb"
)

type TMDBMovie struct {
	Title       string `json:"title,omitempty"`
	ReleaseYear int    `json:"release_year,omitempty"`
}

func NewTMDBClient() (*tmdb.Client, error) {
	key := os.Getenv("TMDB_KEY")
	if key == "" {
		return nil, errors.New("ErrNoTMDBKey")
	}
	return tmdb.Init(key)
}

func GetMovieWithIMDBID(imdbID string) (*tmdb.MovieDetails, error) {
	client, err := NewTMDBClient()
	if err != nil {
		return nil, err
	}
	options := map[string]string{}
	options["external_source"] = "imdb_id"
	search, err := client.GetFindByID(imdbID, options)
	if err != nil {
		return nil, err
	}
	if len(search.MovieResults) == 0 {
		return nil, errors.New("ErrNoMovieFound")
	} else if len(search.MovieResults) > 1 {
		log.WithFields(log.Fields{
			"imdb_id": imdbID,
			"count":   len(search.MovieResults),
		}).Warn("Found more than one movie, using the first one")
	}
	thing := search.MovieResults[0]

	details, err := client.GetMovieDetails(int(thing.ID), nil)
	if err != nil {
		return nil, err
	}
	return details, nil
}
