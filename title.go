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
