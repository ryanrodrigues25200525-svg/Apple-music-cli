package cmd

import (
	"strings"
	"testing"

	"github.com/ryanrodrigues25200525-svg/Apple-music-cli/internal/music"
)

func TestParseSeekDelta(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    float64
		wantErr bool
	}{
		{name: "forward", raw: "+30", want: 30},
		{name: "backward", raw: "-10", want: -10},
		{name: "fractional", raw: "+1.5", want: 1.5},
		{name: "missing sign", raw: "30", wantErr: true},
		{name: "zero", raw: "+0", wantErr: true},
		{name: "invalid", raw: "+soon", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSeekDelta(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseSeekDelta(%q) returned nil error", tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSeekDelta(%q) error = %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("parseSeekDelta(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestFormatPlaylistGroups(t *testing.T) {
	got := formatPlaylistGroups(
		[]music.PlaylistInfo{{Name: "Gym", TrackCount: 12}},
		[]music.PlaylistInfo{{Name: "Chill"}},
	)

	for _, want := range []string{"My Playlists:", "Gym (12 tracks)", "Apple Music & Shared:", "Chill"} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatPlaylistGroups() = %q, want to contain %q", got, want)
		}
	}
}

func TestFormatPlaylistGroupsEmpty(t *testing.T) {
	got := formatPlaylistGroups(nil, nil)
	if got != "No playlists found.\n" {
		t.Fatalf("formatPlaylistGroups(nil, nil) = %q", got)
	}
}
