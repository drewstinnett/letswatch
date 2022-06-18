package cmd

import (
	"testing"

	"github.com/drewstinnett/go-letterboxd"
	"github.com/stretchr/testify/require"
)

func TestParseListArgs(t *testing.T) {
	tests := []struct {
		args    []string
		want    []*letterboxd.ListID
		wantErr bool
	}{
		{
			[]string{"foo/bar"},
			[]*letterboxd.ListID{
				{
					User: "foo",
					Slug: "bar",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		got, err := parseListArgs(tt.args)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.Equal(t, tt.want, got)
		}
	}
}
