package letswatch

import (
	"errors"

	"github.com/rs/zerolog/log"
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
	log.Debug().Str("letterboxd-user", mi.LetterboxdUsername).Msg("Letterboxd User")
	if mi.LetterboxdUsername == "" {
		return nil, errors.New("letterboxd-username is required")
	}

	mi.SubscribedTo = viper.GetStringSlice("subscribed-to")
	log.Debug().Strs("subscribed-to", mi.SubscribedTo).Msg("Subscribed to")
	return mi, nil
}

func NewPersonInfoWithViper(v *viper.Viper) (mi *PersonInfo, err error) {
	mi = &PersonInfo{}
	mi.LetterboxdUsername = v.GetString("letterboxd-username")
	log.Debug().Str("letterboxd-user", mi.LetterboxdUsername).Msg("Letterboxd User")
	if mi.LetterboxdUsername == "" {
		return nil, errors.New("letterboxd-username is required")
	}

	mi.SubscribedTo = v.GetStringSlice("subscribed-to")
	log.Debug().Strs("subscribed-to", mi.SubscribedTo).Msg("Subscribed to")
	return mi, nil
}
