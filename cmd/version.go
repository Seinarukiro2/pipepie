package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// Version is set by goreleaser at build time.
var Version = "dev"

var (
	vPurple = lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")).Bold(true)
	vGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Bold(true)
	vYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#f1fa8c")).Bold(true)
	vDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4"))
	vCyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd"))
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and check for updates",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s %s %s\n", vPurple.Render("pie"), vGreen.Render(Version),
			vDim.Render(runtime.GOOS+"/"+runtime.GOARCH))

		// Check for updates
		latest, err := getLatestVersion()
		if err != nil || latest == "" {
			return
		}
		current := strings.TrimPrefix(Version, "v")
		latest = strings.TrimPrefix(latest, "v")
		if current != "dev" && current != latest {
			fmt.Println()
			fmt.Println(vYellow.Render("  Update available: ") + vGreen.Render("v"+latest))
			fmt.Println(vDim.Render("  Run: ") + vCyan.Render("pie update"))
		}
	},
}

func getLatestVersion() (string, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/Seinarukiro2/pipepie/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
