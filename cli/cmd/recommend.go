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
		// Set up scrape client for letterboxd
		sc := letterboxd.NewScrapeClient(nil)

		earliest, err := cmd.Flags().GetInt("earliest")
		log.Debugf("earliest: %d", earliest)
		cobra.CheckErr(err)

		language, err := cmd.Flags().GetString("language")
		cobra.CheckErr(err)

		maxRuntime, err := cmd.Flags().GetDuration("max-runtime")
		cobra.CheckErr(err)

		// wantWatches, err := letswatch.ParseRadarrMoviesWithFile(wantWatchF)
		wantWatches, err := sc.List.ListFilms(nil, &letterboxd.ListFilmsOpt{
			User: "dave",
			Slug: "official-top-250-narrative-feature-films",
		})
		cobra.CheckErr(err)

		// Collect watched films first
		// watchedIDs := []string{}
		letterboxdUser, err := cmd.Flags().GetString("letterboxd-user")
		cobra.CheckErr(err)

		log.Info("Getting watched films")
		watchedIDs := []string{}
		watchedFilms, _, err := sc.User.ListWatched(nil, letterboxdUser)
		cobra.CheckErr(err)
		for _, watchedFilm := range watchedFilms {
			watchedIDs = append(watchedIDs, watchedFilm.ExternalIDs.IMDB)
		}
		// watched := []*letswatch.Movie{}

		var possibleRecs []letswatch.Movie
		var recommendations []*letswatch.Movie

		for _, wantWatchMovie := range wantWatches {
			/*
				if wantWatchMovie.ReleaseYear < earliest {
					log.WithFields(log.Fields{
						"film": wantWatchMovie.Title,
					}).Debug("Released too early")
					continue
				}
			*/
			var haveWatched bool
			for _, watchedMovie := range watchedIDs {
				if wantWatchMovie.ExternalIDs.IMDB == watchedMovie {
					log.WithFields(log.Fields{
						"film": wantWatchMovie.Title,
					}).Debug("Already watched")
					haveWatched = true
					break
				}
			}
			if !haveWatched {
				possibleRecs = append(possibleRecs, letswatch.Movie{
					Title:  wantWatchMovie.Title,
					IMDBID: wantWatchMovie.ExternalIDs.IMDB,
					// ReleaseYear: wantWatchMovie.ReleaseYear,
				})
			}
		}

		// Next pass, enhance movies with TMDB data
		for _, possibleRec := range possibleRecs {
			m, err := letswatch.GetMovieWithIMDBID(possibleRec.IMDBID)
			cobra.CheckErr(err)

			// Filter based on TMDB data here
			if language != "" && m.OriginalLanguage != language {
				log.WithFields(log.Fields{
					"film":     m.Title,
					"language": m.OriginalLanguage,
				}).Debug("Wrong language")
				continue
			}
			rt := time.Duration(m.Runtime) * time.Minute
			if maxRuntime != 0 && rt > maxRuntime {
				log.WithFields(log.Fields{
					"film":     m.Title,
					"runtime":  m.Runtime,
					"max-time": maxRuntime,
				}).Debug("Too long")
				continue
			}

			// Add to possibilities
			recommendations = append(recommendations, &letswatch.Movie{
				Title:       m.Title,
				Language:    m.OriginalLanguage,
				ReleaseYear: possibleRec.ReleaseYear,
				IMDBLink:    fmt.Sprintf("https://www.imdb.com/title/%s", m.IMDbID),
				RunTime:     time.Duration(m.Runtime) * time.Minute,
			})

		}

		data, err := yaml.Marshal(recommendations)
		cobra.CheckErr(err)
		fmt.Println(string(data))
		/*
			log.WithFields(log.Fields{
				"film":     rec.Title,
				"released": rec.ReleaseDate,
				"language": rec.OriginalLanguage,
			}).Info("Recommendations")
		*/

		log.WithFields(log.Fields{
			"possible_recs": len(recommendations),
		}).Info("Found Possible Recs")
	},
}

func init() {
	rootCmd.AddCommand(recommendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	recommendCmd.PersistentFlags().String("want-watch", "./testdata/top250.json", "JSON List of Movies to watch")
	// recommendCmd.PersistentFlags().String("watched", "./testdata/watched.json", "JSON List of Movies you have already watched")
	recommendCmd.PersistentFlags().String("letterboxd-user", "mondodrew", "Letterboxd User")
	recommendCmd.PersistentFlags().Int("earliest", 1900, "Earliest release year of a film to recommend")
	recommendCmd.PersistentFlags().String("language", "", "Original language of the movie")
	recommendCmd.PersistentFlags().Duration("max-runtime", 0, "Maximum runtime of a movie to recommend")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// recommendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
