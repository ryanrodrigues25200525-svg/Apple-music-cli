package cmd

import (
	"fmt"
	"os"

	"github.com/ryanrodrigues25200525-svg/Apple-music-cli/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mu",
	Short: "Muse is a CLI and TUI controller for Apple Music on macOS",
	Long:  `A Bubble Tea-powered terminal controller for Apple Music. Controls playback, queries current track, and manages library playlists.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Run interactive TUI
		if err := tui.RunTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
