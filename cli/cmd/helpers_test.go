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
			args: []string{"foo/bar"},
			want: []*letterboxd.ListID{
				{
					User: "foo",
					Slug: "bar",
				},
			},
			wantErr: false,
		},
		{
			args:    []string{"foobar"},
			wantErr: true,
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

func TestMustParseListArgs(t *testing.T) {
	tests := []struct {
		args      []string
		want      []*letterboxd.ListID
		wantPanic bool
	}{
		{
			args: []string{"foo/bar"},
			want: []*letterboxd.ListID{
				{
					User: "foo",
					Slug: "bar",
				},
			},
			wantPanic: false,
		},
		{
			args:      []string{"foobar"},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		if tt.wantPanic {
			require.Panics(t, func() { mustParseListArgs(tt.args) })
		} else {
			got := mustParseListArgs(tt.args)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestMatchesGlobOf(t *testing.T) {
	tests := map[string]struct {
		item  string
		globs []string
		want  bool
	}{
		"good": {
			item: "foo", globs: []string{"f*"}, want: true,
		},
		"bad": {
			item: "foo", globs: []string{"b*"}, want: false,
		},
	}
	for k, tt := range tests {
		got := MatchesGlobOf(tt.item, tt.globs)
		require.Equal(t, tt.want, got, k)
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		s []string
		e string
		r bool
	}{
		{s: []string{"a", "b", "c"}, e: "a", r: true},
		{s: []string{"a", "b", "c"}, e: "d", r: false},
	}

	for _, test := range tests {
		got := ContainsString(test.s, test.e)
		require.Equal(t, test.r, got)
	}
}
