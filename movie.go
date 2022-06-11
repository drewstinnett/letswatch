package letswatch

import "time"

type Movie struct {
	Title       string        `yaml:"title,omitempty"`
	ReleaseYear int           `yaml:"release_year,omitempty"`
	IMDBID      string        `yaml:"imdb_id,omitempty"`
	IMDBLink    string        `yaml:"imdb_link,omitempty"`
	TMDBID      string        `yaml:"tmdb_id,omitempty"`
	Language    string        `yaml:"language,omitempty"`
	RunTime     time.Duration `yaml:"runtime,omitempty"`
	StreamingOn []string      `yaml:"streaming_on,omitempty"`
}
