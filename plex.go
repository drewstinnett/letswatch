package letswatch

import (
	"context"

	"github.com/jrudio/go-plex-client"
)

type PlexService interface {
	// GetWithIMDBID(context.Context, string) (*tmdb.MovieDetails, error)
	IsAvailable(context.Context, string, int) (bool, error)
}

type PlexServiceOp struct {
	client     *Client
	plexClient *plex.Plex
}

func (p *PlexServiceOp) IsAvailable(ctx context.Context, title string, year int) (bool, error) {
	res, err := p.plexClient.Search(title)
	if err != nil {
		return false, err
	}
	padding := 2
	earliest := year - padding
	latest := year + padding
	// fmt.Fprintf(os.Stderr, "%+v\n", res.MediaContainer.Metadata)
	for _, d := range res.MediaContainer.Metadata {
		if d.Title == title && inBetween(d.Year, earliest, latest) {
			return true, nil
		}
	}
	return false, nil
}
