package letswatch

import (
	"encoding/json"
	"io/ioutil"
)

type RadarrMovie struct {
	ID          float64 `json:"id,omitempty"`
	Title       string  `json:"title,omitempty"`
	CleanTitle  string  `json:"clean_title,omitempty"`
	IMDBID      string  `json:"imdb_id,omitempty"`
	ReleaseYear int     `json:"release_year,string,omitempty"`
	Adult       bool    `json:"adult,omitempty"`
}

func ParseRadarrMovies(data []byte) ([]RadarrMovie, error) {
	var movies []RadarrMovie
	err := json.Unmarshal(data, &movies)
	return movies, err
}

func ParseRadarrMoviesWithFile(fileP string) ([]RadarrMovie, error) {
	data, err := ioutil.ReadFile(fileP)
	if err != nil {
		return nil, err
	}
	return ParseRadarrMovies(data)
}
