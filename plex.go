package letswatch

import (
	"context"
)

type PlexService interface {
	// GetWithIMDBID(context.Context, string) (*tmdb.MovieDetails, error)
	IsAvailable(context.Context, string, int) (bool, error)
}

type PlexServiceOp struct {
	client *Client
}

func (p *PlexServiceOp) IsAvailable(ctx context.Context, title string, year int) (bool, error) {
	res, err := p.client.PlexClient.Search(title)
	if err != nil {
		return false, err
	}
	for _, d := range res.MediaContainer.Metadata {
		if d.Title == title && d.Year == year {
			return true, nil
		}
	}
	return false, nil
}
