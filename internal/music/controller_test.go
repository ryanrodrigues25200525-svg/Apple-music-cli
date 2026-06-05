package music

import (
	"reflect"
	"testing"
)

func TestEscapeAS(t *testing.T) {
	got := escapeAS(`a "quoted" \ path`)
	want := `a \"quoted\" \\ path`
	if got != want {
		t.Fatalf("escapeAS() = %q, want %q", got, want)
	}
}

func TestParseTrackLines(t *testing.T) {
	input := "\nSong|Artist|Album\nbad line\nOther|Another|Album | Deluxe\n"

	got := parseTrackLines(input)
	want := []TrackInfo{
		{Title: "Song", Artist: "Artist", Album: "Album"},
		{Title: "Other", Artist: "Another", Album: "Album | Deluxe"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseTrackLines() = %#v, want %#v", got, want)
	}
}

func TestParseRecordTrackLines(t *testing.T) {
	input := "Song\x1fArtist\x1fAlbum\nEmoji\x1f\x1fSingle\n"

	got := parseRecordTrackLines(input, recordSep)
	want := []TrackInfo{
		{Title: "Song", Artist: "Artist", Album: "Album"},
		{Title: "Emoji", Artist: "", Album: "Single"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseRecordTrackLines() = %#v, want %#v", got, want)
	}
}

func TestParsePlaylistTrackOutput(t *testing.T) {
	if _, err := parsePlaylistTrackOutput("NOT_FOUND"); err == nil {
		t.Fatal("parsePlaylistTrackOutput(NOT_FOUND) error = nil, want error")
	}

	if _, err := parsePlaylistTrackOutput("ERROR" + recordSep + "bad playlist"); err == nil {
		t.Fatal("parsePlaylistTrackOutput(ERROR) error = nil, want error")
	}
}

func TestPlayPlaylistTrackAtIndexRejectsNegativeIndex(t *testing.T) {
	if err := PlayPlaylistTrackAtIndexByPersistentID("playlist", -1); err == nil {
		t.Fatal("PlayPlaylistTrackAtIndexByPersistentID negative index error = nil, want error")
	}
}
