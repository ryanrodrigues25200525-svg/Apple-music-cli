# Contributing

## Local Checks

Run the same checks as CI before opening a pull request:

```sh
go mod tidy
go test ./...
go test -race ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go build -o /tmp/mu-check main.go
```

## Manual Smoke Test

On macOS with Music.app available:

```sh
mu doctor
mu now
mu play
mu pause
mu seek +10
mu shuffle toggle
mu repeat
mu playlists
mu queue
mu mcp
```

Do not commit local binaries or release artifacts.
