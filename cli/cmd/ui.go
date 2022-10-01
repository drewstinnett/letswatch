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

	"github.com/drewstinnett/go-letterboxd"
	"github.com/drewstinnett/letswatch"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// uiCmd represents the ui command
var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Interactive User Interface for browsing films",
	Run: func(cmd *cobra.Command, args []string) {
		meInfo, movieFilterOpts, movieCollectOpts, err := letswatch.GetFilterMiscWithCmd(cmd)
		cobra.CheckErr(err)
		var isoFilms []*letterboxd.Film
		isoBatchFilter := &letterboxd.FilmBatchOpts{}

		if len(movieCollectOpts.Lists) > 0 {
			log.Info().Msg("Getting lists")
			isoBatchFilter.List = movieCollectOpts.Lists
		}

		if movieCollectOpts.Watchlist {
			log.Info().Msg("Adding Watchlist to ISO")
			isoBatchFilter.WatchList = []string{meInfo.LetterboxdUsername}
		}

		isoC := make(chan *letterboxd.Film)
		done := make(chan error)
		go lwc.LetterboxdClient.Film.StreamBatch(ctx, isoBatchFilter, isoC, done)

		for loop := true; loop; {
			select {
			case film := <-isoC:
				isoFilms = append(isoFilms, film)
			case err := <-done:
				if err != nil {
					log.Error().Err(err).Msg("Failed to get iso films")
					done <- err
				} else {
					log.Debug().Msg("Finished streaming ISO films")
					loop = false
				}
			}
		}

		// Collect watched films first
		watchedIDs := []string{}
		if !movieFilterOpts.IncludeWatched {
			log.Info().Msg("Getting watched films")
			wfilmC := make(chan *letterboxd.Film)
			wdoneC := make(chan error)
			go lwc.LetterboxdClient.User.StreamWatched(nil, meInfo.LetterboxdUsername, wfilmC, wdoneC)

			for loop := true; loop; {
				select {
				case film := <-wfilmC:
					if film.ExternalIDs != nil {
						watchedIDs = append(watchedIDs, film.ExternalIDs.IMDB)
					} else {
						log.Debug().Str("title", film.Title).Msg("No external IDs, skipping")
					}
				case err := <-wdoneC:
					if err != nil {
						log.Error().Err(err).Msg("Failed to get watched films")
						wdoneC <- err
					} else {
						log.Debug().Msg("Finished getting watched films")
						loop = false
					}
				}
			}
		}

		// Convert letterboxd films to letswatch films
		var lwFilms []*letswatch.Movie
		for _, film := range isoFilms {
			var disqualified bool
			for _, watchedId := range watchedIDs {
				if film.ExternalIDs != nil && film.ExternalIDs.IMDB == watchedId {
					disqualified = true
				}
			}

			if disqualified {
				continue
			}
			m, err := lwc.TMDB.GetWithIMDBID(ctx, film.ExternalIDs.IMDB)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get movie")
				continue
			}
			lwFilms = append(lwFilms, &letswatch.Movie{
				Title:       film.Title,
				ReleaseYear: film.Year,
				IMDBLink:    fmt.Sprintf("https://www.imdb.com/title/%s", m.IMDbID),
				Language:    m.OriginalLanguage,
			})
		}

		err = letswatch.NewUI(lwFilms)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uiCmd.PersistentFlags().String("foo", "", "A help for foo")
	uiCmd.PersistentFlags().StringArray("list", []string{}, "Include the list as part of the recommendations in the format <username>/<list-name>")
	uiCmd.PersistentFlags().Bool("include-watched", false, "Include films you have watched films the list")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
