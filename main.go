package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func run() error {

	script := `tell application "System Events" to get name of every window of process "Alacritty"`

	var m model

	{

		buf := new(bytes.Buffer)

		cmd := exec.Command("osascript", "-e", script)
		cmd.Stderr = os.Stderr
		cmd.Stdout = buf

		if err := cmd.Run(); err != nil {
			return err
		}

		m.windows = strings.Split(buf.String(), ", ")
	}

	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("tea: run: %w", err)
	}

	return nil
}

type model struct {
	windows []string
	cursor  int
}

func (m model) Init() tea.Cmd {
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "k", "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 0
			}
		case "j", "down":
			m.cursor++
			if m.cursor > len(m.windows)-1 {
				m.cursor = len(m.windows) - 1
			}
		}
	}
	return m, nil
}
func (m model) View() tea.View {
	b := new(strings.Builder)
	for i, title := range m.windows {
		if m.cursor == i {
			fmt.Fprint(b, "> ")
		} else {
			fmt.Fprint(b, "  ")
		}
		fmt.Fprintln(b, title)
	}
	return tea.NewView(b.String())
}
