package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)

	if m.workDuration != 25*time.Minute {
		t.Errorf("expected work duration 25m, got %v", m.workDuration)
	}
	if m.breakDuration != 5*time.Minute {
		t.Errorf("expected break duration 5m, got %v", m.breakDuration)
	}
	if m.remaining != 25*time.Minute {
		t.Errorf("expected remaining 25m, got %v", m.remaining)
	}
	if m.state != WorkState {
		t.Errorf("expected WorkState, got %v", m.state)
	}
	if m.running {
		t.Error("expected timer to start paused")
	}
}

func TestNewModelCustomDurations(t *testing.T) {
	m := NewModel(50, 10, NotifyBoth)

	if m.workDuration != 50*time.Minute {
		t.Errorf("expected work duration 50m, got %v", m.workDuration)
	}
	if m.breakDuration != 10*time.Minute {
		t.Errorf("expected break duration 10m, got %v", m.breakDuration)
	}
}

func TestSwitchPhaseFromWorkToBreak(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)
	m.state = WorkState
	m.running = true

	m = m.switchPhase()

	if m.state != BreakState {
		t.Errorf("expected BreakState, got %v", m.state)
	}
	if m.remaining != 5*time.Minute {
		t.Errorf("expected remaining 5m, got %v", m.remaining)
	}
	if m.totalDuration != 5*time.Minute {
		t.Errorf("expected totalDuration 5m, got %v", m.totalDuration)
	}
	if m.running {
		t.Error("expected timer to pause after phase switch")
	}
}

func TestSwitchPhaseFromBreakToWork(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)
	m.state = BreakState
	m.remaining = 5 * time.Minute
	m.totalDuration = 5 * time.Minute

	m = m.switchPhase()

	if m.state != WorkState {
		t.Errorf("expected WorkState, got %v", m.state)
	}
	if m.remaining != 25*time.Minute {
		t.Errorf("expected remaining 25m, got %v", m.remaining)
	}
	if m.totalDuration != 25*time.Minute {
		t.Errorf("expected totalDuration 25m, got %v", m.totalDuration)
	}
}

func TestUpdateKeyG(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)

	// First press - should start
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = newModel.(Model)
	if !m.running {
		t.Error("expected timer to start after 'g' press")
	}

	// Second press - should pause
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = newModel.(Model)
	if m.running {
		t.Error("expected timer to pause after second 'g' press")
	}
}

func TestUpdateKeyS(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)
	m.state = WorkState

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newModel.(Model)

	if m.state != BreakState {
		t.Error("expected skip to switch to BreakState")
	}
}

func TestUpdateKeyR(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)
	m.remaining = 10 * time.Minute
	m.running = true

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	if m.remaining != 25*time.Minute {
		t.Errorf("expected reset to 25m, got %v", m.remaining)
	}
	if m.running {
		t.Error("expected timer to pause after reset")
	}
}

func TestUpdateKeyQ(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// tea.Quit returns a special command
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestTickDecreasesRemaining(t *testing.T) {
	m := NewModel(25, 5, NotifyNone)
	m.running = true
	m.remaining = 10 * time.Second

	newModel, _ := m.Update(TickMsg(time.Now()))
	m = newModel.(Model)

	if m.remaining != 9*time.Second {
		t.Errorf("expected 9s remaining, got %v", m.remaining)
	}
}

func TestTickDoesNotDecreaseWhenPaused(t *testing.T) {
	m := NewModel(25, 5, NotifyNone)
	m.running = false
	m.remaining = 10 * time.Second

	newModel, _ := m.Update(TickMsg(time.Now()))
	m = newModel.(Model)

	if m.remaining != 10*time.Second {
		t.Errorf("expected 10s remaining (paused), got %v", m.remaining)
	}
}

func TestTickSwitchesPhaseAtZero(t *testing.T) {
	m := NewModel(25, 5, NotifyNone)
	m.running = true
	m.remaining = 1 * time.Second
	m.state = WorkState

	newModel, _ := m.Update(TickMsg(time.Now()))
	m = newModel.(Model)

	if m.state != BreakState {
		t.Error("expected phase switch to BreakState when timer reaches zero")
	}
	if m.running {
		t.Error("expected timer to pause after phase switch")
	}
}

func TestViewContainsElements(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)
	view := m.View()

	// Check for essential elements
	if len(view) == 0 {
		t.Error("view should not be empty")
	}

	// Check for timer display (should contain colon for MM:SS format)
	if !contains(view, ":") {
		t.Error("view should contain time display with colon")
	}

	// Check for progress bar
	if !contains(view, "[") || !contains(view, "]") {
		t.Error("view should contain progress bar brackets")
	}

	// Check for help text
	if !contains(view, "quit") {
		t.Error("view should contain help text")
	}
}

func TestViewShowsCorrectState(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)

	m.state = WorkState
	view := m.View()
	if !contains(view, "WORK") {
		t.Error("view should show WORK state")
	}

	m.state = BreakState
	view = m.View()
	if !contains(view, "BREAK") {
		t.Error("view should show BREAK state")
	}
}

func TestViewShowsCorrectStatus(t *testing.T) {
	m := NewModel(25, 5, NotifyBoth)

	m.running = false
	view := m.View()
	if !contains(view, "Paused") {
		t.Error("view should show Paused status")
	}

	m.running = true
	view = m.View()
	if !contains(view, "Running") {
		t.Error("view should show Running status")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
