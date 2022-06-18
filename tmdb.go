package letswatch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/apex/log"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
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

type TMDBClientWithCache struct {
	client *tmdb.Client
	Cache  *cache.Cache
}

func NewTMDBClientWithCache() (*TMDBClientWithCache, error) {
	tc, err := NewTMDBClient()
	if err != nil {
		return nil, err
	}
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"localhost": ":6379",
		},
	})

	t := &TMDBClientWithCache{
		client: tc,
	}
	t.Cache = cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	return t, nil
}

func GetMovieWithIMDBID(imdbID string) (*tmdb.MovieDetails, error) {
	var err error
	client, err := NewTMDBClientWithCache()
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("/letswatch/tmdb/by-imdb-id/%s", imdbID)
	ctx := context.Background()
	var movie *tmdb.MovieDetails

	var inCache bool

	if client.Cache != nil {
		log.WithFields(log.Fields{
			"key": key,
			"ctx": ctx,
		}).Debug("Using cache for lookup")
		if err := client.Cache.Get(ctx, key, &movie); err == nil {
			log.WithField("key", key).Debug("Found page in cache")
			inCache = true
		} else {
			log.WithError(err).WithField("key", key).Debug("TMDB Entry NOT in cache")
		}

	}
	if !inCache {
		options := map[string]string{}
		options["external_source"] = "imdb_id"
		search, err := client.client.GetFindByID(imdbID, options)
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

		movie, err = client.client.GetMovieDetails(int(thing.ID), nil)
		if err != nil {
			return nil, err
		}
		if client.Cache != nil {
			if err := client.Cache.Set(&cache.Item{
				Ctx:   ctx,
				Key:   key,
				Value: movie,
				TTL:   time.Hour * 24,
			}); err != nil {
				log.WithError(err).Warn("Error Writing Cache")
			}
		}
	}
	return movie, nil
}

func GetStreamingChannels(id int) ([]string, error) {
	client, err := NewTMDBClient()
	if err != nil {
		return nil, err
	}
	watchP, err := client.GetMovieWatchProviders(id, nil)
	if err != nil {
		return nil, err
	}
	providers := []string{}
	if val, ok := watchP.MovieWatchProvidersResults.Results["US"]; ok {
		for _, item := range val.Flatrate {
			providers = append(providers, item.ProviderName)
		}
	}
	return providers, nil
}
