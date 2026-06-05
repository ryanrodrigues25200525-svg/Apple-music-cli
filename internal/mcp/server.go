package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ryanrodrigues25200525-svg/Apple-music-cli/internal/lyrics"
	"github.com/ryanrodrigues25200525-svg/Apple-music-cli/internal/music"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type InitializeResult struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    map[string]object `json:"capabilities"`
	ServerInfo      ServerInfo        `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type object map[string]interface{}

type musicActions struct {
	Play                 func() error
	Pause                func() error
	Toggle               func() error
	Next                 func() error
	Prev                 func() error
	Stop                 func() error
	NowPlaying           func() (*music.TrackInfo, error)
	SetVolume            func(int) error
	Search               func(string) ([]music.TrackInfo, error)
	PlayTrackByName      func(string) error
	PlayPlaylist         func(string) error
	PlayPlaylistShuffled func(string) error
	GetQueue             func() ([]music.TrackInfo, error)
	GetTrackStats        func() (*music.TrackStats, error)
	ToggleLove           func() (bool, error)
	ToggleShuffle        func() (bool, error)
	CycleRepeat          func() (string, error)
	Seek                 func(float64) error
	FetchLyrics          func(string, string, string, float64) ([]lyrics.Line, error)
}

var musicAPI = musicActions{
	Play:                 music.Play,
	Pause:                music.Pause,
	Toggle:               music.Toggle,
	Next:                 music.Next,
	Prev:                 music.Prev,
	Stop:                 music.Stop,
	NowPlaying:           music.NowPlaying,
	SetVolume:            music.SetVolume,
	Search:               music.Search,
	PlayTrackByName:      music.PlayTrackByName,
	PlayPlaylist:         music.PlayPlaylist,
	PlayPlaylistShuffled: music.PlayPlaylistShuffled,
	GetQueue:             music.GetQueue,
	GetTrackStats:        music.GetTrackStats,
	ToggleLove:           music.ToggleLove,
	ToggleShuffle:        music.ToggleShuffle,
	CycleRepeat:          music.CycleRepeat,
	Seek:                 music.Seek,
	FetchLyrics:          lyrics.Fetch,
}

func StartServer() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		handleRequest(&req)
	}
}

func sendResponse(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	bz, err := json.Marshal(resp)
	if err == nil {
		os.Stdout.Write(bz)
		os.Stdout.Write([]byte("\n"))
	}
}

func sendError(id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	bz, err := json.Marshal(resp)
	if err == nil {
		os.Stdout.Write(bz)
		os.Stdout.Write([]byte("\n"))
	}
}

func handleRequest(req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		sendResponse(req.ID, InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]object{
				"tools": {},
			},
			ServerInfo: ServerInfo{
				Name:    "muse-mcp",
				Version: "0.1.0",
			},
		})

	case "notifications/initialized":
		// No-op

	case "tools/list":
		sendResponse(req.ID, object{"tools": availableTools()})

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			sendError(req.ID, -32602, "Invalid params", err.Error())
			return
		}

		result, isErr := executeTool(params.Name, params.Arguments)
		sendResponse(req.ID, object{
			"content": []object{
				{
					"type": "text",
					"text": result,
				},
			},
			"isError": isErr,
		})

	default:
		sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method), nil)
	}
}

func availableTools() []object {
	return []object{
		{
			"name":        "play",
			"description": "Resume playback or start playing music",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "pause",
			"description": "Pause music playback",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "toggle",
			"description": "Toggle play/pause state",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "next",
			"description": "Skip to the next song",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "previous",
			"description": "Skip to the previous song",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "stop",
			"description": "Stop music playback",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "now_playing",
			"description": "Retrieve information about the currently playing track",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "set_volume",
			"description": "Set the Music.app volume (0-100)",
			"inputSchema": object{
				"type": "object",
				"properties": object{
					"volume": object{
						"type":        "integer",
						"description": "Volume level between 0 and 100",
					},
				},
				"required": []string{"volume"},
			},
		},
		{
			"name":        "search",
			"description": "Search the user's music library for songs by title, artist, or album",
			"inputSchema": object{
				"type": "object",
				"properties": object{
					"query": object{
						"type":        "string",
						"description": "The search term (song title, artist, or album name)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "play_song",
			"description": "Search for and play a specific song by its name, artist, or album",
			"inputSchema": object{
				"type": "object",
				"properties": object{
					"query": object{
						"type":        "string",
						"description": "The name, artist, or album of the song to play",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "play_playlist",
			"description": "Play a playlist by name",
			"inputSchema": object{
				"type": "object",
				"properties": object{
					"name": object{
						"type":        "string",
						"description": "The exact name of the playlist to play",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			"name":        "play_playlist_shuffled",
			"description": "Enable shuffle and play a playlist by name",
			"inputSchema": object{
				"type": "object",
				"properties": object{
					"name": object{
						"type":        "string",
						"description": "The exact name of the playlist to play",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			"name":        "lyrics",
			"description": "Fetch lyrics for the currently playing track",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "queue",
			"description": "Return upcoming tracks in the current playlist",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "stats",
			"description": "Return play count, rating, loved status, and metadata for the current track",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "love",
			"description": "Toggle loved status for the current track",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "shuffle",
			"description": "Toggle shuffle mode",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "repeat",
			"description": "Cycle repeat mode",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
		{
			"name":        "seek",
			"description": "Seek forward or backward by seconds",
			"inputSchema": object{
				"type": "object",
				"properties": object{
					"seconds": object{
						"type":        "number",
						"description": "Signed number of seconds to seek, for example 30 or -10",
					},
				},
				"required": []string{"seconds"},
			},
		},
		{
			"name":        "music_context",
			"description": "Return now playing, stats, queue, and lyrics when available",
			"inputSchema": object{
				"type":       "object",
				"properties": object{},
			},
		},
	}
}

func executeTool(name string, args map[string]interface{}) (string, bool) {
	switch name {
	case "play":
		if err := musicAPI.Play(); err != nil {
			return fmt.Sprintf("Error playing: %v", err), true
		}
		return "Playback started/resumed.", false

	case "pause":
		if err := musicAPI.Pause(); err != nil {
			return fmt.Sprintf("Error pausing: %v", err), true
		}
		return "Playback paused.", false

	case "toggle":
		if err := musicAPI.Toggle(); err != nil {
			return fmt.Sprintf("Error toggling playback: %v", err), true
		}
		return "Playback toggled.", false

	case "next":
		if err := musicAPI.Next(); err != nil {
			return fmt.Sprintf("Error playing next track: %v", err), true
		}
		return "Skipped to next track.", false

	case "previous":
		if err := musicAPI.Prev(); err != nil {
			return fmt.Sprintf("Error playing previous track: %v", err), true
		}
		return "Skipped to previous track.", false

	case "stop":
		if err := musicAPI.Stop(); err != nil {
			return fmt.Sprintf("Error stopping: %v", err), true
		}
		return "Playback stopped.", false

	case "now_playing":
		info, err := musicAPI.NowPlaying()
		if err != nil {
			return fmt.Sprintf("Error fetching now playing: %v", err), true
		}
		return jsonText(info), false

	case "set_volume":
		// JSON numbers unmarshal as float64
		volFloat, ok := args["volume"].(float64)
		if !ok {
			return "volume argument is required and must be a number", true
		}
		vol := int(volFloat)
		if err := musicAPI.SetVolume(vol); err != nil {
			return fmt.Sprintf("Error setting volume: %v", err), true
		}
		return fmt.Sprintf("Volume set to %d%%.", vol), false

	case "search":
		query, ok := args["query"].(string)
		if !ok || query == "" {
			return "query argument is required", true
		}
		results, err := musicAPI.Search(query)
		if err != nil {
			return fmt.Sprintf("Error searching library: %v", err), true
		}
		return jsonText(results), false

	case "play_song":
		query, ok := args["query"].(string)
		if !ok || query == "" {
			return "query argument is required", true
		}
		if err := musicAPI.PlayTrackByName(query); err != nil {
			return fmt.Sprintf("Error playing track: %v", err), true
		}
		return fmt.Sprintf("Playing track matching: '%s'", query), false

	case "play_playlist":
		name, ok := args["name"].(string)
		if !ok || name == "" {
			return "name argument is required", true
		}
		if err := musicAPI.PlayPlaylist(name); err != nil {
			return fmt.Sprintf("Error playing playlist: %v", err), true
		}
		return fmt.Sprintf("Playing playlist: %s", name), false

	case "play_playlist_shuffled":
		name, ok := args["name"].(string)
		if !ok || name == "" {
			return "name argument is required", true
		}
		if err := musicAPI.PlayPlaylistShuffled(name); err != nil {
			return fmt.Sprintf("Error playing playlist shuffled: %v", err), true
		}
		return fmt.Sprintf("Playing playlist shuffled: %s", name), false

	case "lyrics":
		lines, err := lyricsForCurrentTrack()
		if err != nil {
			return fmt.Sprintf("Error fetching lyrics: %v", err), true
		}
		return jsonText(lines), false

	case "queue":
		tracks, err := musicAPI.GetQueue()
		if err != nil {
			return fmt.Sprintf("Error fetching queue: %v", err), true
		}
		return jsonText(tracks), false

	case "stats":
		stats, err := musicAPI.GetTrackStats()
		if err != nil {
			return fmt.Sprintf("Error fetching stats: %v", err), true
		}
		return jsonText(stats), false

	case "love":
		loved, err := musicAPI.ToggleLove()
		if err != nil {
			return fmt.Sprintf("Error toggling loved status: %v", err), true
		}
		return jsonText(object{"loved": loved}), false

	case "shuffle":
		enabled, err := musicAPI.ToggleShuffle()
		if err != nil {
			return fmt.Sprintf("Error toggling shuffle: %v", err), true
		}
		return jsonText(object{"shuffle": enabled}), false

	case "repeat":
		mode, err := musicAPI.CycleRepeat()
		if err != nil {
			return fmt.Sprintf("Error cycling repeat: %v", err), true
		}
		return jsonText(object{"repeat": mode}), false

	case "seek":
		seconds, ok := args["seconds"].(float64)
		if !ok || seconds == 0 {
			return "seconds argument is required and must be a non-zero number", true
		}
		if err := musicAPI.Seek(seconds); err != nil {
			return fmt.Sprintf("Error seeking: %v", err), true
		}
		return jsonText(object{"seek_seconds": seconds}), false

	case "music_context":
		return jsonText(buildMusicContext()), false

	default:
		return fmt.Sprintf("Unknown tool: %s", name), true
	}
}

func jsonText(v interface{}) string {
	bz, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(bz)
}

func lyricsForCurrentTrack() ([]lyrics.Line, error) {
	info, err := musicAPI.NowPlaying()
	if err != nil {
		return nil, err
	}
	if info == nil || info.State == "stopped" {
		return []lyrics.Line{}, nil
	}
	lines, err := musicAPI.FetchLyrics(info.Title, info.Artist, info.Album, info.Duration)
	if err != nil {
		return nil, err
	}
	if lines == nil {
		lines = []lyrics.Line{}
	}
	return lines, nil
}

func buildMusicContext() object {
	ctx := object{}

	info, err := musicAPI.NowPlaying()
	if err != nil {
		ctx["now_playing_error"] = err.Error()
		return ctx
	}
	ctx["now_playing"] = info
	if info == nil || info.State == "stopped" {
		return ctx
	}

	if stats, err := musicAPI.GetTrackStats(); err == nil {
		ctx["stats"] = stats
	} else {
		ctx["stats_error"] = err.Error()
	}
	if queue, err := musicAPI.GetQueue(); err == nil {
		ctx["queue"] = queue
	} else {
		ctx["queue_error"] = err.Error()
	}
	if lines, err := musicAPI.FetchLyrics(info.Title, info.Artist, info.Album, info.Duration); err == nil {
		if lines == nil {
			lines = []lyrics.Line{}
		}
		ctx["lyrics"] = lines
	} else {
		ctx["lyrics_error"] = err.Error()
	}

	return ctx
}
