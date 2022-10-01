package letswatch

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/apex/log"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/go-redis/cache/v8"
)

type TMDBService interface {
	GetWithIMDBID(context.Context, string) (*tmdb.MovieDetails, error)
	GetStreamingChannels(id int) ([]string, error)
}

type TMDBServiceOp struct {
	client     *Client
	tmdbClient *tmdb.Client
}

type TMDBMovie struct {
	Title       string `json:"title,omitempty"`
	ReleaseYear int    `json:"release_year,omitempty"`
}

func NewTMDBClient(key string) (*tmdb.Client, error) {
	if key == "" {
		return nil, errors.New("ErrNoTMDBKey")
	}
	return tmdb.Init(key)
}

type TMDBClientWithCache struct {
	client *tmdb.Client
	Cache  *cache.Cache
}

/*
func (svc *TMDBServiceOp) NewTMDBClientWithCache() (*TMDBClientWithCache, error) {
	tc, err := NewTMDBClient(svc.client.Config.TMDBKey)
	if err != nil {
		return nil, err
	}
	t := &TMDBClientWithCache{
		client: tc,
	}
	var redisH string
	if os.Getenv("REDIS_HOST") != "" {
		redisH = os.Getenv("REDIS_HOST")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:6379", redisH),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	t.Cache = cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	return t, nil
}
*/

func (t *TMDBServiceOp) GetWithIMDBID(ctx context.Context, imdbID string) (*tmdb.MovieDetails, error) {
	key := fmt.Sprintf("/letswatch/tmdb/by-imdb-id/%s", imdbID)
	if ctx == nil {
		ctx = context.Background()
	}
	var movie *tmdb.MovieDetails

	var inCache bool

	if t.client.Cache != nil {
		log.WithFields(log.Fields{
			"key": key,
			"ctx": ctx,
		}).Debug("Using cache for lookup")
		if err := t.client.Cache.Get(ctx, key, &movie); err == nil {
			log.WithField("key", key).Debug("Found page in cache")
			inCache = true
		} else {
			log.WithError(err).WithField("key", key).Debug("TMDB Entry NOT in cache")
		}

	}
	if !inCache {
		options := map[string]string{}
		options["external_source"] = "imdb_id"
		search, err := t.tmdbClient.GetFindByID(imdbID, options)
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

		options = map[string]string{}
		options["append_to_response"] = "credits"
		movie, err = t.tmdbClient.GetMovieDetails(int(thing.ID), options)
		if err != nil {
			return nil, err
		}
		if t.client.Cache != nil {
			if err := t.client.Cache.Set(&cache.Item{
				Ctx:   ctx,
				Key:   key,
				Value: movie,
				TTL:   time.Hour * 24,
			}); err != nil {
				log.WithError(err).Warn("Error Writing TMDB Film to Cache")
			}
		}
	}
	return movie, nil
}

func (svc *TMDBServiceOp) GetStreamingChannels(id int) ([]string, error) {
	watchP, err := svc.tmdbClient.GetMovieWatchProviders(id, nil)
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
