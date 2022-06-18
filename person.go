package letswatch

import (
	"errors"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type PersonInfo struct {
	LetterboxdUsername string
	SubscribedTo       []string
}

func NewPersonInfoWithCmd(cmd *cobra.Command) (*PersonInfo, error) {
	mi := &PersonInfo{}
	mi.LetterboxdUsername = viper.GetString("letterboxd-username")
	log.WithFields(log.Fields{
		"letterboxd-user": mi.LetterboxdUsername,
	}).Debug("Letterboxd User")
	if mi.LetterboxdUsername == "" {
		return nil, errors.New("letterboxd-username is required")
	}

	mi.SubscribedTo = viper.GetStringSlice("subscribed-to")
	log.WithFields(log.Fields{
		"subscribed-to": mi.SubscribedTo,
	}).Debug("Subscribed to")
	return mi, nil
}
