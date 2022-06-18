package cmd

import (
	"errors"
	"strings"

	"github.com/drewstinnett/go-letterboxd"
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
