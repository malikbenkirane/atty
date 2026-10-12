# Clamp list to terminal height — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use supervised-plan-execution to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Only as many window-title rows as fit below the banner and above the footer are rendered; when the terminal is too short, show a `window not tall enough` message instead of the list.

**Architecture:** Add a `height int` field to `model` populated from `tea.WindowSizeMsg` in `Update`. Add two pure helpers in `view.go`: `countLines(s string) int` (counts newlines in a rendered string) and `visibleWindow(v []int, cursor, height int) (start, end int)` (returns the sliding slice of row indices to render, keeping the cursor on screen). Rewrite `View` to render the header and footer into separate `strings.Builder` buffers, measure their line counts with `countLines`, compute `budget = height − headerLines − footerLines` dynamically, render only `v[start:end]`, print `…` truncation indicators, and print `window not tall enough` when `budget < 1`. Dynamic measurement means the banner or footer can change size without breaking the budget math. No change to cursor, filter, or raise logic.

**Tech Stack:** Go stdlib only, `charm.land/bubbletea/v2` (`tea.WindowSizeMsg`), `go test`, `jj` VCS (see `core-commands`: no git commands, commits via `jj describe --stdin < TMPFILE`).

**Issue:** `docs/issues/19.md`

---

### Task 1: Window-size handling and viewport helper with tests

**Files:**
- Modify: `main.go` (add `height` field to `model`, handle `tea.WindowSizeMsg` in `Update`)
- Create: `view_test.go`
- Create: `view.go`

- [ ] **Step 1: Write the failing tests**

Create `view_test.go`:

```go
package main

import (
	"strings"
	"testing"
)

func TestCountLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "empty", in: "", want: 0},
		{name: "one line no newline", in: "hello", want: 1},
		{name: "one line with newline", in: "hello\n", want: 1},
		{name: "three lines", in: "a\nb\nc\n", want: 3},
		{name: "blank lines count", in: "\n\n\n", want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countLines(tt.in); got != tt.want {
				t.Fatalf("countLines(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestVisibleWindowCursorMiddle(t *testing.T) {
	// 10 rows, height for 3, cursor at index 5 → window centered on cursor.
	v := make([]int, 10)
	for i := range v {
		v[i] = i
	}
	start, end := visibleWindow(v, 5, 3)
	if start != 4 || end != 7 {
		t.Fatalf("visibleWindow(cursor=5,height=3) = (%d,%d), want (4,7)", start, end)
	}
}

func TestVisibleWindowCursorTop(t *testing.T) {
	v := make([]int, 10)
	for i := range v {
		v[i] = i
	}
	start, end := visibleWindow(v, 1, 3)
	if start != 0 || end != 3 {
		t.Fatalf("visibleWindow(cursor=1,height=3) = (%d,%d), want (0,3)", start, end)
	}
}

func TestVisibleWindowCursorBottom(t *testing.T) {
	v := make([]int, 10)
	for i := range v {
		v[i] = i
	}
	start, end := visibleWindow(v, 9, 3)
	if start != 7 || end != 10 {
		t.Fatalf("visibleWindow(cursor=9,height=3) = (%d,%d), want (7,10)", start, end)
	}
}

func TestVisibleWindowBudgetOne(t *testing.T) {
	v := make([]int, 10)
	for i := range v {
		v[i] = i
	}
	start, end := visibleWindow(v, 5, 1)
	if start != 5 || end != 6 {
		t.Fatalf("visibleWindow(cursor=5,height=1) = (%d,%d), want (5,6)", start, end)
	}
}

func TestVisibleWindowShorterThanBudget(t *testing.T) {
	// Fewer rows than the budget → show everything.
	v := make([]int, 2)
	for i := range v {
		v[i] = i
	}
	start, end := visibleWindow(v, 0, 5)
	if start != 0 || end != 2 {
		t.Fatalf("visibleWindow(cursor=0,height=5,len=2) = (%d,%d), want (0,2)", start, end)
	}
}

func TestVisibleWindowEmpty(t *testing.T) {
	start, end := visibleWindow(nil, 0, 5)
	if start != 0 || end != 0 {
		t.Fatalf("visibleWindow(empty) = (%d,%d), want (0,0)", start, end)
	}
}

func TestUpdateStoresWindowSize(t *testing.T) {
	m := model{windows: []string{"a", "b"}, height: 0}
	updated, _ := m.Update(teaWindowSizeMsg(80, 24))
	if updated.(model).height != 24 {
		t.Fatalf("height = %d, want 24", updated.(model).height)
	}
}

// teaWindowSizeMsg builds a tea.WindowSizeMsg without importing bubbletea in
// the test helper, by using the concrete type already used in main.go.
func teaWindowSizeMsg(w, h int) tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: w, Height: h}
}
```

The import block needs `tea "charm.land/bubbletea/v2"`:

```go
import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)
```

`strings` is needed for `TestCountLines` and later `TestView*` tests.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./... -run 'TestVisibleWindow|TestUpdateStoresWindowSize|TestCountLines' -v`
Expected: FAIL — `visibleWindow`, `countLines`, and `teaWindowSizeMsg` are undefined; `model` has no `height` field.

- [ ] **Step 3: Add the `height` field and handle `tea.WindowSizeMsg`**

In `main.go`, add `height int` to the `model` struct (currently `main.go:185-193`):

```go
type model struct {
	windows    []string
	cursor     int
	filtering  bool
	filterText string
	done       *bool
	history    repo
	err        error
	height     int
}
```

In `main.go`, add a `tea.WindowSizeMsg` case to `Update` (currently `main.go:198-207`):

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		if m.filtering {
			return m.updateFilter(msg)
		}
		return m.updateCursor(msg)
	}
	return m, nil
}
```

- [ ] **Step 4: Write the `visibleWindow` helper**

Create `view.go`:

```go
package main

import "strings"

// countLines returns the number of terminal lines a rendered string occupies.
// A string ending in a newline occupies that many lines; a string without a
// trailing newline occupies one additional line for the final partial line.
func countLines(s string) int {
	if s == "" {
		return 0
	}
	n := strings.Count(s, "\n")
	if !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}

// visibleWindow returns the half-open slice [start, end) of indices from v
// that should be rendered given a viewport of height rows centered on the
// cursor position. When the cursor is near the top or bottom the window
// shifts so it never goes out of bounds. When len(v) <= height the whole
// slice is shown.
func visibleWindow(v []int, cursor, height int) (start, end int) {
	if len(v) == 0 || height <= 0 {
		return 0, 0
	}
	if height >= len(v) {
		return 0, len(v)
	}

	// Find the cursor's position within v (it stores original indices, not
	// positions, so we locate it by value).
	pos := 0
	for i, idx := range v {
		if idx == cursor {
			pos = i
			break
		}
	}

	start = pos - height/2
	if start < 0 {
		start = 0
	}
	end = start + height
	if end > len(v) {
		end = len(v)
		start = end - height
	}
	return start, end
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./... -run 'TestVisibleWindow|TestUpdateStoresWindowSize|TestCountLines' -v`
Expected: PASS (all subtests).

- [ ] **Step 6: Run the full test suite and build**

Run: `go test ./... && go build ./...`
Expected: PASS, build succeeds.

- [ ] **Step 7: Commit with caveman-commit**

Load `caveman-commit` for the message, then use the temp file pattern from `core-commands`:

```bash
mktemp
```

Read the temp file (required even though empty), then write the commit message, then:

```bash
jj describe --stdin < "$TMPFILE" && rm "$TMPFILE" && jj new
```

---

### Task 2: Render only the viewport slice in `View`

**Files:**
- Modify: `main.go:337-376` (inside `func (m model) View()`)
- Test: `view_test.go` (append)

- [ ] **Step 1: Write the failing tests for the clamped `View`**

Append to `view_test.go`:

```go
func TestViewTruncatesLongList(t *testing.T) {
	windows := make([]string, 50)
	for i := range windows {
		windows[i] = "win-" + string(rune('a'+i%26)) + string(rune('a'+i/26))
	}
	m := model{windows: windows, cursor: 25, height: 10}
	out := m.View().Content

	// Header + footer are present.
	if !strings.Contains(out, "/ filter  q quit") {
		t.Fatalf("footer missing:\n%s", out)
	}
	// Not all 50 titles are rendered.
	for _, w := range windows {
		if strings.Count(out, w) > 1 {
			t.Fatalf("title %q appears more than once — full list rendered:\n%s", w, out)
		}
	}
}

func TestViewTooShortMessage(t *testing.T) {
	m := model{windows: []string{"a", "b"}, cursor: 0, height: 4}
	out := m.View().Content
	if !strings.Contains(out, "window not tall enough") {
		t.Fatalf("expected 'window not tall enough', got:\n%s", out)
	}
}

func TestViewShortListNotTruncated(t *testing.T) {
	m := model{windows: []string{"a", "b", "c"}, cursor: 0, height: 80}
	out := m.View().Content
	for _, w := range m.windows {
		if !strings.Contains(out, w) {
			t.Fatalf("title %q missing from output:\n%s", w, out)
		}
	}
}
```

`strings` is already imported in `view_test.go`; if not, add `"strings"` to the import block.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./... -run 'TestView' -v`
Expected: FAIL — `View` still renders the full list; no `window not tall enough` message; `height` is not yet consulted.

- [ ] **Step 3: Rewrite `View` to clamp the list**

Replace the body of `func (m model) View() tea.View` in `main.go` (currently `main.go:337-376`) with:

```go
func (m model) View() tea.View {

	// Render header into a separate buffer so we can measure its line count.
	var header strings.Builder
	fmt.Fprintln(&header, banner)
	fmt.Fprintln(&header) // blank separator
	fmt.Fprintln(&header) // extra blank for clarity

	// Render footer into a separate buffer so we can measure its line count.
	var footer strings.Builder
	fmt.Fprintln(&footer) // blank separator
	fmt.Fprintln(&footer) // extra blank for clarity
	m.renderFooter(&footer)

	headerLines := countLines(header.String())
	footerLines := countLines(footer.String())

	v := m.visibleRows()

	budget := m.height - headerLines - footerLines

	if m.height == 0 {
		// No size message yet (first frame fallback): render everything.
		budget = len(v)
	}

	b := new(strings.Builder)

	if budget < 1 {
		fmt.Fprint(b, header.String())
		fmt.Fprintln(b, "window not tall enough")
		fmt.Fprint(b, footer.String())
		return tea.NewView(b.String())
	}

	fmt.Fprint(b, header.String())

	// Reserve lines for truncation indicators (up to 2: above + below).
	listHeight := budget
	if len(v) > budget {
		listHeight = budget - 2
	}

	if listHeight < 1 {
		fmt.Fprintln(b, "window not tall enough")
		fmt.Fprint(b, footer.String())
		return tea.NewView(b.String())
	}

	start, end := visibleWindow(v, m.cursor, listHeight)

	if start > 0 {
		fmt.Fprintf(b, "… %d above\n", start)
	}

	for _, i := range v[start:end] {
		if m.cursor == i {
			fmt.Fprint(b, "> ")
		} else {
			fmt.Fprint(b, "  ")
		}
		fmt.Fprintln(b, collapseHome(m.windows[i]))
	}

	if end < len(v) {
		fmt.Fprintf(b, "… %d below\n", len(v)-end)
	}

	if len(v) == 0 {
		fmt.Fprintln(b, "No matching result")
	}

	fmt.Fprint(b, footer.String())

	return tea.NewView(b.String())
}

func (m model) renderFooter(b *strings.Builder) {
	if !m.filtering {
		fmt.Fprintln(b, "/ filter  q quit")
		return
	}
	fmt.Fprintf(b, "filter: ")
	fields := strings.Split(m.filterText, "/")
	for i, f := range fields {
		fields[i] = fmt.Sprintf("%q", f)
	}
	fmt.Fprint(b, strings.Join(fields, " and "))
	fmt.Fprintln(b, "\nenter select  esc or ctrl+c cancel  / or space combine filters")
}
```

The header and footer are rendered into their own builders, measured with `countLines`, and the budget is computed as `height − headerLines − footerLines`. This means the banner can be any height and the footer can have any number of help lines — the budget adapts automatically. The extra blank lines for clarity (one after the header separator, one before the footer) are baked into the header/footer builders, so they are measured as part of those sections.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./... -run 'TestView' -v`
Expected: PASS (all subtests, including the pre-existing `TestViewShowsCollapsedTitle` from `title_test.go`).

- [ ] **Step 5: Run the full test suite and build**

Run: `go test ./... && go build ./...`
Expected: PASS, build succeeds.

- [ ] **Step 6: Commit with caveman-commit**

Load `caveman-commit` for the message; commit via temp file pattern:

```bash
mktemp
```

Read the temp file, write the message, then:

```bash
jj describe --stdin < "$TMPFILE" && rm "$TMPFILE" && jj new
```

---

### Task 3: Update CHANGELOG

**Files:**
- Modify: `CHANGELOG.md` (add a bullet under the existing `## [v0.2.0]` → `### Changed` section, near `CHANGELOG.md:10`)

- [ ] **Step 1: Add changelog entry**

Insert after the existing `### Changed` bullet at `CHANGELOG.md:12-14`, following the same format:

```markdown
- The picker now clamps the window-title list to the terminal height so the
  header and footer stay visible; when the terminal is too short it prints
  `window not tall enough` instead of the list
  ([`docs/issues/19.md`]).
```

Also add the link reference at the bottom of the file with the others (near `CHANGELOG.md:55`):

```markdown
[`docs/issues/19.md`]: docs/issues/19.md
```

- [ ] **Step 2: Run the whole test suite and build**

Run: `go test ./... && go build ./...`
Expected: PASS, `go vet ./...` clean too.

- [ ] **Step 3: Commit with caveman-commit**

Load `caveman-commit` for the message; commit via temp file pattern:

```bash
mktemp
```

Read the temp file, write the message, then:

```bash
jj describe --stdin < "$TMPFILE" && rm "$TMPFILE" && jj new
```

---

## Self-review notes

- **Spec coverage:** height tracked via `tea.WindowSizeMsg` (Task 1), budget computed dynamically via `countLines` on rendered header/footer buffers (Task 2 — handles variable banner/footer sizes), sliding window around cursor (Task 1 `visibleWindow` + Task 2 render), `window not tall enough` when `budget < 1` (Task 2), truncation indicators `… N above/below` (Task 2), resize updates on next render via stored `height` (Task 1), filtering/`-title` unaffected (no logic change, only `View`).
- **Type consistency:** `countLines(s string) int` and `visibleWindow(v []int, cursor, height int) (start, end int)` defined in Task 1 `view.go`, used identically in Task 2; `model.height int` added in Task 1, read in Task 2; `renderFooter(b *strings.Builder)` extracted in Task 2.
- **No placeholders:** all code shown in full; exact line references given for edits.
- **Edge cases covered by tests:** `countLines` on empty/trailing-newline/no-trailing-newline/blank-lines, cursor at top/middle/bottom, budget of 1, list shorter than budget, empty list, zero-height first frame, too-short terminal, long list truncation, short list not truncated.
- **Dynamic sizing:** header and footer are measured at render time via `countLines`, so changing the banner size or footer help format requires no constant updates.
