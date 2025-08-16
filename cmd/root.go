package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "coolify-cli",
	Short: "A CLI tool for interacting with Coolify API",
	Long: `Coolify CLI is a command-line interface for interacting with your Coolify instance.
It allows you to manage applications, view logs, and perform various operations
through the Coolify API.`,
	Version: "1.0.0",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global flags here if needed
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Customize help template
	rootCmd.SetHelpTemplate(`{{.Long}}

Usage:
  {{.UseLine}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}

// checkConfigAndConnection is a helper function to validate config and connection
func checkConfigAndConnection() error {
	// This can be used by commands that need to ensure the API is accessible
	return nil
}
