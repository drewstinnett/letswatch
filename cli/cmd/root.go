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
	"os"
	"time"

	"github.com/drewstinnett/letswatch"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	Verbose bool
	err     error
	start   = time.Now()
	stats   *runStats
	// sc      *letterboxd.Client
	ctx context.Context
	lwc *letswatch.Client
)

type runStats struct {
	TotalItems int           `json:"total_items"`
	Duration   time.Duration `json:"duration"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "letswatch",
	Short: "Pick something to watch!",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		stats = &runStats{}
		// config := &letswatch.ClientConfig{}
		// config.LetterboxdConfig = lbc
		// lwc, err = letswatch.NewClient(*config)
		lwc, err = letswatch.NewClientWithViper(*viper.GetViper())
		cobra.CheckErr(err)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		end := time.Now()
		stats.Duration = end.Sub(start)
		log.Info().Int("total_items", stats.TotalItems).Str("duration", fmt.Sprint(stats.Duration)).Msg("Run stats")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.letswatch.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Verbose logging")
	rootCmd.PersistentFlags().String("letterboxd-username", "", "My Letterboxd Username")
	viper.BindPFlag("letterboxd-username", rootCmd.PersistentFlags().Lookup("letterboxd-username"))
	rootCmd.PersistentFlags().StringArray("subscribed-to", []string{}, "Streaming services that you are subscribed to")
	viper.BindPFlag("subscribed-to", rootCmd.PersistentFlags().Lookup("subscribed-to"))
	rootCmd.PersistentFlags().String("redis-host", "localhost:6379", "URL for Redis cluster")
	viper.BindPFlag("redis-host", rootCmd.PersistentFlags().Lookup("redis-host"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigName(".letswatch")
	}

	viper.AutomaticEnv() // read in environment variables that match
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug().Str("config-file", viper.ConfigFileUsed()).Msg("Using config file")
	}
}

// Remove dups from slice.
func removeDups(elements []string) (nodups []string) {
	encountered := make(map[string]bool)
	for _, element := range elements {
		if !encountered[element] {
			nodups = append(nodups, element)
			encountered[element] = true
		}
	}
	return
}

func intersection(s1, s2 []string) (inter []string) {
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[e] = true
	}
	for _, e := range s2 {
		// If elements present in the hashmap then append intersection list.
		if hash[e] {
			inter = append(inter, e)
		}
	}
	// Remove dups from slice.
	inter = removeDups(inter)
	return
}
