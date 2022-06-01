package letswatch

import "time"

type Movie struct {
	Title       string        `yaml:"title,omitempty"`
	ReleaseYear int           `yaml:"release_year,omitempty"`
	IMDBID      string        `yaml:"imdb_id,omitempty"`
	Language    string        `yaml:"language,omitempty"`
	IMDBLink    string        `yaml:"imdb_link,omitempty"`
	RunTime     time.Duration `yaml:"runtime,omitempty"`
}
