package letswatch

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/rs/zerolog/log"
	"golift.io/starr/radarr"
)

type RadarrService interface {
	QualityProfiles() ([]*radarr.QualityProfile, error)
	QualityProfileWithName(string) (*radarr.QualityProfile, error)
	AddTag(string) (int, error)
	MustAddTag(string) int
	MoviesWithTMDBID(int64) ([]*radarr.Movie, error)
	AddMovie(*radarr.AddMovieInput) (*radarr.AddMovieOutput, error)
}

type RadarrServiceOp struct {
	client       *Client
	radarrClient *radarr.Radarr
}

type RadarrMovie struct {
	ID          float64 `json:"id,omitempty"`
	Title       string  `json:"title,omitempty"`
	CleanTitle  string  `json:"clean_title,omitempty"`
	IMDBID      string  `json:"imdb_id,omitempty"`
	ReleaseYear int     `json:"release_year,string,omitempty"`
	Adult       bool    `json:"adult,omitempty"`
}

func (svc *RadarrServiceOp) AddMovie(mi *radarr.AddMovieInput) (*radarr.AddMovieOutput, error) {
	return svc.radarrClient.AddMovie(mi)
}

func (svc *RadarrServiceOp) MoviesWithTMDBID(id int64) ([]*radarr.Movie, error) {
	return svc.radarrClient.GetMovie(id)
}

func (svc *RadarrServiceOp) QualityProfileWithName(n string) (*radarr.QualityProfile, error) {
	profiles, err := svc.QualityProfiles()
	if err != nil {
		return nil, err
	}
	for _, p := range profiles {
		if p.Name == n {
			return p, nil
		}
	}
	return nil, errors.New("No profile with that name found")
}

func (svc *RadarrServiceOp) QualityProfiles() ([]*radarr.QualityProfile, error) {
	return svc.radarrClient.GetQualityProfiles()
}

func (svc *RadarrServiceOp) AddTag(t string) (int, error) {
	return svc.radarrClient.AddTag(t)
}

func (svc *RadarrServiceOp) MustAddTag(t string) int {
	i, err := svc.AddTag(t)
	if err != nil {
		log.Warn().Err(err).Str("tag", t).Msg("Error getting tag")
		panic(err)
	}
	return i
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
