package letswatch

import (
	"fmt"
	"os"
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
