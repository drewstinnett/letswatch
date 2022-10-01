package letswatch

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateEnvs(t *testing.T) {
	tests := map[string]struct {
		requiredEnv []string
		env         map[string]string
		wantErr     string
	}{
		"good": {
			requiredEnv: []string{"FOO"},
			env: map[string]string{
				"FOO": "bar",
			},
			wantErr: "",
		},
		"bad": {
			requiredEnv: []string{"FOO"},
			env: map[string]string{
				"BAZ": "bar",
			},
			wantErr: "Missing the following env vars: [FOO]",
		},
	}
	for k, tt := range tests {
		os.Clearenv()
		for e, v := range tt.env {
			t.Setenv(e, v)
		}
		_, err := ValidateEnv(tt.requiredEnv...)
		if tt.wantErr != "" {
			require.Error(t, err, k)
			require.EqualError(t, err, tt.wantErr)
		} else {
			require.NoError(t, err, k)
		}
	}
}
