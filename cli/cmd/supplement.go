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
	"strconv"

	"github.com/drewstinnett/go-letterboxd"
	"github.com/drewstinnett/letswatch"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golift.io/starr/radarr"
)

var (
	matchGlobs []string
	listsA     []string
	dryRun     bool
)

// supplementCmd represents the supplement command
var supplementCmd = &cobra.Command{
	Use:   "supplement",
	Short: "Supplement your streaming content with missing films",
	Long:  `Get a list of moves we can't find streaming, and send them in to another API for requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		lists := mustParseListArgs(listsA)

		moviesToAdd := []radarr.AddMovieInput{}

		// Quality profiles
		profile, err := lwc.Radarr.QualityProfileWithName(lwc.Config.RadarrQuality)
		cobra.CheckErr(err)

		// Do a special tag with al lthese
		// var tagID int
		tagID := lwc.Radarr.MustAddTag("letswatch-supplement")

		isoBatchFilter := &letterboxd.FilmBatchOpts{
			List: lists,
		}
		isoC := make(chan *letterboxd.Film)
		done := make(chan error)
		go lwc.LetterboxdClient.Film.StreamBatch(ctx, isoBatchFilter, isoC, done)
		isoFilms, err := letterboxd.SlurpFilms(isoC, done)
		cobra.CheckErr(err)

		log.Info().Msg("Pruning film list")
		prunedFilms, err := lwc.PruneFilms(isoFilms, letswatch.PruneOpts{
			RemoveTitleGlobs:  matchGlobs,
			RemoveWatched:     true,
			RemoveMyStreaming: true,
			RemoveMyPlex:      true,
			RemoveMyRadarr:    true,
		})
		cobra.CheckErr(err)

		stats.TotalItems = len(prunedFilms)
		for _, item := range prunedFilms {

			tmdbID, err := strconv.ParseInt(item.ExternalIDs.TMDB, 10, 64)
			cobra.CheckErr(err)
			mi := radarr.AddMovieInput{
				Title:            item.Title,
				Year:             item.Year,
				TmdbID:           tmdbID,
				QualityProfileID: profile.ID,
				RootFolderPath:   lwc.Config.RadarrPath,
				Monitored:        true,
				Tags:             []int{tagID},
				AddOptions: &radarr.AddMovieOptions{
					SearchForMovie: true,
				},
			}
			moviesToAdd = append(moviesToAdd, mi)
		}
		if !dryRun {
			for _, mi := range moviesToAdd {
				log.Info().Str("title", mi.Title).Int("year", mi.Year).Msg("Adding to radarr")
				_, err = lwc.Radarr.AddMovie(&mi)
				cobra.CheckErr(err)
			}
		} else {
			for _, mi := range moviesToAdd {
				log.Info().Str("movie", mi.Title).Int("year", mi.Year).Msg("Dry run, not adding to radarr")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(supplementCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// supplementCmd.PersistentFlags().String("foo", "", "A help for foo")
	supplementCmd.PersistentFlags().StringArrayVar(&listsA, "list", []string{}, "Include the list as part of the recommendations in the format <username>/<list-name>")
	supplementCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Don't actually add anything to radarr")
	supplementCmd.PersistentFlags().StringArrayVar(&matchGlobs, "match-globs", []string{}, "Only recommend movies matching these globs")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// supplementCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
