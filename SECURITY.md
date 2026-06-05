# Security Policy

## Supported Versions

Muse is currently pre-1.0. Security fixes target the latest release only.

## Reporting A Vulnerability

Do not open a public issue for sensitive security reports. Send the report privately to the project maintainer with:

- affected version or commit
- macOS version
- reproduction steps
- impact
- any relevant terminal output with personal data removed

## Privacy Notes

Muse controls the local macOS Music.app through AppleScript. It does not require Apple Music API credentials. Lyrics lookup uses `lrclib.net` and sends track title, artist, album, and duration for the current track.
