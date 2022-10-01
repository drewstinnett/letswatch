package cmd

import (
	"errors"
	"strings"

	"github.com/drewstinnett/go-letterboxd"
	"github.com/gobwas/glob"
	"github.com/rs/zerolog/log"
)

// Given a slice of strings, return a slice of ListIDs
func parseListArgs(args []string) ([]*letterboxd.ListID, error) {
	var ret []*letterboxd.ListID
	for _, argS := range args {
		if !strings.Contains(argS, "/") {
			return nil, errors.New("List Arg must contain a '/' (Example: username/list-slug)")
		}
		parts := strings.Split(argS, "/")
		lid := &letterboxd.ListID{
			User: parts[0],
			Slug: parts[1],
		}
		ret = append(ret, lid)
	}
	return ret, nil
}

func mustParseListArgs(args []string) []*letterboxd.ListID {
	lists, err := parseListArgs(args)
	if err != nil {
		panic(err)
	}
	return lists
}

// MatchesGlobOf returns true if an item matches any of the given globs
func MatchesGlobOf(item string, globs []string) bool {
	for _, matchGlob := range globs {
		g := glob.MustCompile(matchGlob)
		got := g.Match(item)
		if got {
			return true
		}
	}
	log.Debug().Str("title", item).Msg("Skipping because it matches no globs")
	return false
}

func ContainsString(ss []string, s string) bool {
	for _, item := range ss {
		if s == item {
			return true
		}
	}
	return false
}
