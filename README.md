# Muse

Muse is a terminal controller for Apple Music on macOS. It provides a fast CLI, an interactive Bubble Tea dashboard, and an MCP stdio server for agent workflows.

[![CI](https://github.com/ryanrodrigues25200525-svg/Apple-music-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/ryanrodrigues25200525-svg/Apple-music-cli/actions/workflows/ci.yml)

## Highlights

- Control Music.app from the terminal with `mu`.
- Search and play tracks, playlists, queue, lyrics, artwork, love, shuffle, repeat, sleep timer, and stats.
- Launch a Bubble Tea dashboard with keyboard controls.
- Expose playback and library actions as MCP tools for agents.
- Built for macOS with AppleScript, no Apple Music API credentials required.

## Install

Muse requires macOS, Music.app, and AppleScript Automation permission for your terminal.

### Homebrew

```sh
brew tap ryanrodrigues25200525-svg/homebrew-tap
brew install muse
```

Upgrade later with:

```sh
brew upgrade muse
```

### GitHub Releases

Download the latest macOS universal archive from:

```text
https://github.com/ryanrodrigues25200525-svg/Apple-music-cli/releases/latest
```

Then unpack it and move `mu` into a directory on your `PATH`, such as `/usr/local/bin` or `~/.local/bin`.

### Install Script

From a checked-out repo:

```sh
./scripts/install.sh
```

Or directly from GitHub:

```sh
curl -fsSL https://raw.githubusercontent.com/ryanrodrigues25200525-svg/Apple-music-cli/main/scripts/install.sh | sh
```

### Build From Source

```sh
git clone https://github.com/ryanrodrigues25200525-svg/Apple-music-cli.git
cd Apple-music-cli
make build
./mu version
```

Install to `~/.local/bin`:

```sh
make install
```

## Commands

Playback:

```sh
mu play
mu play "Nights"
mu pause
mu toggle
mu next
mu prev
mu stop
mu seek +30
mu seek -10
```

Library and metadata:

```sh
mu now
mu now --json
mu search "Frank Ocean"
mu search "Frank Ocean" --json
mu playlists
mu playlists --json
mu playlist "Gym"
mu playlist "Gym" --shuffle
mu queue
mu queue --json
mu stats
mu stats --json
mu love
```

Playback modes and volume:

```sh
mu volume
mu volume 50
mu shuffle toggle
mu repeat
mu sleep 30
```

Interactive modes:

```sh
mu
mu mini
```

Diagnostics and shell integration:

```sh
mu doctor
mu completion zsh
mu completion bash
mu completion fish
```

## TUI

Running `mu` opens the dashboard. Current keybindings include:

- `Space`: toggle playback
- `n` / `p`: next / previous
- `s`: shuffle
- `r`: repeat
- `l`: love current track
- `[` / `]`: seek backward / forward
- `+` / `-`: volume up / down
- `/`: search
- `q`: quit

Screenshots or GIFs can be added under a future `docs/` or `assets/` directory.

## MCP

Start the MCP server over stdio:

```sh
mu mcp
```

Available tools include playback controls, `now_playing`, `search`, `play_song`, `play_playlist`, `play_playlist_shuffled`, `lyrics`, `queue`, `stats`, `love`, `shuffle`, `repeat`, `seek`, and `music_context`.

Data-heavy tools return JSON text so agents can parse the response reliably.

## Troubleshooting

Run:

```sh
mu doctor
```

Common issues:

- Music.app is not installed or cannot be launched.
- `osascript` is unavailable.
- macOS Automation permission is missing. Open System Settings, Privacy & Security, Automation, then grant your terminal access to Music.app.
- Lyrics require network access to `lrclib.net`.

## Privacy And Security

Muse does not require Apple Music API credentials. Local playback control happens through AppleScript and Music.app. Lyrics lookup sends the current track title, artist, album, and duration to `lrclib.net`.

Report security issues using [SECURITY.md](SECURITY.md).

## Release

Release builds inject the version with:

```sh
go build -ldflags "-X github.com/ryanrodrigues25200525-svg/Apple-music-cli/cmd.Version=vX.Y.Z" -o mu main.go
```

The included GoReleaser config builds macOS archives for Apple Silicon and Intel.

## License

MIT. See [LICENSE](LICENSE).
