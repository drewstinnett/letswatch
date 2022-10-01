package letswatch

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"

	"github.com/drewstinnett/go-letterboxd"
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
	MovieInputWithLetterboxdFilm(*letterboxd.Film) (*radarr.AddMovieInput, error)
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

func (svc *RadarrServiceOp) MovieInputWithLetterboxdFilm(item *letterboxd.Film) (*radarr.AddMovieInput, error) {
	// Figure out tmdb id in a usable format
	tmdbID, err := strconv.ParseInt(item.ExternalIDs.TMDB, 10, 64)
	if err != nil {
		return nil, err
	}
	profile, err := svc.QualityProfileWithName(svc.client.Config.RadarrQuality)
	if err != nil {
		return nil, err
	}
	tagID := svc.MustAddTag("letswatch-supplement")
	mi := &radarr.AddMovieInput{
		Title:            item.Title,
		Year:             item.Year,
		TmdbID:           tmdbID,
		QualityProfileID: profile.ID,
		RootFolderPath:   svc.client.Config.RadarrPath,
		Monitored:        true,
		Tags:             []int{tagID},
		AddOptions: &radarr.AddMovieOptions{
			SearchForMovie: true,
		},
	}
	return mi, nil
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
