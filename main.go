package main

import (
	"os"

	"github.com/Seinarukiro2/pipepie/cmd"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	// Support NO_COLOR standard (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		lipgloss.SetHasDarkBackground(false)
		os.Setenv("TERM", "dumb")
	}
	cmd.Execute()
}
