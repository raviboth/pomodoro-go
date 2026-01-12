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

Then run with:
```bash
pomodoro-go
```

> **Note:** If `pomodoro-go` isn't found, add Go's bin directory to your PATH:
> ```bash
> echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc  # or ~/.zshrc
> source ~/.bashrc
> ```

Or build from source:

```bash
git clone https://github.com/raviboth/pomodoro-go.git
cd pomodoro-go
go build -o pomodoro .
./pomodoro
```

## Usage

```bash
# Default: 25 min work, 5 min break
pomodoro-go

# Custom durations
pomodoro-go -work 30 -break 10

# Notification modes: none, visual, audio, both (default)
pomodoro-go -notify audio
```

## Controls

| Key | Action |
|-----|--------|
| `space` / `g` | Start/pause timer |
| `s` | Skip to next phase |
| `r` | Reset current timer |
| `q` | Quit |

## Requirements

- Linux with PulseAudio/PipeWire for audio notifications
- Desktop notification daemon for visual notifications

## License

MIT
