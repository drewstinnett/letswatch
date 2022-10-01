package letswatch

import (
	"net/http"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/drewstinnett/go-letterboxd"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/jrudio/go-plex-client"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

type Client struct {
	TMDBClient       *tmdb.Client
	LetterboxdClient *letterboxd.Client
	PlexClient       *plex.Plex
	HTTPClient       *http.Client
	Cache            *cache.Cache
	RadarrClient     *radarr.Radarr
	TMDB             TMDBService
	Plex             PlexService
	UserAgent        string
	Config           *ClientConfig
}

type ClientConfig struct {
	HTTPClient       *http.Client
	UseCache         bool
	Cache            *cache.Cache
	TMDBKey          string
	PlexURL          string
	PlexToken        string
	RadarrURL        string
	RadarrKey        string
	RadarrQuality    string
	RadarrPath       string
	LetterboxdConfig *letterboxd.ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}
	userAgent := "letswatch"
	c := &Client{
		HTTPClient: config.HTTPClient,
		UserAgent:  userAgent,
	}
	// Enable cache if configured
	if config.Cache != nil {
		c.Cache = config.Cache
	}
	// Setup TMDB
	tmdbC, err := tmdb.Init(config.TMDBKey)
	if err != nil {
		log.Warn().Err(err).Msg("Error initializing tmdb client")
		return nil, err
	}
	c.TMDBClient = tmdbC
	c.TMDB = &TMDBServiceOp{client: c}

	// Plex Client
	plexC, err := plex.New(config.PlexURL, config.PlexToken)
	if err != nil {
		log.Warn().Err(err).Msg("Error initializing plex client")
		return nil, err
	}
	c.PlexClient = plexC
	c.Plex = &PlexServiceOp{client: c}

	c.LetterboxdClient = letterboxd.NewClient(config.LetterboxdConfig)
	// Radarr Stuff
	sc := starr.New(config.RadarrKey, config.RadarrURL, 0)
	c.RadarrClient = radarr.New(sc)

	c.Config = &config
	return c, nil
}

func NewClientWithViper(v viper.Viper) (*Client, error) {
	config := ClientConfig{}
	// var err error

	config.TMDBKey = v.GetString("tmdb_key")
	config.PlexURL = v.GetString("plex_url")
	config.PlexToken = v.GetString("plex_token")
	config.RadarrURL = v.GetString("radarr_url")
	config.RadarrKey = v.GetString("radarr_key")
	config.RadarrQuality = v.GetString("radarr_quality")
	config.RadarrPath = v.GetString("radarr_path")

	if v.GetBool("use_cache") {
		rdb := redis.NewClient(&redis.Options{
			Addr:     v.GetString("redis-host"),
			Password: "",
			DB:       0,
		})
		config.Cache = cache.New(&cache.Options{
			Redis:      rdb,
			LocalCache: cache.NewTinyLFU(1000, time.Minute),
		})
	}

	lbc := &letterboxd.ClientConfig{}
	lbc.RedisHost = v.GetString("redis-host")
	config.LetterboxdConfig = lbc
	// config.RedisHost = viper.GetString("redis-host")
	// config.LetterboxdConfig = lbc
	return NewClient(config)
}

func NewClientWithEnv(config ClientConfig) (*Client, error) {
	vars, err := ValidateEnv(
		"TMDB_KEY",
		"PLEX_URL",
		"PLEX_TOKEN",
	)
	if err != nil {
		return nil, err
	}

	config.TMDBKey = vars["TMDB_KEY"]
	config.PlexURL = vars["PLEX_URL"]
	config.PlexToken = vars["PLEX_TOKEN"]

	return NewClient(config)
}
