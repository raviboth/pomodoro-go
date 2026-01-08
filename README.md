# pomodoro-go

A terminal-based Pomodoro timer built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- Work/break cycle management (default: 25min work, 5min break)
- Visual progress bar
- Desktop notifications and audio alerts
- Configurable durations and notification modes

## Installation

```bash
go install github.com/raviboth/pomodoro-go@latest
```

Or build from source:

```bash
git clone https://github.com/raviboth/pomodoro-go.git
cd pomodoro-go
go build -o pomodoro .
```

## Usage

```bash
# Default: 25 min work, 5 min break
./pomodoro

# Custom durations
./pomodoro -work 30 -break 10

# Notification modes: none, visual, audio, both (default)
./pomodoro -notify audio
```

## Controls

| Key | Action |
|-----|--------|
| `g` | Start/pause timer |
| `s` | Skip to next phase |
| `r` | Reset current timer |
| `q` | Quit |

## Requirements

- Linux with PulseAudio/PipeWire for audio notifications
- Desktop notification daemon for visual notifications

## License

MIT
