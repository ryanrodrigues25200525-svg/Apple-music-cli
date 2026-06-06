package music

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TrackInfo holds details about the currently playing track.
type TrackInfo struct {
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Album    string  `json:"album"`
	State    string  `json:"state"`
	Volume   int     `json:"volume"`
	Position float64 `json:"position"`
	Duration float64 `json:"duration"`
	Shuffle  bool    `json:"shuffle"`
	Repeat   string  `json:"repeat"`
	Loved    bool    `json:"loved"`
}

// TrackStats holds extended metadata for the current track.
type TrackStats struct {
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	PlayCount int    `json:"play_count"`
	Rating    int    `json:"rating"` // 0–100
	Loved     bool   `json:"loved"`
	DateAdded string `json:"date_added"`
}

// PlaylistInfo identifies a Music.app playlist without relying on its display name.
type PlaylistInfo struct {
	Name         string `json:"name"`
	ID           int    `json:"id"`
	PersistentID string `json:"persistent_id"`
	Kind         string `json:"kind"`
	TrackCount   int    `json:"track_count"`
}

const recordSep = "\x1f"

// RunAppleScript executes an AppleScript snippet and returns trimmed stdout.
func RunAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("applescript error: %w (output: %s)", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// IsMusicRunning reports whether Music.app is currently running.
func IsMusicRunning() (bool, error) {
	cmd := exec.Command("pgrep", "-x", "Music")
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// EnsureMusicRunning launches Music.app if it is not already running.
func EnsureMusicRunning() error {
	running, err := IsMusicRunning()
	if err != nil {
		return err
	}
	if !running {
		_, err = RunAppleScript(`tell application "Music" to activate`)
	}
	return err
}

func Play() error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	_, err := RunAppleScript(`tell application "Music" to play`)
	return err
}

func Pause() error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	_, err := RunAppleScript(`tell application "Music" to pause`)
	return err
}

func Toggle() error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	_, err := RunAppleScript(`tell application "Music" to playpause`)
	return err
}

func Next() error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	return advanceTrack("next")
}

func Prev() error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	return advanceTrack("previous")
}

func advanceTrack(direction string) error {
	nativeCommand := "next track"
	step := 1
	wrapIndex := "1"
	if direction == "previous" {
		nativeCommand = "previous track"
		step = -1
		wrapIndex = "total"
	}

	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				if not (exists current track) then return "NO_TRACK"
				set curPID to persistent ID of current track
				try
					%s
					delay 0.15
					if (exists current track) and persistent ID of current track is not curPID then return "PLAYING"
				end try

				try
					set p to container of current track
					set total to count of tracks of p
					if total is 0 then return "NO_TRACK"
					repeat with i from 1 to total
						try
							if persistent ID of track i of p is curPID then
								set targetIndex to i + (%d)
								if targetIndex > total then set targetIndex to %s
								if targetIndex < 1 then set targetIndex to %s
								set targetTrack to item targetIndex of tracks of p
								set targetPID to persistent ID of targetTrack
								repeat with attempt from 1 to 4
									play targetTrack
									delay 0.25
									if (exists current track) and persistent ID of current track is targetPID then return "PLAYING"
								end repeat
								return "NO_CHANGE"
							end if
						end try
					end repeat
				end try

				try
					set p to library playlist 1
					set total to count of tracks of p
					if total is 0 then return "NO_TRACK"
					repeat with i from 1 to total
						try
							if persistent ID of track i of p is curPID then
								set targetIndex to i + (%d)
								if targetIndex > total then set targetIndex to %s
								if targetIndex < 1 then set targetIndex to %s
								set targetTrack to item targetIndex of tracks of p
								set targetPID to persistent ID of targetTrack
								repeat with attempt from 1 to 4
									play targetTrack
									delay 0.25
									if (exists current track) and persistent ID of current track is targetPID then return "PLAYING"
								end repeat
								return "NO_CHANGE"
							end if
						end try
					end repeat
				end try

				return "NO_CHANGE"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, nativeCommand, step, wrapIndex, wrapIndex, step, wrapIndex, wrapIndex))
	if err != nil {
		return err
	}
	switch {
	case out == "PLAYING":
		return nil
	case out == "NO_TRACK":
		return fmt.Errorf("no current track")
	case out == "NO_CHANGE":
		return fmt.Errorf("could not advance track")
	case strings.HasPrefix(out, "ERROR|"):
		return fmt.Errorf("applescript error: %s", strings.TrimPrefix(out, "ERROR|"))
	default:
		return nil
	}
}

func Stop() error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	_, err := RunAppleScript(`tell application "Music" to stop`)
	return err
}

// GetVolume returns Music.app's own volume (0–100), not the system volume.
func GetVolume() (int, error) {
	if err := EnsureMusicRunning(); err != nil {
		return 0, err
	}
	out, err := RunAppleScript(`tell application "Music" to get sound volume`)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(out)
}

// SetVolume sets Music.app's volume (0–100), not the system volume.
func SetVolume(vol int) error {
	if vol < 0 {
		vol = 0
	} else if vol > 100 {
		vol = 100
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	_, err := RunAppleScript(fmt.Sprintf(`tell application "Music" to set sound volume to %d`, vol))
	return err
}

func ToggleShuffle() (bool, error) {
	if err := EnsureMusicRunning(); err != nil {
		return false, err
	}
	out, err := RunAppleScript(`
		tell application "Music"
			set shuffle enabled to not shuffle enabled
			return shuffle enabled
		end tell
	`)
	if err != nil {
		return false, err
	}
	return out == "true", nil
}

func CycleRepeat() (string, error) {
	if err := EnsureMusicRunning(); err != nil {
		return "", err
	}
	return RunAppleScript(`
		tell application "Music"
			set cur to song repeat as string
			if cur is "off" then
				set song repeat to all
				return "all"
			else if cur is "all" or cur is "yes" then
				set song repeat to one
				return "one"
			else
				set song repeat to off
				return "off"
			end if
		end tell
	`)
}

func Seek(delta float64) error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	_, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			set cur to player position
			set newPos to cur + (%f)
			if newPos < 0 then set player position to 0
			else set player position to newPos
		end tell
	`, delta))
	return err
}

// SetPlayerPosition jumps to an absolute timestamp within the current song.
func SetPlayerPosition(seconds float64) error {
	if seconds < 0 {
		seconds = 0
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			if not (exists current track) then return "NO_TRACK"
			set player position to %f
		end tell
	`, seconds))
	if err != nil {
		return err
	}
	if out == "NO_TRACK" {
		return fmt.Errorf("no current track")
	}
	return nil
}

// FadeOut gradually reduces Music.app volume to zero over the requested duration.
func FadeOut(seconds float64) error {
	if seconds <= 0 {
		seconds = 5
	}
	startVolume, err := GetVolume()
	if err != nil {
		return err
	}
	steps := 20
	if seconds < 2 {
		steps = 10
	}
	delay := time.Duration(seconds * float64(time.Second) / float64(steps))
	for i := 1; i <= steps; i++ {
		nextVolume := startVolume - int(float64(startVolume)*float64(i)/float64(steps))
		if err := SetVolume(nextVolume); err != nil {
			return err
		}
		if i < steps {
			time.Sleep(delay)
		}
	}
	return nil
}

// SetShuffleMode enables or disables shuffle mode.
func SetShuffleMode(enabled bool) error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	_, err := RunAppleScript(fmt.Sprintf(`tell application "Music" to set shuffle enabled to %t`, enabled))
	return err
}

// SetRepeatMode sets repeat mode to off, one, or all.
func SetRepeatMode(mode string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "off" && mode != "one" && mode != "all" {
		return fmt.Errorf("repeat mode must be off, one, or all")
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	_, err := RunAppleScript(fmt.Sprintf(`tell application "Music" to set song repeat to %s`, mode))
	return err
}

// ToggleLove toggles the loved state of the current track and returns the new state.
// Falls back gracefully for streaming tracks that may not support the loved property.
func ToggleLove() (bool, error) {
	if err := EnsureMusicRunning(); err != nil {
		return false, err
	}
	out, err := RunAppleScript(`
		tell application "Music"
			try
				set loved of current track to not loved of current track
				return loved of current track
			on error
				return "UNSUPPORTED"
			end try
		end tell
	`)
	if err != nil {
		return false, err
	}
	if out == "UNSUPPORTED" {
		return false, fmt.Errorf("loved is not supported for this track")
	}
	return out == "true", nil
}

// SetLove sets the loved state of the current track and returns the new state.
func SetLove(loved bool) (bool, error) {
	if err := EnsureMusicRunning(); err != nil {
		return false, err
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set loved of current track to %t
				return loved of current track
			on error
				return "UNSUPPORTED"
			end try
		end tell
	`, loved))
	if err != nil {
		return false, err
	}
	if out == "UNSUPPORTED" {
		return false, fmt.Errorf("loved is not supported for this track")
	}
	return out == "true", nil
}

// DislikeCurrentTrack marks the current track as disliked when Music.app exposes that property.
func DislikeCurrentTrack() error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	out, err := RunAppleScript(`
		tell application "Music"
			try
				set loved of current track to false
			end try
			try
				set disliked of current track to true
				return "OK"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`)
	if err != nil {
		return err
	}
	if strings.HasPrefix(out, "ERROR|") {
		return fmt.Errorf("disliked is not supported for this track: %s", strings.TrimPrefix(out, "ERROR|"))
	}
	return nil
}

// RateCurrentTrack sets the current track rating. Stars are 1-5 and map to Music.app's 20-100 scale.
func RateCurrentTrack(stars int) error {
	if stars < 1 || stars > 5 {
		return fmt.Errorf("rating must be 1-5 stars")
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	_, err := RunAppleScript(fmt.Sprintf(`tell application "Music" to set rating of current track to %d`, stars*20))
	return err
}

// NowPlaying returns full playback state including loved status.
func NowPlaying() (*TrackInfo, error) {
	running, err := IsMusicRunning()
	if err != nil {
		return nil, err
	}
	if !running {
		return &TrackInfo{State: "stopped"}, nil
	}

	out, err := RunAppleScript(`
		tell application "Music"
			try
				set tVolume to sound volume
				if player state is stopped then return "STOPPED|" & tVolume
				if not (exists current track) then return "STOPPED|" & tVolume
				set t to current track
				set tName    to name of t
				set tArtist  to artist of t
				set tAlbum   to album of t
				set tPos     to player position
				set tDur     to duration of t
				set pState   to player state as string
				set tShuffle to shuffle enabled
				set tRepeat  to song repeat as string
				set tLoved   to false
				try
					set tLoved to loved of t
				end try
				return tName & "|" & tArtist & "|" & tAlbum & "|" & pState & "|" & tVolume & "|" & tPos & "|" & tDur & "|" & tShuffle & "|" & tRepeat & "|" & tLoved
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(out, "|")
	if len(parts) == 0 {
		return nil, errors.New("empty response from AppleScript")
	}
	switch parts[0] {
	case "ERROR":
		return nil, fmt.Errorf("applescript: %s", parts[1])
	case "STOPPED":
		vol := 0
		if len(parts) > 1 {
			vol, _ = strconv.Atoi(parts[1])
		}
		return &TrackInfo{State: "stopped", Volume: vol}, nil
	}
	if len(parts) < 7 {
		return nil, fmt.Errorf("unexpected output: %s", out)
	}

	vol, _ := strconv.Atoi(parts[4])
	pos, _ := strconv.ParseFloat(parts[5], 64)
	dur, _ := strconv.ParseFloat(parts[6], 64)

	state := strings.ToLower(parts[3])
	switch {
	case strings.Contains(state, "play"):
		state = "playing"
	case strings.Contains(state, "paus"):
		state = "paused"
	default:
		state = "stopped"
	}

	shuffle := len(parts) > 7 && parts[7] == "true"
	repeat := "off"
	if len(parts) > 8 {
		repeat = strings.ToLower(parts[8])
	}
	loved := len(parts) > 9 && parts[9] == "true"

	return &TrackInfo{
		Title: parts[0], Artist: parts[1], Album: parts[2],
		State: state, Volume: vol, Position: pos, Duration: dur,
		Shuffle: shuffle, Repeat: repeat, Loved: loved,
	}, nil
}

// GetTrackStats returns play count, rating, loved, and date added for the current track.
func GetTrackStats() (*TrackStats, error) {
	if err := EnsureMusicRunning(); err != nil {
		return nil, err
	}
	out, err := RunAppleScript(`
		tell application "Music"
			try
				if not (exists current track) then return "NO_TRACK"
				set t to current track
				set tCount to played count of t
				set tRating to rating of t
				set tAdded to ""
				try
					set tAdded to date added of t as string
				end try
				set tLoved to false
				try
					set tLoved to loved of t
				end try
				return (name of t) & "|" & (artist of t) & "|" & (album of t) & "|" & tCount & "|" & tRating & "|" & tLoved & "|" & tAdded
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`)
	if err != nil {
		return nil, err
	}
	if out == "NO_TRACK" {
		return nil, fmt.Errorf("no track playing")
	}
	if strings.HasPrefix(out, "ERROR") {
		parts := strings.SplitN(out, "|", 2)
		return nil, fmt.Errorf("applescript: %s", parts[1])
	}

	parts := strings.Split(out, "|")
	if len(parts) < 7 {
		return nil, fmt.Errorf("unexpected output: %s", out)
	}
	count, _ := strconv.Atoi(parts[3])
	rating, _ := strconv.Atoi(parts[4])
	return &TrackStats{
		Title:     parts[0],
		Artist:    parts[1],
		Album:     parts[2],
		PlayCount: count,
		Rating:    rating,
		Loved:     parts[5] == "true",
		DateAdded: parts[6],
	}, nil
}

// GetQueue returns up to 10 tracks following the current track in its playlist.
func GetQueue() ([]TrackInfo, error) {
	if err := EnsureMusicRunning(); err != nil {
		return nil, err
	}
	out, err := RunAppleScript(`
		tell application "Music"
			try
				set curPlaylist to container of current track
				set curPID to persistent ID of current track
				set total to count of tracks of curPlaylist
				set output to ""
				set curIdx to 0
				repeat with i from 1 to total
					try
						if persistent ID of track i of curPlaylist is curPID then
							set curIdx to i
							exit repeat
						end if
					end try
				end repeat
				if curIdx is 0 then return "ERROR|current track not found in its playlist"
				set emitted to 0
				set i to curIdx + 1
				repeat while emitted < 10 and total > 0
					if i > total then set i to 1
					if i is curIdx then exit repeat
					set t to item i of tracks of curPlaylist
					set output to output & name of t & "|" & artist of t & "|" & album of t & "\n"
					set emitted to emitted + 1
					set i to i + 1
				end repeat
				return output
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(out, "ERROR|") {
		return nil, fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR|"))
	}
	return parseTrackLines(out), nil
}

// GetArtworkPath saves the current track's artwork to /tmp/muse_art.jpg and returns the path.
// Returns empty string when no artwork is available.
func GetArtworkPath() (string, error) {
	if err := EnsureMusicRunning(); err != nil {
		return "", err
	}
	out, err := RunAppleScript(`
		tell application "Music"
			try
				if not (exists current track) then return "NO_ART"
				set artList to artworks of current track
				if (count of artList) = 0 then return "NO_ART"
				set artData to raw data of item 1 of artList
				set tmpPath to "/tmp/muse_art.jpg"
				set fp to open for access POSIX file tmpPath with write permission
				set eof fp to 0
				write artData to fp
				close access fp
				return tmpPath
			on error
				return "NO_ART"
			end try
		end tell
	`)
	if err != nil || out == "NO_ART" || out == "" {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Search queries the library for tracks matching query in title, artist, or album.
func Search(query string) ([]TrackInfo, error) {
	if err := EnsureMusicRunning(); err != nil {
		return nil, err
	}
	q := escapeAS(query)
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			set output to ""
			try
				set trackList to (every track of library playlist 1 whose name contains "%s" or artist contains "%s" or album contains "%s")
				set total to count of trackList
				if total > 50 then set total to 50
				repeat with i from 1 to total
					set t to item i of trackList
					set output to output & name of t & "|" & artist of t & "|" & album of t & "\n"
				end repeat
			on error errText
				return "ERROR|" & errText
			end try
			return output
		end tell
	`, q, q, q))
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(out, "ERROR|") {
		return nil, fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR|"))
	}
	return parseTrackLines(out), nil
}

// PlayTrackByName plays the first track whose title, artist, or album contains name.
func PlayTrackByName(name string) error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	n := escapeAS(name)
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set targetTrack to missing value
				set t1 to (every track of library playlist 1 whose name contains "%s")
				if (count of t1) > 0 then
					set targetTrack to item 1 of t1
				else
					set t2 to (every track of library playlist 1 whose artist contains "%s")
					if (count of t2) > 0 then
						set targetTrack to item 1 of t2
					else
						set t3 to (every track of library playlist 1 whose album contains "%s")
						if (count of t3) > 0 then
							set targetTrack to item 1 of t3
						end if
					end if
				end if
				if targetTrack is missing value then return "NOT_FOUND"
				set targetPID to persistent ID of targetTrack
				repeat with attempt from 1 to 4
					play targetTrack
					delay 0.25
					if (exists current track) and persistent ID of current track is targetPID then return "PLAYING"
				end repeat
				return "NO_CHANGE"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, n, n, n))
	if err != nil {
		return err
	}
	if out == "NOT_FOUND" {
		return fmt.Errorf("no track found matching: %s", name)
	}
	if out == "NO_CHANGE" {
		return fmt.Errorf("could not play track: %s", name)
	}
	if strings.HasPrefix(out, "ERROR") {
		return fmt.Errorf("applescript error: %s", out)
	}
	return nil
}

// PlayPlaylist plays the named playlist from the beginning.
func PlayPlaylist(name string) error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	n := escapeAS(name)
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				if exists playlist "%s" then
					play playlist "%s"
					return "PLAYING"
				else
					return "NOT_FOUND"
				end if
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, n, n))
	if err != nil {
		return err
	}
	if out == "NOT_FOUND" {
		return fmt.Errorf("no playlist found: %s", name)
	}
	if strings.HasPrefix(out, "ERROR") {
		return fmt.Errorf("applescript error: %s", out)
	}
	return nil
}

// PlayPlaylistByPersistentID plays a specific playlist from the beginning.
func PlayPlaylistByPersistentID(persistentID string) error {
	return playPlaylistByPersistentID(persistentID, false)
}

// GetPlaylists returns the names of all playlists (used by MCP).
func GetPlaylists() ([]string, error) {
	mine, others, err := GetCategorizedPlaylistInfos()
	if err != nil {
		return nil, err
	}
	all := append(mine, others...)
	names := make([]string, 0, len(all))
	for _, p := range all {
		names = append(names, p.Name)
	}
	return names, nil
}

// GetCategorizedPlaylists returns playlists split into user-created and subscription/shared.
func GetCategorizedPlaylists() (mine []string, others []string, err error) {
	mineInfo, othersInfo, err := GetCategorizedPlaylistInfos()
	if err != nil {
		return nil, nil, err
	}
	for _, p := range mineInfo {
		mine = append(mine, p.Name)
	}
	for _, p := range othersInfo {
		others = append(others, p.Name)
	}
	return mine, others, nil
}

// GetCategorizedPlaylistInfos returns playable playlists split into user/library
// playlists and Apple Music/shared playlists. Folder playlists without direct
// tracks are skipped because opening them as track lists produces empty screens.
func GetCategorizedPlaylistInfos() (mine []PlaylistInfo, others []PlaylistInfo, err error) {
	if err = EnsureMusicRunning(); err != nil {
		return
	}
	out, err := RunAppleScript(`
		tell application "Music"
			set sep to ASCII character 31
			set nl to ASCII character 10
			set output to ""
			try
				repeat with p in every playlist
					try
						set pClass to class of p as string
						set pName to name of p
						set pID to id of p
						set pPersistentID to persistent ID of p
						set pTrackCount to 0
						try
							set pTrackCount to count of tracks of p
						end try

						if pClass is "folder playlist" and pTrackCount is 0 then
							-- skip empty folders; they are containers, not playable track lists
						else if pClass is "user playlist" or pClass is "library playlist" or pClass is "folder playlist" then
							set output to output & "MINE" & sep & pID & sep & pPersistentID & sep & pClass & sep & pTrackCount & sep & pName & nl
						else
							set output to output & "OTHER" & sep & pID & sep & pPersistentID & sep & pClass & sep & pTrackCount & sep & pName & nl
						end if
					end try
				end repeat
			on error errText
				return "ERROR" & sep & errText
			end try
			return output
		end tell
	`)
	if err != nil {
		return
	}
	if strings.HasPrefix(out, "ERROR"+recordSep) {
		return nil, nil, fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR"+recordSep))
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, recordSep, 6)
		if len(parts) < 6 {
			continue
		}
		id, _ := strconv.Atoi(parts[1])
		trackCount, _ := strconv.Atoi(parts[4])
		info := PlaylistInfo{
			ID:           id,
			PersistentID: parts[2],
			Kind:         parts[3],
			TrackCount:   trackCount,
			Name:         parts[5],
		}
		if parts[0] == "MINE" {
			mine = append(mine, info)
		} else if parts[0] == "OTHER" {
			others = append(others, info)
		}
	}
	return
}

// GetPlaylistTracks returns up to 100 tracks from the named playlist.
func GetPlaylistTracks(playlistName string) ([]TrackInfo, error) {
	if err := EnsureMusicRunning(); err != nil {
		return nil, err
	}
	n := escapeAS(playlistName)
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set matches to (every playlist whose name is "%s")
				if (count of matches) is 0 then return "NOT_FOUND"
				set p to item 1 of matches
				%s
			on error errText
				return "ERROR" & (ASCII character 31) & errText
			end try
		end tell
	`, n, playlistTracksAppleScriptBody()))
	if err != nil {
		return nil, err
	}
	return parsePlaylistTrackOutput(out)
}

// CreatePlaylist creates a new user playlist.
func CreatePlaylist(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("playlist name is required")
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				if exists playlist "%s" then return "EXISTS"
				make new user playlist with properties {name:"%s"}
				return "OK"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, escapeAS(name), escapeAS(name)))
	if err != nil {
		return err
	}
	if out == "EXISTS" {
		return fmt.Errorf("playlist already exists: %s", name)
	}
	if strings.HasPrefix(out, "ERROR|") {
		return fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR|"))
	}
	return nil
}

// AddTrackToPlaylist duplicates the first matching library track into a playlist.
func AddTrackToPlaylist(playlistName, query string) error {
	playlistName = strings.TrimSpace(playlistName)
	query = strings.TrimSpace(query)
	if playlistName == "" || query == "" {
		return fmt.Errorf("playlist name and query are required")
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	q := escapeAS(query)
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				if not (exists playlist "%s") then return "PLAYLIST_NOT_FOUND"
				set targetTrack to missing value
				set matches to (every track of library playlist 1 whose name contains "%s" or artist contains "%s" or album contains "%s")
				if (count of matches) > 0 then set targetTrack to item 1 of matches
				if targetTrack is missing value then return "TRACK_NOT_FOUND"
				duplicate targetTrack to playlist "%s"
				return "OK"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, escapeAS(playlistName), q, q, q, escapeAS(playlistName)))
	if err != nil {
		return err
	}
	switch out {
	case "PLAYLIST_NOT_FOUND":
		return fmt.Errorf("playlist not found: %s", playlistName)
	case "TRACK_NOT_FOUND":
		return fmt.Errorf("no library track found matching: %s", query)
	}
	if strings.HasPrefix(out, "ERROR|") {
		return fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR|"))
	}
	return nil
}

// RemoveTrackFromPlaylist removes a matching track from a playlist. If index is positive,
// it removes that 1-based position; otherwise it removes the first track matching query.
func RemoveTrackFromPlaylist(playlistName, query string, index int) error {
	playlistName = strings.TrimSpace(playlistName)
	query = strings.TrimSpace(query)
	if playlistName == "" {
		return fmt.Errorf("playlist name is required")
	}
	if index <= 0 && query == "" {
		return fmt.Errorf("query or 1-based index is required")
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	var selector string
	if index > 0 {
		selector = fmt.Sprintf(`
				if %d > (count of tracks of p) then return "TRACK_NOT_FOUND"
				delete track %d of p
		`, index, index)
	} else {
		q := escapeAS(query)
		selector = fmt.Sprintf(`
				set matches to (every track of p whose name contains "%s" or artist contains "%s" or album contains "%s")
				if (count of matches) is 0 then return "TRACK_NOT_FOUND"
				delete item 1 of matches
		`, q, q, q)
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set matches to (every playlist whose name is "%s")
				if (count of matches) is 0 then return "PLAYLIST_NOT_FOUND"
				set p to item 1 of matches
				%s
				return "OK"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, escapeAS(playlistName), selector))
	if err != nil {
		return err
	}
	switch out {
	case "PLAYLIST_NOT_FOUND":
		return fmt.Errorf("playlist not found: %s", playlistName)
	case "TRACK_NOT_FOUND":
		return fmt.Errorf("playlist track not found")
	}
	if strings.HasPrefix(out, "ERROR|") {
		return fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR|"))
	}
	return nil
}

// GetPlaylistTracksByPersistentID returns up to 100 tracks from a specific playlist.
func GetPlaylistTracksByPersistentID(persistentID string) ([]TrackInfo, error) {
	if err := EnsureMusicRunning(); err != nil {
		return nil, err
	}
	pid := escapeAS(persistentID)
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set matches to (every playlist whose persistent ID is "%s")
				if (count of matches) is 0 then return "NOT_FOUND"
				set p to item 1 of matches
				%s
			on error errText
				return "ERROR" & (ASCII character 31) & errText
			end try
		end tell
	`, pid, playlistTracksAppleScriptBody()))
	if err != nil {
		return nil, err
	}
	return parsePlaylistTrackOutput(out)
}

func playlistTracksAppleScriptBody() string {
	return `
				set sep to ASCII character 31
				set nl to ASCII character 10
				set total to count of tracks of p
				if total > 100 then set total to 100
				set output to ""
				repeat with i from 1 to total
					set t to track i of p
					set tName to ""
					set tArtist to ""
					set tAlbum to ""
					try
						set tName to name of t
					end try
					try
						set tArtist to artist of t
					end try
					try
						set tAlbum to album of t
					end try
					if tName is not "" then
						set output to output & tName & sep & tArtist & sep & tAlbum & nl
					end if
				end repeat
				return output
	`
}

func parsePlaylistTrackOutput(out string) ([]TrackInfo, error) {
	switch {
	case out == "NOT_FOUND":
		return nil, fmt.Errorf("playlist not found")
	case strings.HasPrefix(out, "ERROR"+recordSep):
		return nil, fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR"+recordSep))
	default:
		return parseRecordTrackLines(out, recordSep), nil
	}
}

// PlayPlaylistShuffled enables shuffle and plays the named playlist.
func PlayPlaylistShuffled(playlistName string) error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	n := escapeAS(playlistName)
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set shuffle enabled to true
				play playlist "%s"
				return "PLAYING"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, n))
	if err != nil {
		return err
	}
	if strings.HasPrefix(out, "ERROR") {
		return fmt.Errorf("applescript error: %s", out)
	}
	return nil
}

// PlayPlaylistShuffledByPersistentID enables shuffle and plays a specific playlist.
func PlayPlaylistShuffledByPersistentID(persistentID string) error {
	return playPlaylistByPersistentID(persistentID, true)
}

func playPlaylistByPersistentID(persistentID string, shuffled bool) error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	pid := escapeAS(persistentID)
	shuffleLine := ""
	if shuffled {
		shuffleLine = "set shuffle enabled to true"
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set matches to (every playlist whose persistent ID is "%s")
				if (count of matches) is 0 then return "NOT_FOUND"
				%s
				play item 1 of matches
				return "PLAYING"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, pid, shuffleLine))
	if err != nil {
		return err
	}
	if out == "NOT_FOUND" {
		return fmt.Errorf("playlist not found")
	}
	if strings.HasPrefix(out, "ERROR") {
		return fmt.Errorf("applescript error: %s", out)
	}
	return nil
}

// PlayPlaylistTrack plays a specific track by title within a playlist.
func PlayPlaylistTrack(playlistName, trackTitle string) error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set targetTrack to (first track of playlist "%s" whose name is "%s")
				set targetPID to persistent ID of targetTrack
				repeat with attempt from 1 to 4
					play targetTrack
					delay 0.25
					if (exists current track) and persistent ID of current track is targetPID then return "PLAYING"
				end repeat
				return "NO_CHANGE"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, escapeAS(playlistName), escapeAS(trackTitle)))
	if err != nil {
		return err
	}
	if out == "NO_CHANGE" {
		return fmt.Errorf("could not play playlist track")
	}
	if strings.HasPrefix(out, "ERROR") {
		return fmt.Errorf("applescript error: %s", out)
	}
	return nil
}

// PlayPlaylistTrackByPersistentID plays a specific track by title within a playlist.
func PlayPlaylistTrackByPersistentID(persistentID, trackTitle string) error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set matches to (every playlist whose persistent ID is "%s")
				if (count of matches) is 0 then return "NOT_FOUND"
				set p to item 1 of matches
				set targetTrack to (first track of p whose name is "%s")
				set targetPID to persistent ID of targetTrack
				repeat with attempt from 1 to 4
					play targetTrack
					delay 0.25
					if (exists current track) and persistent ID of current track is targetPID then return "PLAYING"
				end repeat
				return "NO_CHANGE"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, escapeAS(persistentID), escapeAS(trackTitle)))
	if err != nil {
		return err
	}
	if out == "NOT_FOUND" {
		return fmt.Errorf("playlist not found")
	}
	if out == "NO_CHANGE" {
		return fmt.Errorf("could not play playlist track")
	}
	if strings.HasPrefix(out, "ERROR") {
		return fmt.Errorf("applescript error: %s", out)
	}
	return nil
}

// PlayPlaylistTrackAtIndexByPersistentID plays a specific 0-based track index
// within a playlist. This is preferred by the TUI because duplicate track names
// are common in real libraries.
func PlayPlaylistTrackAtIndexByPersistentID(persistentID string, index int) error {
	if index < 0 {
		return fmt.Errorf("track index must be non-negative")
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set matches to (every playlist whose persistent ID is "%s")
				if (count of matches) is 0 then return "NOT_FOUND"
				set p to item 1 of matches
				set oneBasedIndex to %d
				if oneBasedIndex > (count of tracks of p) then return "NOT_FOUND"
				set targetTrack to item oneBasedIndex of tracks of p
				set targetPID to persistent ID of targetTrack
				repeat with attempt from 1 to 4
					play targetTrack
					delay 0.25
					if (exists current track) and persistent ID of current track is targetPID then return "PLAYING"
				end repeat
				return "NO_CHANGE"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, escapeAS(persistentID), index+1))
	if err != nil {
		return err
	}
	if out == "NOT_FOUND" {
		return fmt.Errorf("playlist track not found")
	}
	if out == "NO_CHANGE" {
		return fmt.Errorf("could not play playlist track")
	}
	if strings.HasPrefix(out, "ERROR") {
		return fmt.Errorf("applescript error: %s", out)
	}
	return nil
}

// PlayPlaylistTrackAtIndex plays a specific 1-based track index within a named playlist.
func PlayPlaylistTrackAtIndex(playlistName string, index int) error {
	if strings.TrimSpace(playlistName) == "" {
		return fmt.Errorf("playlist name is required")
	}
	if index <= 0 {
		return fmt.Errorf("track index must be 1 or greater")
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set matches to (every playlist whose name is "%s")
				if (count of matches) is 0 then return "PLAYLIST_NOT_FOUND"
				set p to item 1 of matches
				if %d > (count of tracks of p) then return "TRACK_NOT_FOUND"
				set targetTrack to item %d of tracks of p
				set targetPID to persistent ID of targetTrack
				repeat with attempt from 1 to 4
					play targetTrack
					delay 0.25
					if (exists current track) and persistent ID of current track is targetPID then return "PLAYING"
				end repeat
				return "NO_CHANGE"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, escapeAS(playlistName), index, index))
	if err != nil {
		return err
	}
	switch out {
	case "PLAYLIST_NOT_FOUND":
		return fmt.Errorf("playlist not found: %s", playlistName)
	case "TRACK_NOT_FOUND":
		return fmt.Errorf("playlist track not found")
	case "NO_CHANGE":
		return fmt.Errorf("could not play playlist track")
	}
	if strings.HasPrefix(out, "ERROR|") {
		return fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR|"))
	}
	return nil
}

// AddCurrentTrackToLibrary duplicates the current track into the library playlist when supported.
func AddCurrentTrackToLibrary() error {
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	out, err := RunAppleScript(`
		tell application "Music"
			try
				if not (exists current track) then return "NO_TRACK"
				duplicate current track to library playlist 1
				return "OK"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`)
	if err != nil {
		return err
	}
	if out == "NO_TRACK" {
		return fmt.Errorf("no current track")
	}
	if strings.HasPrefix(out, "ERROR|") {
		return fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR|"))
	}
	return nil
}

// GetRecentlyPlayed returns tracks from Music.app's Recently Played playlist when present.
func GetRecentlyPlayed(limit int) ([]TrackInfo, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	if err := EnsureMusicRunning(); err != nil {
		return nil, err
	}
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set matches to (every playlist whose name is "Recently Played")
				if (count of matches) is 0 then return "NOT_FOUND"
				set p to item 1 of matches
				set sep to ASCII character 31
				set nl to ASCII character 10
				set output to ""
				set total to count of tracks of p
				if total > %d then set total to %d
				repeat with i from 1 to total
					set t to track i of p
					set output to output & name of t & sep & artist of t & sep & album of t & nl
				end repeat
				return output
			on error errText
				return "ERROR" & (ASCII character 31) & errText
			end try
		end tell
	`, limit, limit))
	if err != nil {
		return nil, err
	}
	return parsePlaylistTrackOutput(out)
}

// GetTopTracks returns the most-played library tracks.
func GetTopTracks(limit int) ([]TrackStats, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	if err := EnsureMusicRunning(); err != nil {
		return nil, err
	}
	out, err := RunAppleScript(`
		tell application "Music"
			set sep to ASCII character 31
			set nl to ASCII character 10
			set output to ""
			try
				repeat with t in every track of library playlist 1
					try
						set output to output & (name of t) & sep & (artist of t) & sep & (album of t) & sep & (played count of t) & sep & (rating of t) & nl
					end try
				end repeat
			on error errText
				return "ERROR" & sep & errText
			end try
			return output
		end tell
	`)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(out, "ERROR"+recordSep) {
		return nil, fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR"+recordSep))
	}
	var tracks []TrackStats
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, recordSep, 5)
		if len(parts) < 5 {
			continue
		}
		playCount, _ := strconv.Atoi(parts[3])
		rating, _ := strconv.Atoi(parts[4])
		tracks = append(tracks, TrackStats{
			Title:     parts[0],
			Artist:    parts[1],
			Album:     parts[2],
			PlayCount: playCount,
			Rating:    rating,
		})
	}
	sort.SliceStable(tracks, func(i, j int) bool {
		return tracks[i].PlayCount > tracks[j].PlayCount
	})
	if len(tracks) > limit {
		tracks = tracks[:limit]
	}
	return tracks, nil
}

// PlayAlbumByName plays the first matching album from the local library.
func PlayAlbumByName(query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return fmt.Errorf("album query is required")
	}
	if err := EnsureMusicRunning(); err != nil {
		return err
	}
	q := escapeAS(query)
	out, err := RunAppleScript(fmt.Sprintf(`
		tell application "Music"
			try
				set matches to (every track of library playlist 1 whose album contains "%s")
				if (count of matches) is 0 then return "NOT_FOUND"
				set firstTrack to item 1 of matches
				set albumName to album of firstTrack
				set albumArtist to artist of firstTrack
				set albumTracks to (every track of library playlist 1 whose album is albumName and artist is albumArtist)
				if (count of albumTracks) is 0 then return "NOT_FOUND"
				play item 1 of albumTracks
				return "PLAYING"
			on error errText
				return "ERROR|" & errText
			end try
		end tell
	`, q))
	if err != nil {
		return err
	}
	if out == "NOT_FOUND" {
		return fmt.Errorf("no album found matching: %s", query)
	}
	if strings.HasPrefix(out, "ERROR|") {
		return fmt.Errorf("applescript: %s", strings.TrimPrefix(out, "ERROR|"))
	}
	return nil
}

// escapeAS escapes backslashes and double-quotes for embedding in AppleScript strings.
func escapeAS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// parseTrackLines splits pipe-delimited newline-separated track output.
func parseTrackLines(out string) []TrackInfo {
	return parseRecordTrackLines(out, "|")
}

func parseRecordTrackLines(out, sep string) []TrackInfo {
	var tracks []TrackInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, sep, 3)
		if len(parts) >= 3 {
			tracks = append(tracks, TrackInfo{Title: parts[0], Artist: parts[1], Album: parts[2]})
		}
	}
	return tracks
}
