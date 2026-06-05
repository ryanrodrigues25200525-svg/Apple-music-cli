package mcp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ryanrodrigues25200525-svg/Apple-music-cli/internal/lyrics"
	"github.com/ryanrodrigues25200525-svg/Apple-music-cli/internal/music"
)

func TestAvailableToolsIncludesExpectedSchemas(t *testing.T) {
	tools := availableTools()
	names := map[string]object{}
	for _, tool := range tools {
		name, ok := tool["name"].(string)
		if !ok || name == "" {
			t.Fatalf("tool has invalid name: %#v", tool)
		}
		names[name] = tool
	}

	for _, name := range []string{
		"play", "pause", "toggle", "next", "previous", "stop",
		"now_playing", "set_volume", "search", "play_song", "play_playlist",
		"lyrics", "queue", "stats", "love", "shuffle", "repeat", "seek",
		"play_playlist_shuffled", "music_context",
	} {
		if _, ok := names[name]; !ok {
			t.Fatalf("missing tool %q in %#v", name, names)
		}
	}

	setVolume := names["set_volume"]
	schema, ok := setVolume["inputSchema"].(object)
	if !ok {
		t.Fatalf("set_volume inputSchema = %#v", setVolume["inputSchema"])
	}
	required, ok := schema["required"].([]string)
	if !ok || len(required) != 1 || required[0] != "volume" {
		t.Fatalf("set_volume required = %#v, want [volume]", schema["required"])
	}
}

func TestExecuteToolValidatesArgumentsBeforeCallingMusic(t *testing.T) {
	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{name: "set_volume", args: nil},
		{name: "search", args: nil},
		{name: "play_song", args: nil},
		{name: "play_playlist", args: nil},
		{name: "play_playlist_shuffled", args: nil},
		{name: "seek", args: nil},
		{name: "unknown", args: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, isErr := executeTool(tt.name, tt.args)
			if !isErr {
				t.Fatalf("executeTool(%q) isErr = false, result %q", tt.name, result)
			}
			if result == "" {
				t.Fatalf("executeTool(%q) returned empty error text", tt.name)
			}
		})
	}
}

func TestExecuteToolReturnsStructuredSearchJSON(t *testing.T) {
	withStubbedMusic(t)

	result, isErr := executeTool("search", map[string]interface{}{"query": "song"})
	if isErr {
		t.Fatalf("executeTool(search) isErr = true, result = %q", result)
	}

	var tracks []music.TrackInfo
	if err := json.Unmarshal([]byte(result), &tracks); err != nil {
		t.Fatalf("search result is not JSON: %v\n%s", err, result)
	}
	if len(tracks) != 1 || tracks[0].Title != "Song" {
		t.Fatalf("tracks = %#v", tracks)
	}
}

func TestExecuteToolMusicContextIncludesAvailableData(t *testing.T) {
	withStubbedMusic(t)

	result, isErr := executeTool("music_context", nil)
	if isErr {
		t.Fatalf("executeTool(music_context) isErr = true, result = %q", result)
	}
	for _, want := range []string{"now_playing", "stats", "queue", "lyrics"} {
		if !strings.Contains(result, want) {
			t.Fatalf("music_context result = %q, want %q", result, want)
		}
	}
}

func withStubbedMusic(t *testing.T) {
	t.Helper()
	original := musicAPI
	musicAPI = musicActions{
		Play:   func() error { return nil },
		Pause:  func() error { return nil },
		Toggle: func() error { return nil },
		Next:   func() error { return nil },
		Prev:   func() error { return nil },
		Stop:   func() error { return nil },
		SetVolume: func(int) error {
			return nil
		},
		NowPlaying: func() (*music.TrackInfo, error) {
			return &music.TrackInfo{
				Title:    "Song",
				Artist:   "Artist",
				Album:    "Album",
				State:    "playing",
				Duration: 180,
			}, nil
		},
		Search: func(string) ([]music.TrackInfo, error) {
			return []music.TrackInfo{{Title: "Song", Artist: "Artist", Album: "Album"}}, nil
		},
		PlayTrackByName:      func(string) error { return nil },
		PlayPlaylist:         func(string) error { return nil },
		PlayPlaylistShuffled: func(string) error { return nil },
		GetQueue: func() ([]music.TrackInfo, error) {
			return []music.TrackInfo{{Title: "Next", Artist: "Artist"}}, nil
		},
		GetTrackStats: func() (*music.TrackStats, error) {
			return &music.TrackStats{Title: "Song", Artist: "Artist", PlayCount: 7, Rating: 80}, nil
		},
		ToggleLove:    func() (bool, error) { return true, nil },
		ToggleShuffle: func() (bool, error) { return true, nil },
		CycleRepeat:   func() (string, error) { return "all", nil },
		Seek:          func(float64) error { return nil },
		FetchLyrics: func(string, string, string, float64) ([]lyrics.Line, error) {
			return []lyrics.Line{{Time: 1.2, Text: "Line"}}, nil
		},
	}
	t.Cleanup(func() { musicAPI = original })
}
