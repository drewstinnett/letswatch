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
	"fmt"
	"time"

	"github.com/apex/log"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/drewstinnett/go-letterboxd"
	"github.com/drewstinnett/letswatch"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// recommendCmd represents the recommend command
var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Recommend a movie to watch!",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// This is what we need to do a proper filter
		meInfo, movieFilterOpts, movieCollectOpts, err := letswatch.GetFilterMiscWithCmd(cmd)
		cobra.CheckErr(err)

		var isoFilms []*letterboxd.Film
		isoBatchFilter := &letterboxd.FilmBatchOpts{}

		if len(movieCollectOpts.Lists) > 0 {
			log.Info("Getting lists")
			isoBatchFilter.List = movieCollectOpts.Lists
		}

		if movieCollectOpts.Watchlist {
			log.Info("Adding Watchlist to ISO")
			isoBatchFilter.WatchList = []string{meInfo.LetterboxdUsername}
		}

		// Collect watched films first
		watchedIDs := []string{}
		if !movieFilterOpts.IncludeWatched {
			log.Info("Getting watched films")
			wfilmC := make(chan *letterboxd.Film)
			wdoneC := make(chan error)
			go sc.User.StreamWatched(nil, meInfo.LetterboxdUsername, wfilmC, wdoneC)

			for loop := true; loop; {
				select {
				case film := <-wfilmC:
					if film.ExternalIDs != nil {
						watchedIDs = append(watchedIDs, film.ExternalIDs.IMDB)
					} else {
						log.WithFields(log.Fields{
							"title": film.Title,
						}).Debugf("No external IDs, skipping")
					}
				case err := <-wdoneC:
					if err != nil {
						log.WithError(err).Error("Failed to get watched films")
						wdoneC <- err
					} else {
						log.Debug("Finished getting watched films")
						loop = false
					}
				}
			}
		}

		isoC := make(chan *letterboxd.Film)
		done := make(chan error)
		go sc.Film.StreamBatch(ctx, isoBatchFilter, isoC, done)

		for loop := true; loop; {
			select {
			case film := <-isoC:
				isoFilms = append(isoFilms, film)
			case err := <-done:
				if err != nil {
					log.WithError(err).Error("Failed to get iso films")
					done <- err
				} else {
					log.Debug("Finished streaming ISO films")
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
			// Do some checking on the y ear
			if (movieFilterOpts.Earliest > 0) && (item.Year < movieFilterOpts.Earliest) {
				log.WithFields(log.Fields{
					"film": item.Title,
				}).Debug("Released too early")
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

			if m == nil {
				log.WithFields(log.Fields{
					"imdbid": item.ExternalIDs.IMDB,
					"tmdbid": item.ExternalIDs.TMDB,
					"title":  item.Title,
				}).Warn("No TMDB data for film")
				continue
			}

			// Filter based on language
			if movieFilterOpts.Language != "" && m.OriginalLanguage != movieFilterOpts.Language {
				log.WithFields(log.Fields{
					"film":     m.Title,
					"language": m.OriginalLanguage,
				}).Debug("Wrong language")
				continue
			}

			rt := time.Duration(m.Runtime) * time.Minute
			if movieFilterOpts.MaxRuntime != 0 && rt > movieFilterOpts.MaxRuntime {
				log.WithFields(log.Fields{
					"film":     m.Title,
					"runtime":  m.Runtime,
					"max-time": movieFilterOpts.MaxRuntime,
				}).Debug("Too long")
				continue
			}
			if movieFilterOpts.MinRuntime != 0 && rt < movieFilterOpts.MinRuntime {
				log.WithFields(log.Fields{
					"film":     m.Title,
					"runtime":  m.Runtime,
					"min-time": movieFilterOpts.MinRuntime,
				}).Debug("Too short")
				continue
			}

			// Ok, looks good, lets find where it's streaming
			streaming, err := letswatch.GetStreamingChannels(int(m.ID))
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"title": item.Title,
				}).Warn("Error getting streaming channels")
			}

			// Just my streaming?
			streamingOnMy := intersection(meInfo.SubscribedTo, streaming)
			if movieFilterOpts.OnlyMyStreaming {
				if len(streamingOnMy) == 0 {
					log.WithFields(log.Fields{
						"film":         m.Title,
						"streaming":    streaming,
						"my-streaming": meInfo.SubscribedTo,
					}).Debug("Film not on any of my streaming subscriptions")
					continue
				}
			} else if movieFilterOpts.OnlyNotMyStreaming {
				if len(streamingOnMy) > 0 {
					log.WithFields(log.Fields{
						"film":         m.Title,
						"streaming":    streaming,
						"my-streaming": meInfo.SubscribedTo,
					}).Debug("Film is on one of my streaming subscriptions")
					continue
				}
			}

			// Add in Genre information
			genres := []string{}
			for _, genre := range m.Genres {
				genres = append(genres, genre.Name)
			}
			if len(movieFilterOpts.Genres) > 0 {
				genreIntersection := intersection(movieFilterOpts.Genres, genres)
				if len(genreIntersection) == 0 {
					log.WithFields(log.Fields{
						"film":       m.Title,
						"genres":     genres,
						"want-genre": movieFilterOpts.Genres,
					}).Debug("Film does not have any of the genres we want")
					continue
				}
			}

			stats.TotalItems++
			rec := &letswatch.Movie{
				Title:         item.Title,
				Language:      m.OriginalLanguage,
				Budget:        float64(m.Budget) / float64(1000000),
				ReleaseYear:   item.Year,
				IMDBLink:      fmt.Sprintf("https://www.imdb.com/title/%s", m.IMDbID),
				RunTime:       time.Duration(m.Runtime) * time.Minute,
				StreamingOn:   streaming,
				StreamingOnMy: streamingOnMy,
				Genres:        genres,
			}
			recL := []*letswatch.Movie{
				rec,
			}
			d, err := yaml.Marshal(recL)
			cobra.CheckErr(err)
			fmt.Print(string(d))
		}
	},
}

func init() {
	rootCmd.AddCommand(recommendCmd)

	// Here you will define your flags and configuration settings.

	// Filter Flags
	recommendCmd.PersistentFlags().Int("earliest", 1900, "Earliest release year of a film to recommend")
	recommendCmd.PersistentFlags().String("language", "", "Original language of the movie")
	recommendCmd.PersistentFlags().Duration("max-runtime", 0, "Maximum runtime of a movie to recommend")
	recommendCmd.PersistentFlags().Duration("min-runtime", 15*time.Minute, "Minimum runtime of a movie to recommend")
	recommendCmd.PersistentFlags().Bool("include-watched", false, "Include films you have watched films the list")
	// recommendCmd.PersistentFlags().Bool("include-not-streaming", true, "Include films that aren't streaming anywhere")
	recommendCmd.PersistentFlags().Bool("only-my-streaming", false, "Only include films that are streaming on your streaming services")
	recommendCmd.PersistentFlags().Bool("only-not-my-streaming", false, "Only include films that are NOT streaming on your streaming services")
	recommendCmd.PersistentFlags().StringArray("genre", []string{}, "Only include films that have this genre")

	// Request Flags
	recommendCmd.PersistentFlags().BoolP("watchlist", "w", false, "Include the users watchlist as part of the recommendations")
	recommendCmd.PersistentFlags().Bool("top250", false, "Include the top 250 narrative films as part of the recommendations")
	recommendCmd.PersistentFlags().StringArray("list", []string{}, "Include the list as part of the recommendations in the format <username>/<list-name>")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// recommendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
