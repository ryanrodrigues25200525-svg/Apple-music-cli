# Apple Music CLI Project Plan

## Project Name

**Muse**

A Mole-inspired Apple Music CLI for macOS.

---

## Goal

Build a beautiful terminal app that lets users control Apple Music from the command line.

```bash
mu
mu play "Nights"
mu pause
mu next
mu now
mu playlist "Gym"
Approach

Use the easiest reliable path:

Go CLI + Bubble Tea TUI + AppleScript + Music.app

No Apple Music API for v1.

Phase 1: Setup
Tasks
Create GitHub repo
Initialize Go project
Add Cobra for CLI
Add Bubble Tea for TUI
Add Lip Gloss for styling
Add basic Makefile
Commands
mkdir muse
cd muse
go mod init github.com/yourname/muse

go get github.com/spf13/cobra
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
Folder Structure
muse/
├── cmd/
├── internal/
│   ├── music/
│   ├── tui/
│   └── config/
├── scripts/
├── Makefile
├── README.md
└── go.mod
Phase 2: Apple Music Controller
Goal

Create a Go wrapper around AppleScript.

Features
mu play
mu pause
mu toggle
mu next
mu prev
mu stop
mu now
mu volume 50
Core Function
func RunAppleScript(script string) (string, error) {
    cmd := exec.Command("osascript", "-e", script)
    out, err := cmd.Output()
    return strings.TrimSpace(string(out)), err
}
AppleScript Examples
tell application "Music" to play
tell application "Music" to pause
tell application "Music" to next track
tell application "Music" to previous track
tell application "Music" to sound volume
Phase 3: CLI Commands
Implement
mu play
mu pause
mu toggle
mu next
mu prev
mu stop
mu now
mu volume
Example
mu now

Output:

🎵 Nights
👤 Frank Ocean
💿 Blonde
▶ Playing
🔊 70%
Phase 4: Search & Play
Goal

Allow users to search their Music.app library.

Commands
mu search "Frank Ocean"
mu play "Nights"
mu artist "Frank Ocean"
mu album "Blonde"
mu playlist "Gym"
Search Strategy

Start simple:

tell application "Music"
    set results to every track of library playlist 1 whose name contains "Nights"
end tell

Then improve later with fuzzy matching.

Phase 5: Interactive TUI
Goal

Make mu open a beautiful dashboard.

Screens
Muse

Now Playing
────────────────────────
🎵 Nights
👤 Frank Ocean
💿 Blonde
▶ Playing
🔊 70%

Menu
────────────────────────
▶ Search
  Playlists
  Albums
  Artists
  Queue
  Settings

↑↓ Navigate  Enter Select  Space Play/Pause  n Next  q Quit
Keybinds
space  play/pause
n      next
p      previous
/      search
q      quit
r      refresh
Phase 6: Polish
Add
mu --json
mu completion zsh
mu completion bash
mu version
mu doctor
mu doctor

Checks:

macOS detected
Music.app installed
AppleScript permission works
Music.app can be launched
Current track can be read
Phase 7: Packaging
Install Script
curl -fsSL https://raw.githubusercontent.com/yourname/muse/main/scripts/install.sh | bash
Homebrew

Create:

homebrew-tap/Formula/muse.rb
Releases

Use:

GitHub Releases
GoReleaser
Universal macOS binary
Phase 8: Optional MCP Server
Goal

Expose Muse as MCP tools.

mu mcp
Tools
{
  "play": {},
  "pause": {},
  "next": {},
  "previous": {},
  "now_playing": {},
  "search": {
    "query": "Frank Ocean"
  },
  "play_song": {
    "query": "Nights"
  }
}
Roadmap
v0.1
Basic playback commands
Now playing
Volume
v0.2
Search
Play song by name
Play playlist
v0.3
Interactive dashboard
v0.4
JSON output
Shell completions
Doctor command
v1.0
Homebrew install
Polished README
Stable commands
GitHub release
v1.1+
Artwork
Lyrics
Queue
MCP server
Optional Apple Music API integration
Build Priority
mu now
mu play
mu pause
mu next
mu prev
mu volume
mu play "song"
mu playlist "name"
mu
mu mcp
Success Criteria
Works with Apple Music on macOS
Installs with one command
Opens instantly
Looks good in terminal
Feels like a native developer tool
Does not require Apple Developer account
Can become an MCP server later