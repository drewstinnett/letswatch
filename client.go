package letswatch

import (
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/drewstinnett/go-letterboxd"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type Client struct {
	TMDBClient       *tmdb.Client
	LetterboxdClient *letterboxd.Client
	HTTPClient       *http.Client
	Cache            *cache.Cache
	UserAgent        string
	TMDB             TMDBService
}

type ClientConfig struct {
	HTTPClient *http.Client
	UseCache   bool
	Cache      *cache.Cache
	TMDBKey    string
}

func NewClient(config *ClientConfig) (*Client, error) {
	var c *Client

	if config == nil {
		config = &ClientConfig{
			HTTPClient: http.DefaultClient,
		}
	}

	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}
	userAgent := "letswatch"
	c = &Client{
		HTTPClient: config.HTTPClient,
		UserAgent:  userAgent,
	}
	if config.UseCache {
		log.Debug("Configuring local cache inside client")
		if config.Cache != nil {
			c.Cache = config.Cache
		} else {
			ring := redis.NewRing(&redis.RingOptions{
				Addrs: map[string]string{
					"localhost": ":6379",
				},
			})

			c.Cache = cache.New(&cache.Options{
				Redis:      ring,
				LocalCache: cache.NewTinyLFU(1000, time.Minute),
			})

		}
	}

	var tmdbKey string
	if config.TMDBKey != "" {
		tmdbKey = config.TMDBKey
	} else if os.Getenv("TMDB_KEY") != "" {
		tmdbKey = os.Getenv("TMDB_KEY")
	} else {
		log.Warn("No TMDB Key found, skipping that part of the client")
	}
	if tmdbKey != "" {
		tmdbC, err := tmdb.Init(tmdbKey)
		if err != nil {
			return nil, err
		}
		c.TMDBClient = tmdbC
	}
	// c.TMDBKey = config.TMDBKey
	c.LetterboxdClient = letterboxd.NewClient(&letterboxd.ClientConfig{
		UseCache: true,
	})

	c.TMDB = &TMDBServiceOp{client: c}
	return c, nil
}
