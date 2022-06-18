package letswatch

import (
	"errors"
	"strings"
	"time"

	"github.com/drewstinnett/go-letterboxd"
	"github.com/spf13/cobra"
)

type Movie struct {
	Title         string        `yaml:"title,omitempty"`
	ReleaseYear   int           `yaml:"release_year,omitempty"`
	IMDBID        string        `yaml:"imdb_id,omitempty"`
	IMDBLink      string        `yaml:"imdb_link,omitempty"`
	TMDBID        string        `yaml:"tmdb_id,omitempty"`
	Language      string        `yaml:"language,omitempty"`
	RunTime       time.Duration `yaml:"runtime,omitempty"`
	StreamingOn   []string      `yaml:"streaming_on,omitempty"`
	StreamingOnMy []string      `yaml:"streaming_on_my,omitempty"`
	Genres        []string      `yaml:"genres,omitempty"`
	Budget        float64       `yaml:"budget,omitempty"`
}

type MovieFilterOpts struct {
	Earliest            int           `yaml:"earliest,omitempty"`
	Latest              int           `yaml:"latest,omitempty"`
	Language            string        `yaml:"language,omitempty"`
	MaxRuntime          time.Duration `yaml:"max_runtime,omitempty"`
	MinRuntime          time.Duration `yaml:"min_runtime,omitempty"`
	IncludeWatched      bool          `yaml:"include_watched,omitempty"`
	IncludeNotStreaming bool          `yaml:"include_not_streaming,omitempty"`
	OnlyMyStreaming     bool          `yaml:"only_my_streaming,omitempty"`
	OnlyNotMyStreaming  bool          `yaml:"only_not_my_streaming,omitempty"`
	Genres              []string      `yaml:"genre,omitempty"`
}

func (m *MovieFilterOpts) ValidateWithPerson(p *PersonInfo) error {
	// If we say just get MY streaming, make sure I actually list what my streaming are
	if m.OnlyMyStreaming && len(p.SubscribedTo) == 0 {
		return errors.New("You must have at least one subscribed to to use only-my-streaming")
	}

	// Inverse of above
	if m.OnlyNotMyStreaming && len(p.SubscribedTo) == 0 {
		return errors.New("You must have at least one subscribed to to use only-not-my-streaming")
	}

	return nil
}

func NewMovieFilterOptsWithCmd(cmd *cobra.Command) (*MovieFilterOpts, error) {
	opts := &MovieFilterOpts{}
	var err error
	opts.Earliest, err = cmd.Flags().GetInt("earliest")
	if err != nil {
		return nil, err
	}

	opts.Language, err = cmd.Flags().GetString("language")
	if err != nil {
		return nil, err
	}
	opts.MaxRuntime, err = cmd.Flags().GetDuration("max-runtime")
	if err != nil {
		return nil, err
	}
	opts.MinRuntime, err = cmd.Flags().GetDuration("min-runtime")
	if err != nil {
		return nil, err
	}

	opts.Genres, err = cmd.Flags().GetStringArray("genre")
	if err != nil {
		return nil, err
	}

	opts.OnlyMyStreaming, err = cmd.Flags().GetBool("only-my-streaming")
	if err != nil {
		return nil, err
	}

	opts.OnlyNotMyStreaming, err = cmd.Flags().GetBool("only-not-my-streaming")
	if err != nil {
		return nil, err
	}

	opts.IncludeWatched, err = cmd.Flags().GetBool("include-watched")
	if err != nil {
		return nil, err
	}

	return opts, nil
}

type MovieCollectOpts struct {
	Watchlist bool                 `yaml:"use_watchlist,omitempty"`
	Lists     []*letterboxd.ListID `yaml:"lists,omitempty"`
}

func NewMovieCollectOptsWithCmd(cmd *cobra.Command) (*MovieCollectOpts, error) {
	opts := &MovieCollectOpts{}
	var err error
	opts.Watchlist, err = cmd.Flags().GetBool("watchlist")
	if err != nil {
		return nil, err
	}

	listArg, err := cmd.Flags().GetStringArray("list")
	if err != nil {
		return nil, err
	}
	opts.Lists, err = parseListArgs(listArg)
	if err != nil {
		return nil, err
	}
	return opts, nil
}

// Given a slice of strings, return a slice of ListIDs
func parseListArgs(args []string) ([]*letterboxd.ListID, error) {
	var ret []*letterboxd.ListID
	for _, argS := range args {
		if !strings.Contains(argS, "/") {
			return nil, errors.New("List Arg must contain a '/' (Example: username/list-slug)")
		}
		parts := strings.Split(argS, "/")
		lid := &letterboxd.ListID{
			User: parts[0],
			Slug: parts[1],
		}
		ret = append(ret, lid)
	}
	return ret, nil
}

func GetFilterMiscWithCmd(cmd *cobra.Command) (*PersonInfo, *MovieFilterOpts, *MovieCollectOpts, error) {
	meInfo, err := NewPersonInfoWithCmd(cmd)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get filters
	movieFilterOpts, err := NewMovieFilterOptsWithCmd(cmd)
	if err != nil {
		return nil, nil, nil, err
	}

	// Collect lists
	movieCollectOpts, err := NewMovieCollectOptsWithCmd(cmd)
	if err != nil {
		return nil, nil, nil, err
	}

	// Error checking
	err = movieFilterOpts.ValidateWithPerson(meInfo)
	if err != nil {
		return nil, nil, nil, err
	}
	return meInfo, movieFilterOpts, movieCollectOpts, nil
}
