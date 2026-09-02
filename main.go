package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func run() error {
	script := `tell application "System Events" to get name of every window of process "Alacritty"`
	{
		buf := new(bytes.Buffer)
		cmd := exec.Command("osascript", "-e", script)
		cmd.Stderr = os.Stderr
		cmd.Stdout = buf
		if err := cmd.Run(); err != nil {
			return err
		}
		for token := range strings.SplitSeq(buf.String(), ", ") {
			fmt.Println(token)
		}
	}
	return nil
}
