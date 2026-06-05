package lyrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Line is a single lyric line with optional timestamp in seconds.
// Time == 0 for plain (unsynced) lyrics.
type Line struct {
	Time float64
	Text string
}

type lrclibResp struct {
	SyncedLyrics string `json:"syncedLyrics"`
	PlainLyrics  string `json:"plainLyrics"`
}

// Fetch retrieves lyrics from lrclib.net.
// Prefers synced (timestamped) lyrics; falls back to plain.
// Returns nil, nil when the track simply has no lyrics entry.
func Fetch(title, artist, album string, duration float64) ([]Line, error) {
	params := url.Values{}
	params.Set("track_name", title)
	params.Set("artist_name", artist)
	if album != "" {
		params.Set("album_name", album)
	}
	if duration > 0 {
		params.Set("duration", strconv.Itoa(int(duration)))
	}

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Get("https://lrclib.net/api/get?" + params.Encode())
	if err != nil {
		if strings.Contains(err.Error(), "deadline exceeded") || strings.Contains(err.Error(), "timeout") {
			return nil, fmt.Errorf("lyrics service timed out — check your internet connection")
		}
		return nil, fmt.Errorf("could not reach lyrics service")
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var data lrclibResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	if data.SyncedLyrics != "" {
		return parseLRC(data.SyncedLyrics), nil
	}
	if data.PlainLyrics != "" {
		return parsePlain(data.PlainLyrics), nil
	}
	return nil, nil
}

// lrcRe matches lines like: [02:14.53] Some lyric text
var lrcRe = regexp.MustCompile(`^\[(\d+):(\d+)\.(\d+)\]\s*(.*)$`)

func parseLRC(raw string) []Line {
	var lines []Line
	for _, l := range strings.Split(raw, "\n") {
		m := lrcRe.FindStringSubmatch(strings.TrimSpace(l))
		if m == nil {
			continue
		}
		min, _ := strconv.ParseFloat(m[1], 64)
		sec, _ := strconv.ParseFloat(m[2], 64)
		cs, _ := strconv.ParseFloat(m[3], 64)
		t := min*60 + sec + cs/100
		lines = append(lines, Line{Time: t, Text: m[4]})
	}
	sort.Slice(lines, func(i, j int) bool { return lines[i].Time < lines[j].Time })
	return lines
}

func parsePlain(raw string) []Line {
	var lines []Line
	for _, l := range strings.Split(raw, "\n") {
		lines = append(lines, Line{Text: strings.TrimSpace(l)})
	}
	return lines
}

// IsSynced reports whether the lines have timestamps (LRC format).
func IsSynced(lines []Line) bool {
	return len(lines) > 0 && lines[len(lines)-1].Time > 0
}

// ActiveIndex returns the index of the currently playing lyric line.
func ActiveIndex(lines []Line, pos float64) int {
	if len(lines) == 0 || !IsSynced(lines) {
		return 0
	}
	idx := 0
	for i, l := range lines {
		if l.Time <= pos {
			idx = i
		}
	}
	return idx
}
