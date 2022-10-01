package letswatch

import (
	"context"
	"net/http"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/drewstinnett/go-letterboxd"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/jrudio/go-plex-client"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golift.io/starr"
	"golift.io/starr/radarr"
)

type Client struct {
	// Should this have a ServiceOp?
	LetterboxdClient *letterboxd.Client
	HTTPClient       *http.Client
	Cache            *cache.Cache
	// Service Clients
	Plex      PlexService
	TMDB      TMDBService
	Radarr    RadarrService
	UserAgent string
	Config    *ClientConfig
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

func (c *Client) PruneFilms(films []*letterboxd.Film, popt PruneOpts) ([]*letterboxd.Film, error) {
	var matchesGlob bool
	var err error
	meInfo, err := NewPersonInfoWithViper(viper.GetViper())
	if err != nil {
		return nil, err
	}

	// Only get watched IDs if we need to
	watchedIDs := []string{}
	ret := []*letterboxd.Film{}
	if popt.RemoveWatched {
		log.Info().Msg("Fetching watched in order to prune based on them later")
		watchedIDs, err = c.LetterboxdClient.Film.GetWatchedIMDBIDs(context.TODO(), meInfo.LetterboxdUsername)
		if err != nil {
			return nil, err
		}
	}
	log.Info().Int("unpruned", len(films)).Msg("Film list")
	for _, f := range films {
		slog := log.With().
			Str("film", f.Title).
			Str("imdb", f.ExternalIDs.IMDB).
			Str("tmdb", f.ExternalIDs.TMDB).
			Logger()
		// Are we matching title glob removals?
		if len(popt.RemoveTitleGlobs) > 0 {
			if matches := MatchesGlobOf(f.Title, popt.RemoveTitleGlobs); !matches {
				slog.Debug().Msg("Removing because film matches glob")
				matchesGlob = true
				continue
			}
		}
		if matchesGlob {
			continue
		}

		// Remove Watched films if asked
		if popt.RemoveWatched {
			if ContainsString(watchedIDs, f.ExternalIDs.IMDB) {
				slog.Debug().Msg("Already watched")
				continue
			}
		}

		// Get TMDB stuff
		var m *tmdb.MovieDetails
		if f.ExternalIDs.IMDB != "" {
			m, err = c.TMDB.GetWithIMDBID(context.TODO(), f.ExternalIDs.IMDB)
			if err != nil {
				slog.Warn().Err(err).Msg("Error getting movie from TMDB")
				continue
			}
		} else {
			log.Debug().Err(err).Msg("Movie does not have an IMDB entry. Skipping...")
		}
		// Skip if no TMDB data
		if m == nil {
			log.Warn().Err(err).Msg("No TMDB data for film")
			continue
		}

		// Remove if in my streaming?
		if popt.RemoveMyStreaming {
			streaming, err := c.TMDB.GetStreamingChannels(int(m.ID))
			if err != nil {
				slog.Warn().Err(err).Msg("Error getting streaming channels")
			}

			streamingOnMy := Intersection(meInfo.SubscribedTo, streaming)
			if len(streamingOnMy) != 0 {
				slog.Debug().Strs("streaming", streamingOnMy).Msg("Film is streaming on my channels, skipping")
				continue
			}
		}

		// Do we care about Plex?
		if popt.RemoveMyPlex {
			isAvailOnPlex, err := c.Plex.IsAvailable(context.TODO(), f.Title, f.Year)
			cobra.CheckErr(err)
			if isAvailOnPlex {
				slog.Debug().Msg("Film is available on Plex, skipping")
				continue
			}
		}

		if popt.RemoveMyRadarr {
			results, err := c.Radarr.MoviesWithTMDBID(m.ID)
			cobra.CheckErr(err)
			if len(results) > 0 {
				slog.Debug().Msg("Film already in radarr")
				continue
			}
		}

		// Finally, if still keep...
		ret = append(ret, f)
	}
	return ret, nil
}

type PruneOpts struct {
	RemoveTitleGlobs  []string
	RemoveWatched     bool
	RemoveMyStreaming bool
	RemoveMyPlex      bool
	RemoveMyRadarr    bool
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
	tmdbC, err := NewTMDBClient(config.TMDBKey)
	if err != nil {
		log.Warn().Err(err).Msg("Error initializing tmdb client")
		return nil, err
	}
	// c.TMDBClient = tmdbC
	c.TMDB = &TMDBServiceOp{
		client:     c,
		tmdbClient: tmdbC,
	}

	// Plex Client
	plexC, err := plex.New(config.PlexURL, config.PlexToken)
	if err != nil {
		log.Warn().Err(err).Msg("Error initializing plex client")
		return nil, err
	}
	// c.PlexClient = plexC
	c.Plex = &PlexServiceOp{
		client:     c,
		plexClient: plexC,
	}

	c.LetterboxdClient = letterboxd.NewClient(config.LetterboxdConfig)

	// Radarr Stuff
	sc := starr.New(config.RadarrKey, config.RadarrURL, 0)
	// c.RadarrClient = radarr.New(sc)
	c.Radarr = &RadarrServiceOp{
		client:       c,
		radarrClient: radarr.New(sc),
	}

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

type SupplementOpt struct {
	// Lists []*letterboxd.ListID
}
