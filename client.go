package letswatch

import (
	"net/http"
	"os"

	"github.com/apex/log"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/drewstinnett/go-letterboxd"
	"github.com/go-redis/cache/v8"
	"github.com/jrudio/go-plex-client"
)

type Client struct {
	TMDBClient       *tmdb.Client
	LetterboxdClient *letterboxd.Client
	PlexClient       *plex.Plex
	HTTPClient       *http.Client
	Cache            *cache.Cache
	UserAgent        string
	TMDB             TMDBService
	Plex             PlexService
}

type ClientConfig struct {
	HTTPClient       *http.Client
	UseCache         bool
	Cache            *cache.Cache
	TMDBKey          string
	PlexURL          string
	PlexToken        string
	LetterboxdConfig *letterboxd.ClientConfig
}

func NewClient(config *ClientConfig) (*Client, error) {
	var c *Client

	if config == nil {
		config = &ClientConfig{}
		/*if config.RedisHost == "" {
			log.Fatal("Cache is not disabled and no RedisHost or Cache specified")
		}*/
		/*
			rdb := redis.NewClient(&redis.Options{
				Addr:     "192.168.86.3:6379",
				Password: "",
				DB:       0,
			})

			config = &ClientConfig{
				HTTPClient: http.DefaultClient,
				UseCache:   true,
				Cache: cache.New(&cache.Options{
					Redis:      rdb,
					LocalCache: cache.NewTinyLFU(1000, time.Minute),
				}),
			}
		*/
	}

	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}
	userAgent := "letswatch"
	c = &Client{
		HTTPClient: config.HTTPClient,
		UserAgent:  userAgent,
	}
	// Enable cache if configured
	if config.Cache != nil {
		c.Cache = config.Cache
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

	if config.PlexURL != "" && config.PlexToken != "" {
		log.Debug("Configuring Plex client")
		plexC, err := plex.New(config.PlexURL, config.PlexToken)
		if err != nil {
			return nil, err
		}
		c.PlexClient = plexC
		c.Plex = &PlexServiceOp{client: c}
	}
	// c.TMDBKey = config.TMDBKey
	c.LetterboxdClient = letterboxd.NewClient(config.LetterboxdConfig)

	c.TMDB = &TMDBServiceOp{client: c}
	return c, nil
}
