# Collapse home directory to `~` in window titles — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use supervised-plan-execution to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When displaying window titles in the interactive picker, collapse a leading `$HOME` path prefix to `~`, while keeping all logic (`-title` matching, history, recency ranking) operating on the original title.

**Architecture:** One new display-only helper `collapseHome(title string) string` in a new `title.go`, backed by an injectable `collapseHomeWith(title, home string)` for tests. The picker's `View()` renders `collapseHome(...)` of each title; every other code path (`-title` lookup in `main.go:104-127`, `rank` map built from history at `main.go:133-138`, sorting, `raiseAndExit` at `main.go:293-306`, `history.Log`) keeps using `m.windows` untouched.

**Tech Stack:** Go stdlib only (`os`, `strings`), `go test`, `jj` VCS (see `core-commands`: no git commands, commits via `jj describe --stdin < TMPFILE`).

**Issue:** `docs/issues/17.md`

---

### Task 1: Title collapsing helper with tests

**Files:**
- Create: `title.go`
- Test: `title_test.go`

- [ ] **Step 1: Write the failing test**

Create `title_test.go`. Tests pass `$HOME` explicitly via `collapseHomeWith` so they don't depend on the real username:

```go
package main

import "testing"

func TestCollapseHomeWith(t *testing.T) {
	const home = "/Users/bob"
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "home then path collapses",
			title: "crush /Users/bob/some/path",
			want:  "crush ~/some/path",
		},
		{
			name:  "bare home collapses",
			title: "/Users/bob",
			want:  "~",
		},
		{
			name:  "home with trailing slash collapses",
			title: "crush /Users/bob/ ",
			want:  "crush ~/ ",
		},
		{
			name:  "no home prefix unchanged",
			title: "crush /var/log",
			want:  "crush /var/log",
		},
		{
			name:  "different user path unchanged",
			title: "crush /Users/carol/x",
			want:  "crush /Users/carol/x",
		},
		{
			name:  "home mid-string unchanged",
			title: "job=a, dir=/Users/bob/x",
			want:  "job=a, dir=/Users/bob/x",
		},
		{
			name:  "prefix of home unchanged",
			title: "/Users/bobby/x",
			want:  "/Users/bobby/x",
		},
		{
			name:  "empty title",
			title: "",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collapseHomeWith(tt.title, home); got != tt.want {
				t.Fatalf("collapseHomeWith(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestCollapseHomeUsesRealHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("no home dir available")
	}
	got := collapseHome(home + "/x/y")
	want := "~/x/y"
	if got != want {
		t.Fatalf("collapseHome(%q) = %q, want %q", home+"/x/y", got, want)
	}
}
```

Note: `TestCollapseHomeUsesRealHome` needs `import ("os"; "testing")` — adjust the import block accordingly:

```go
import (
	"os"
	"testing"
)
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./... -run TestCollapseHome -v`
Expected: FAIL — `collapseHomeWith` and `collapseHome` are undefined.

- [ ] **Step 3: Write the minimal implementation**

Create `title.go`:

```go
package main

import (
	"os"
	"strings"
)

func collapseHome(title string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return title
	}
	return collapseHomeWith(title, home)
}

// collapseHomeWith replaces the leading path of a title equal to home with
// "~". A home-dir occurrence is only collapsed at the start of the title or
// right after whitespace, and only when the escaped text is "" or starts
// with "/", so "/Users/bobby/x" and "dir=/Users/bob/x" stay untouched.
func collapseHomeWith(title, home string) string {
	var b strings.Builder
	i := 0
	for i < len(title) {
		atSegmentStart := i == 0 || title[i-1] == ' ' || title[i-1] == '\t'
		if atSegmentStart && strings.HasPrefix(title[i:], home) {
			rest := title[i+len(home):]
			if rest == "" || rest[0] == '/' {
				b.WriteString("~")
				i += len(home)
				continue
			}
		}
		b.WriteByte(title[i])
		i++
	}
	return b.String()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./... -run TestCollapseHome -v`
Expected: PASS (all subtests).

- [ ] **Step 5: Commit with caveman-commit**

Load `caveman-commit` for the message, then use the temp file pattern from `core-commands`:

```bash
mktemp
```

Write the commit message with the Write tool, then:

```bash
jj describe --stdin < "$TMPFILE" && rm "$TMPFILE" && jj new
```

---

### Task 2: Render collapsed titles in the picker

**Files:**
- Modify: `main.go:353` (inside `func (m model) View()`)

- [ ] **Step 1: Write the failing test (display transformation)**

Since `View()` is a thin method on the model, cover the wiring indirectly with a unit test on the render path by adding to `title_test.go`:

```go
func TestViewShowsCollapsedTitle(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("no home dir available")
	}
	m := model{windows: []string{"crush " + home + "/some/path"}}
	out := m.View().String()
	want := "crush ~/some/path"
	if !strings.Contains(out, "> "+want) {
		t.Fatalf("View() output does not contain cursor line with %q:\n%s", want, out)
	}
	if strings.Contains(out, home) {
		t.Fatalf("View() leaks raw home path:\n%s", out)
	}
}
```

(`strings` must be added to the import block of `title_test.go`.)

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./... -run TestViewShowsCollapsedTitle -v`
Expected: FAIL — View still prints the raw title.

- [ ] **Step 3: Wire the display transformation**

In `main.go`, `View()` loop (currently `main.go:345-355`):

```go
	v := m.visibleRows()
	for _, i := range v {

		if m.cursor == i {
			fmt.Fprint(b, "> ")
		} else {
			fmt.Fprint(b, "  ")
		}

		fmt.Fprintln(b, collapseHome(m.windows[i]))

	}
```

Only that one line changes: `fmt.Fprintln(b, m.windows[i])` becomes `fmt.Fprintln(b, collapseHome(m.windows[i]))`. Selection, filter matching (`strings.Contains(title, s)` at `main.go:226`), `-title` exact match (`main.go:108`), and `history.Log` (`main.go:296`) all keep reading the original `m.windows[i]`.

- [ ] **Step 4: Run all tests to verify they pass**

Run: `go test ./... -v`
Expected: PASS, including `TestViewShowsCollapsedTitle`, and full `go build ./...` succeeds.

- [ ] **Step 5: Verify acceptance criteria from the issue manually**

Run: `go build -o /tmp/atty && echo "manual check:" && /tmp/atty -title "<no such title>" 2>&1; test ! $? -eq 0 && echo "flag path still uses original titles"`

Manual check in a real window: title a shell with `printf '\033]0;atty /Users/%s/foo\007' "$USER"`, launch `tmp/atty` — entry must render as `atty ~/foo`; selecting it raises the original window; `-title "atty /Users/<you>/foo"` still resolves the original title.

- [ ] **Step 6: Commit with caveman-commit**

Load `caveman-commit` for the message; commit via temp file pattern:

```bash
jj describe --stdin < "$TMPFILE" && rm "$TMPFILE" && jj new
```

---

### Task 3: Update CHANGELOG

**Files:**
- Modify: `CHANGELOG.md` (add a bullet under the existing unreleased-style list near `CHANGELOG.md:12`)

- [ ] **Step 1: Add changelog entry**

Follow the format of existing entries (`CHANGELOG.md:12` `- `-title` flag: …`):

```markdown
- Window titles in the picker now display a leading home directory as `~`;
  `-title` matching, filtering, and history still use the original title
```

- [ ] **Step 2: Run the whole test suite and build**

Run: `go test ./... && go build ./...`
Expected: PASS, no lint regressions (`go vet ./...` clean too).

- [ ] **Step 3: Commit with caveman-commit**

Load `caveman-commit` for the message; commit via temp file pattern:

```bash
jj describe --stdin < "$TMPFILE" && rm "$TMPFILE" && jj new
```

---

## Self-review notes

- **Spec coverage:** display transformation (Task 1+2), leading-segment-only rule (`collapseHomeWith` guard `rest == "" || rest[0] == '/'` combined with whitespace-start check), original title preserved for `-title`/history/rank (no code path reads through `collapseHome` except `View`), `$HOME`-based not hardcoded (`os.UserHomeDir()`). All five acceptance criteria map to tasks.
- **Type consistency:** `collapseHome(title string) string` and `collapseHomeWith(title, home string) string` used consistently in Tasks 1-2; model fields untouched.
- **No placeholders:** all code shown in full above.
