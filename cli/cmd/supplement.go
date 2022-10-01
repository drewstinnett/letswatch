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
	"os"
	"strconv"

	"github.com/apex/log"
	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/drewstinnett/go-letterboxd"
	"github.com/drewstinnett/letswatch"
	"github.com/spf13/cobra"
	"golift.io/starr/radarr"
)

// supplementCmd represents the supplement command
var supplementCmd = &cobra.Command{
	Use:   "supplement",
	Short: "Supplement your streaming content with missing films",
	Long:  `Get a list of moves we can't find streaming, and send them in to another API for requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		listA, err := cmd.Flags().GetStringArray("list")
		cobra.CheckErr(err)
		lists, err := parseListArgs(listA)
		cobra.CheckErr(err)

		dryRun, err := cmd.Flags().GetBool("dry-run")
		cobra.CheckErr(err)

		moviesToAdd := []radarr.AddMovieInput{}

		// Quality profiles
		profiles, err := lwc.RadarrClient.GetQualityProfiles()
		cobra.CheckErr(err)
		var qpID int64
		for _, p := range profiles {
			if p.Name == lwc.Config.RadarrQuality {
				qpID = p.ID
			}
		}
		if qpID == 0 {
			log.Fatal("Could not find quality profile")
		}

		// Do a special tag with al lthese
		tag := "letswatch-supplement"
		var tagID int
		tags, err := lwc.RadarrClient.GetTags()
		cobra.CheckErr(err)
		for _, t := range tags {
			if t.Label == tag {
				tagID = t.ID
			}
		}
		if tagID == 0 {
			tagID, err = lwc.RadarrClient.AddTag(tag)
			cobra.CheckErr(err)
		}

		log.Info("Getting watched films")
		wfilmC := make(chan *letterboxd.Film)
		wdoneC := make(chan error)
		meInfo, err := letswatch.NewPersonInfoWithCmd(cmd)
		cobra.CheckErr(err)

		go lwc.LetterboxdClient.User.StreamWatched(nil, meInfo.LetterboxdUsername, wfilmC, wdoneC)

		var watchedIDs []string
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

		var isoFilms []*letterboxd.Film
		isoBatchFilter := &letterboxd.FilmBatchOpts{
			List: lists,
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
					log.WithError(err).Error("Failed to get iso films")
					done <- err
				} else {
					log.Debug("Finished streaming ISO films")
					loop = false
				}
			}
		}

		for _, item := range isoFilms {
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
				m, err = lwc.TMDB.GetWithIMDBID(ctx, item.ExternalIDs.IMDB)
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

			streaming, err := letswatch.GetStreamingChannels(int(m.ID))
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"title": item.Title,
				}).Warn("Error getting streaming channels")
			}

			streamingOnMy := intersection(meInfo.SubscribedTo, streaming)
			if len(streamingOnMy) != 0 {
				log.WithFields(log.Fields{
					"title":     item.Title,
					"streaming": streamingOnMy,
				}).Debug("Film is streaming on my channels, skipping")
				continue
			}

			isAvailOnPlex, err := lwc.Plex.IsAvailable(ctx, item.Title, item.Year)
			cobra.CheckErr(err)
			if isAvailOnPlex {
				log.WithFields(log.Fields{
					"title": item.Title,
				}).Debug("Film is available on Plex, skipping")
				continue
			}

			// Do we have this movie in radarr already?
			results, err := lwc.RadarrClient.GetMovie(m.ID)
			cobra.CheckErr(err)
			if len(results) > 0 {
				log.WithFields(log.Fields{
					"title": item.Title,
				}).Debug("Film already in radarr")
				continue
			}

			stats.TotalItems++

			tmdbID, err := strconv.ParseInt(item.ExternalIDs.TMDB, 10, 64)
			cobra.CheckErr(err)
			mi := radarr.AddMovieInput{
				Title:            item.Title,
				Year:             item.Year,
				TmdbID:           tmdbID,
				QualityProfileID: qpID,
				// RootFolderPath:   os.Getenv("RADARR_PATH"),
				RootFolderPath: lwc.Config.RadarrPath,
				Monitored:      true,
				Tags:           []int{tagID},
				AddOptions: &radarr.AddMovieOptions{
					SearchForMovie: true,
				},
			}
			moviesToAdd = append(moviesToAdd, mi)
		}
		if !dryRun {
			for _, mi := range moviesToAdd {
				fmt.Fprintf(os.Stderr, "%+v\n", mi)
				log.WithField("title", mi.Title).Info("Adding to radarr")
				_, err = lwc.RadarrClient.AddMovie(&mi)
				cobra.CheckErr(err)
			}
		} else {
			for _, mi := range moviesToAdd {
				log.WithFields(log.Fields{
					"movie": mi.Title,
				}).Info("Dry run, not adding to radarr")
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
	supplementCmd.PersistentFlags().StringArray("list", []string{}, "Include the list as part of the recommendations in the format <username>/<list-name>")
	supplementCmd.PersistentFlags().Bool("dry-run", false, "Don't actually add anything to radarr")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// supplementCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
