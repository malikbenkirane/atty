package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

const (
	listScript  = `tell application "System Events" to get name of every window of process "Alacritty"`
	raiseScript = `tell application "System Events" to tell process "Alacritty" to perform action "AXRaise" of (first window whose name contains %q)`
)

func run() error {

	var m model

	{

		buf := new(bytes.Buffer)

		cmd := exec.Command("osascript", "-e", listScript)
		cmd.Stderr = os.Stderr
		cmd.Stdout = buf

		if err := cmd.Run(); err != nil {
			return err
		}

		m.windows = strings.Split(buf.String(), ", ")
		slices.Sort(m.windows)

		for i, t := range m.windows {
			m.windows[i] = strings.TrimSpace(t)
		}

	}

	m.filtering = true

	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("tea: run: %w", err)
	}

	return nil
}

type model struct {
	windows    []string
	cursor     int
	filtering  bool
	filterText string
	err        error
}

func (m model) Init() tea.Cmd {
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.filtering {
			return m.updateFilter(msg)
		}
		return m.updateCursor(msg)
	}
	return m, nil
}
func (m model) visibleRows() []int {

	isVisible := make(map[int]bool)

	for i := range len(m.windows) {
		isVisible[i] = true
	}

	if m.filterText == "" {
		visible := make([]int, len(m.windows))
		for i := range len(m.windows) {
			visible[i] = i
		}
		return visible
	}

	for s := range strings.SplitSeq(m.filterText, "/") {
		for i, title := range m.windows {
			isVisible[i] = isVisible[i] && strings.Contains(title, s)
		}
	}

	var visible []int

	for i := range len(isVisible) {
		if isVisible[i] {
			visible = append(visible, i)
		}
	}

	return visible

}
func (m *model) clampCursor() {

	v := m.visibleRows()

	var found bool

	for _, i := range m.visibleRows() {
		found = i == m.cursor
	}

	if !found && len(v) > 0 {
		m.cursor = v[0]
	}

}
func (m model) updateCursor(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		for _, i := range m.visibleRows() {
			if i > m.cursor {
				m.cursor = i
				break
			}
		}
	case "k", "up":
		v := m.visibleRows()
		for i := len(v) - 1; i >= 0; i-- {
			if i < m.cursor {
				m.cursor = i
				break
			}
		}
	case "/":
		m.filtering = true
	case "enter":
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(raiseScript, m.windows[m.cursor]))
		if err := cmd.Run(); err != nil {
			m.err = fmt.Errorf("osascript: %w", err)
			return m, nil
		}
		return m, tea.Quit
	}
	return m, nil
}
func (m model) updateFilter(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch s := k.String(); s {
	case "esc":
		m.filtering = false
		m.filterText = ""
		m.cursor = 0
	case "enter", "ctrl+c":
		m.filtering = false
	case "backspace":
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
		}
		m.clampCursor()
	default:
		r := []rune(s)
		if len(r) == 1 && unicode.IsPrint(r[0]) {
			m.filterText += s
			m.clampCursor()
		}
	}
	return m, nil
}
func (m model) View() tea.View {

	b := new(strings.Builder)

	v := m.visibleRows()
	for _, i := range v {

		if m.cursor == i {
			fmt.Fprint(b, "> ")
		} else {
			fmt.Fprint(b, "  ")
		}

		fmt.Fprintln(b, m.windows[i])

	}
	if len(v) == 0 {
		fmt.Fprintln(b, "No matching result")
	}

	fmt.Fprintln(b)

	if !m.filtering {
		fmt.Fprintln(b, "/ filter  q quit")
	} else {
		fmt.Fprintf(b, "filter: ")
		fields := strings.Split(m.filterText, "/")
		for i, f := range fields {
			fields[i] = fmt.Sprintf("%q", f)
		}
		fmt.Fprintf(b, strings.Join(fields, " and "))
		fmt.Fprintln(b, "\nenter select  esc cancel  / combine", m.filterText)
	}

	return tea.NewView(b.String())

}
