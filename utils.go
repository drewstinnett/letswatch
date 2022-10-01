package letswatch

import (
	"fmt"
	"os"

	"github.com/gobwas/glob"
	"github.com/rs/zerolog/log"
)

// ValidateEnv takes a list of strings, and returns an error if they do not
// exist as ENV vars. Also returns a map of the values that are valid
func ValidateEnv(envs ...string) (map[string]string, error) {
	var missingEnvs []string
	ret := map[string]string{}
	for _, item := range envs {
		e := os.Getenv(item)
		if e == "" {
			missingEnvs = append(missingEnvs, item)
		} else {
			ret[item] = e
		}
	}
	if len(missingEnvs) > 0 {
		return ret, fmt.Errorf("Missing the following env vars: %v", missingEnvs)
	}
	return ret, nil
}

func inBetween(i, min, max int) bool {
	if (i >= min) && (i <= max) {
		return true
	} else {
		return false
	}
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

func Intersection(s1, s2 []string) (inter []string) {
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
