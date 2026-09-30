package main

import (
	"os"
	"strings"
	"testing"
)

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

func TestViewShowsCollapsedTitle(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("no home dir available")
	}
	m := model{windows: []string{"crush " + home + "/some/path"}}
	out := m.View().Content
	want := "crush ~/some/path"
	if !strings.Contains(out, "> "+want) {
		t.Fatalf("View() output does not contain cursor line with %q:\n%s", want, out)
	}
	if strings.Contains(out, home) {
		t.Fatalf("View() leaks raw home path:\n%s", out)
	}
}
