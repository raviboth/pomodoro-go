package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

// Default durations in minutes
const (
	defaultWorkMinutes  = 25
	defaultBreakMinutes = 5
)

// NotifyMode represents the notification type
type NotifyMode string

const (
	NotifyNone   NotifyMode = "none"
	NotifyVisual NotifyMode = "visual"
	NotifyAudio  NotifyMode = "audio"
	NotifyBoth   NotifyMode = "both"
)

// TimerState represents whether we're in work or break mode
type TimerState int

const (
	WorkState TimerState = iota
	BreakState
)

// Model holds the application state
type Model struct {
	workDuration  time.Duration
	breakDuration time.Duration
	remaining     time.Duration
	state         TimerState
	running       bool
	totalDuration time.Duration
	notifyMode    NotifyMode
}

// TickMsg is sent every second when the timer is running
type TickMsg time.Time

// Initialize the model with given durations
func NewModel(workMinutes, breakMinutes int, notifyMode NotifyMode) Model {
	workDur := time.Duration(workMinutes) * time.Minute
	return Model{
		workDuration:  workDur,
		breakDuration: time.Duration(breakMinutes) * time.Minute,
		remaining:     workDur,
		state:         WorkState,
		running:       false,
		totalDuration: workDur,
		notifyMode:    notifyMode,
	}
}

// Init starts the timer tick
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// tickCmd returns a command that sends a tick every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "g", " ":
			// Toggle start/pause
			m.running = !m.running
		case "s":
			// Skip to next phase
			m = m.switchPhase()
		case "r":
			// Reset current timer
			m.remaining = m.totalDuration
			m.running = false
		}

	case TickMsg:
		if m.running && m.remaining > 0 {
			m.remaining -= time.Second
			if m.remaining <= 0 {
				m.remaining = 0
				m.running = false
				m.sendNotification()
				m = m.switchPhase()
			}
		}
		return m, tickCmd()
	}

	return m, nil
}

// switchPhase toggles between work and break
func (m Model) switchPhase() Model {
	if m.state == WorkState {
		m.state = BreakState
		m.remaining = m.breakDuration
		m.totalDuration = m.breakDuration
	} else {
		m.state = WorkState
		m.remaining = m.workDuration
		m.totalDuration = m.workDuration
	}
	m.running = false
	return m
}

// playSound plays a notification sound using system audio tools
func playSound() {
	switch runtime.GOOS {
	case "darwin":
		// macOS: use afplay with a system sound
		soundFile := "/System/Library/Sounds/Ping.aiff"
		if path, err := exec.LookPath("afplay"); err == nil {
			cmd := exec.Command(path, soundFile)
			_ = cmd.Start()
		}
	case "linux":
		// Linux: use paplay with freedesktop sound
		soundFile := "/usr/share/sounds/freedesktop/stereo/alarm-clock-elapsed.oga"
		if path, err := exec.LookPath("paplay"); err == nil {
			cmd := exec.Command(path, "--volume=65536", soundFile)
			_ = cmd.Start()
		}
	case "windows":
		// Windows: use PowerShell to play a system sound
		if path, err := exec.LookPath("powershell"); err == nil {
			cmd := exec.Command(path, "-c", `(New-Object Media.SoundPlayer 'C:\Windows\Media\Alarm01.wav').PlaySync()`)
			_ = cmd.Start()
		}
	}
}

// sendNotification sends a notification based on the configured mode
func (m Model) sendNotification() {
	if m.notifyMode == NotifyNone {
		return
	}

	var title, message string
	if m.state == WorkState {
		title = "Pomodoro Timer"
		message = "Work session complete! Time for a break."
	} else {
		title = "Pomodoro Timer"
		message = "Break is over! Ready to work?"
	}

	// Send notification based on mode (errors are best-effort ignored)
	switch m.notifyMode {
	case NotifyVisual:
		_ = beeep.Notify(title, message, "")
	case NotifyAudio:
		playSound()
	case NotifyBoth:
		_ = beeep.Notify(title, message, "")
		playSound()
	}
}

// View renders the UI
func (m Model) View() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	// Build the view
	var b strings.Builder

	// Title
	stateLabel := "WORK"
	if m.state == BreakState {
		stateLabel = "BREAK"
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("Pomodoro Timer - %s", stateLabel)))
	b.WriteString("\n\n")

	// Progress bar with centered clock
	minutes := int(m.remaining.Minutes())
	seconds := int(m.remaining.Seconds()) % 60
	clock := fmt.Sprintf("%02d:%02d", minutes, seconds)

	progressWidth := 40
	elapsed := m.totalDuration - m.remaining
	progress := 0.0
	if m.totalDuration > 0 {
		progress = float64(elapsed) / float64(m.totalDuration)
	}
	filled := int(progress * float64(progressWidth))

	// Build bar with clock centered
	clockStart := (progressWidth - len(clock)) / 2
	clockEnd := clockStart + len(clock)

	var bar strings.Builder
	bar.WriteString("[")
	for i := 0; i < progressWidth; i++ {
		if i >= clockStart && i < clockEnd {
			bar.WriteByte(clock[i-clockStart])
		} else if i < filled {
			bar.WriteString("=")
		} else {
			bar.WriteString(" ")
		}
	}
	bar.WriteString("]")

	percentage := fmt.Sprintf(" %.0f%%", progress*100)
	b.WriteString(progressStyle.Render(bar.String() + percentage))
	b.WriteString("\n\n")

	// Status
	status := "Paused"
	if m.running {
		status = "Running"
	}
	b.WriteString(statusStyle.Render(fmt.Sprintf("Status: %s", status)))
	b.WriteString("\n")

	// Help
	help := "space/g: start/pause | s: skip | r: reset | q: quit"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func main() {
	// Parse flags
	workFlag := flag.Int("work", defaultWorkMinutes, "Work duration in minutes")
	breakFlag := flag.Int("break", defaultBreakMinutes, "Break duration in minutes")
	notifyFlag := flag.String("notify", "both", "Notification mode: none, visual, audio, both")
	flag.Parse()

	workMinutes := *workFlag
	breakMinutes := *breakFlag

	// Validate and set notification mode
	notifyMode := NotifyMode(*notifyFlag)
	switch notifyMode {
	case NotifyNone, NotifyVisual, NotifyAudio, NotifyBoth:
		// valid
	default:
		fmt.Printf("Invalid notify mode: %s (use: none, visual, audio, both)\n", *notifyFlag)
		os.Exit(1)
	}

	// Check for positional arguments (override flags if provided)
	args := flag.Args()
	if len(args) >= 2 {
		if w, err := strconv.Atoi(args[0]); err == nil && w > 0 {
			workMinutes = w
		}
		if b, err := strconv.Atoi(args[1]); err == nil && b > 0 {
			breakMinutes = b
		}
	} else if len(args) == 1 {
		if w, err := strconv.Atoi(args[0]); err == nil && w > 0 {
			workMinutes = w
		}
	}

	// Create and run the program
	model := NewModel(workMinutes, breakMinutes, notifyMode)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
