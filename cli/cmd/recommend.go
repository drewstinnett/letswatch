/*
Copyright © 2022 Drew Stinnett <drew@drewlink.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/drewstinnett/letswatch"
	"github.com/drewstinnett/letterrestd/letterboxd"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// recommendCmd represents the recommend command
var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Recommend a movie to watch!",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		// Set up scrape client for letterboxd
		sc := letterboxd.NewScrapeClient(nil)
		ctx := context.Background()

		// What is our letterboxd user?
		letterboxdUser, err := cmd.Flags().GetString("letterboxd-user")
		cobra.CheckErr(err)

		earliest, err := cmd.Flags().GetInt("earliest")
		log.Debugf("earliest: %d", earliest)
		cobra.CheckErr(err)

		language, err := cmd.Flags().GetString("language")
		cobra.CheckErr(err)

		maxRuntime, err := cmd.Flags().GetDuration("max-runtime")
		cobra.CheckErr(err)
		minRuntime, err := cmd.Flags().GetDuration("min-runtime")
		cobra.CheckErr(err)

		useWatchlist, err := cmd.Flags().GetBool("watchlist")
		cobra.CheckErr(err)

		useTop250, err := cmd.Flags().GetBool("top250")
		cobra.CheckErr(err)

		includeWatched, err := cmd.Flags().GetBool("include-watched")
		cobra.CheckErr(err)

		includeNotStreaming, err := cmd.Flags().GetBool("include-not-streaming")
		cobra.CheckErr(err)

		// preWG := &sync.WaitGroup{}
		// ogLock := &sync.Mutex{}

		var isoFilms []*letterboxd.Film
		isoBatchFilter := &letterboxd.FilmBatchOpts{}

		// ogFilmList := []*letterboxd.Film{}
		if useWatchlist {
			log.Info("Adding Watchlist to ISO")
			isoBatchFilter.WatchList = []string{letterboxdUser}
		}

		if useTop250 {
			log.Info("Adding top 250 narrative films")
			isoBatchFilter.Lists = []*letterboxd.ListID{
				{
					User: "dave",
					Slug: "official-top-250-narrative-feature-films",
				},
			}
		}

		// Collect watched films first
		watchedIDs := []string{}
		if !includeWatched {
			log.Info("Getting watched films")
			wfilmC := make(chan *letterboxd.Film)
			wdoneC := make(chan error)
			go sc.User.StreamWatchedWithChan(nil, letterboxdUser, wfilmC, wdoneC)
			loop := true
			for loop {
				select {
				case film := <-wfilmC:
					watchedIDs = append(watchedIDs, film.ExternalIDs.IMDB)
				case err := <-wdoneC:
					if err != nil {
						log.WithError(err).Error("Failed to get watched films")
						wdoneC <- err
					} else {
						log.Info("Finished")
						loop = false
					}
				}
			}
		}
		log.Info("Waiting for collections to complete")

		isoC := make(chan *letterboxd.Film)
		done := make(chan error)
		var recCount int
		go sc.Film.StreamBatchWithChan(ctx, isoBatchFilter, isoC, done)

		loop := true
		for loop {
			select {
			case film := <-isoC:
				isoFilms = append(isoFilms, film)
			case err := <-done:
				if err != nil {
					log.WithError(err).Error("Failed to get iso films")
					done <- err
				} else {
					log.Info("Finished")
					loop = false
				}
			}
		}

		// Now filter the movies down
		for _, item := range isoFilms {
			// Filter watched films if specified
			var disqualified bool
			for _, watchedMovie := range watchedIDs {
				if item.ExternalIDs.IMDB == watchedMovie {
					log.WithFields(log.Fields{
						"film": item.Title,
					}).Debug("Already watched")
					disqualified = true
					// TODO: Should we break further up?
					break
				}
			}
			if disqualified {
				continue
			}

			// Populate with TMDB Data
			var m *tmdb.MovieDetails
			if item.ExternalIDs.IMDB != "" {
				m, err = letswatch.GetMovieWithIMDBID(item.ExternalIDs.IMDB)
				if err != nil {
					log.WithError(err).WithFields(log.Fields{
						"imdbid": item.ExternalIDs.IMDB,
						"tmdbid": item.ExternalIDs.TMDB,
						"title":  item.Title,
					}).Warn("Error getting movie from TMDB")
					continue
				}
			} else {
				log.WithFields(log.Fields{
					"imdbid": item.ExternalIDs.IMDB,
					"tmdbid": item.ExternalIDs.TMDB,
					"title":  item.Title,
				}).Debug("Movie does not have an IMDB entry. Skipping...")
			}

			// Do some checking on the y ear
			var releaseYear int
			if m.ReleaseDate != "" {
				releaseYearS := strings.Split(m.ReleaseDate, "-")[0]
				releaseYear, err = strconv.Atoi(releaseYearS)
				if err != nil {
					log.WithError(err).WithFields(log.Fields{
						"title": item.Title,
					}).Warn("Error parsing release year")
				} else {
					if releaseYear < earliest {
						log.WithFields(log.Fields{
							"film": m.Title,
						}).Debug("Released too early")
						disqualified = true
					}
				}

				// Filter based on language
				if language != "" && m.OriginalLanguage != language {
					log.WithFields(log.Fields{
						"film":     m.Title,
						"language": m.OriginalLanguage,
					}).Debug("Wrong language")
					disqualified = true
					continue
				}
			}

			rt := time.Duration(m.Runtime) * time.Minute
			if maxRuntime != 0 && rt > maxRuntime {
				log.WithFields(log.Fields{
					"film":     m.Title,
					"runtime":  m.Runtime,
					"max-time": maxRuntime,
				}).Debug("Too long")
				disqualified = true
				continue
			}
			if minRuntime != 0 && rt < minRuntime {
				log.WithFields(log.Fields{
					"film":     m.Title,
					"runtime":  m.Runtime,
					"min-time": minRuntime,
				}).Debug("Too short")
				disqualified = true
				continue
			}

			// Ok, looks good, lets find where it's streaming
			streaming, err := letswatch.GetStreamingChannels(int(m.ID))
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"title": item.Title,
				}).Warn("Error getting streaming channels")
			}
			if !includeNotStreaming && len(streaming) == 0 {
				log.WithFields(log.Fields{
					"film":      m.Title,
					"streaming": streaming,
				}).Debug("Not streaming anywhere, skipping")
				disqualified = true
				continue
			}

			if !disqualified {
				recCount++
				rec := &letswatch.Movie{
					Title:       item.Title,
					Language:    m.OriginalLanguage,
					ReleaseYear: releaseYear,
					IMDBLink:    fmt.Sprintf("https://www.imdb.com/title/%s", m.IMDbID),
					RunTime:     time.Duration(m.Runtime) * time.Minute,
					StreamingOn: streaming,
				}
				recL := []*letswatch.Movie{
					rec,
				}
				d, err := yaml.Marshal(recL)
				cobra.CheckErr(err)
				fmt.Println(string(d))
			}
		}

		end := time.Now()

		log.WithFields(log.Fields{
			"count":    recCount,
			"duration": end.Sub(start),
		}).Info("Completed")
	},
}

func init() {
	rootCmd.AddCommand(recommendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// recommendCmd.PersistentFlags().String("watched", "./testdata/watched.json", "JSON List of Movies you have already watched")
	recommendCmd.PersistentFlags().String("letterboxd-user", "mondodrew", "Letterboxd User")
	recommendCmd.PersistentFlags().Int("earliest", 1900, "Earliest release year of a film to recommend")
	recommendCmd.PersistentFlags().String("language", "", "Original language of the movie")
	recommendCmd.PersistentFlags().Duration("max-runtime", 0, "Maximum runtime of a movie to recommend")
	recommendCmd.PersistentFlags().Duration("min-runtime", 15*time.Minute, "Minimum runtime of a movie to recommend")
	recommendCmd.PersistentFlags().Bool("include-watched", false, "Include films you have watched films the list")
	recommendCmd.PersistentFlags().Bool("include-not-streaming", false, "Include films that aren't streaming anywhere")
	recommendCmd.PersistentFlags().BoolP("watchlist", "w", false, "Include the users watchlist as part of the recommendations")
	recommendCmd.PersistentFlags().Bool("top250", false, "Include the top 250 narrative films as part of the recommendations")

	// Replace this please
	recommendCmd.PersistentFlags().String("want-watch", "./testdata/top250.json", "JSON List of Movies to watch")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// recommendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
