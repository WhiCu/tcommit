package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/WhiCu/TCommit/cmd/cli/bubble"
	"github.com/WhiCu/TCommit/internal/core/git"
	"github.com/WhiCu/TCommit/internal/core/template"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Replacements map[string]string
	TemplateFile string
	ExecuteGit   bool
}

// parseReplacements parses the replacement flags into a map
func parseReplacements(replaceFlags []string) (map[string]string, error) {
	replacements := make(map[string]string)
	for _, rep := range replaceFlags {
		parts := strings.SplitN(rep, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid replacement format: %s (expected key=value)", rep)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty key in replacement: %s", rep)
		}
		replacements[key] = value
	}
	return replacements, nil
}

// processTemplate processes the template file with the given replacements
func processTemplate(cfg *Config) (string, error) {
	file, err := os.Open(cfg.TemplateFile)
	if err != nil {
		return "", fmt.Errorf("failed to open template file: %w", err)
	}
	defer file.Close()

	t, err := template.Parse(file)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	replacer := template.ReplacerFuncFromMap(cfg.Replacements)

	// Use strings.Builder to capture the output
	var output strings.Builder
	if err := t.ExecuteTo(&output, replacer); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return output.String(), nil
}

var rootCmd = &cobra.Command{
	Use:   "tcommit",
	Short: "Template-based commit message generator",
	Long: `TCommit is a tool for generating commit messages from templates.
It supports variable substitution and conditional formatting.

You can provide replacements in two ways:
	1. Using --replace flag: --replace key=value

Examples:
	tcommit template.txt --replace type=feat --replace scope=auth
	tcommit template.txt --replace type=feat --replace scope=auth --execute`,
	Args: cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		replaceFlags := viper.GetStringSlice("replace")
		replacements, err := parseReplacements(replaceFlags)
		if err != nil {
			return fmt.Errorf("invalid replacements: %w", err)
		}
		viper.Set("replacements", replacements)
		viper.Set("message", "")

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := &Config{
			Replacements: viper.GetStringMapString("replacements"),
			TemplateFile: args[0],
		}

		message, err := processTemplate(cfg)
		if err != nil {
			return err
		}

		// Print the message
		fmt.Println(message)

		// Execute git commit if requested
		if cfg.ExecuteGit {

		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if !viper.GetBool("execute") {
			return nil
		}
		if err := git.ValidateGitState(); err != nil {
			return fmt.Errorf("git validation failed: %w", err)
		}

		if err := git.Commit(viper.GetString("message")); err != nil {
			return fmt.Errorf("failed to execute git commit: %w", err)
		}

		return nil
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringSliceP("replace", "r", []string{},
		"Replacements in format key=value (can be specified multiple times)")

	rootCmd.PersistentFlags().BoolP("execute", "e", false,
		"Execute git commit with the generated message")

	if err := viper.BindPFlag("replace", rootCmd.Flags().Lookup("replace")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding flag: %v\n", err)
		os.Exit(1)
	}

	if err := viper.BindPFlag("execute", rootCmd.PersistentFlags().Lookup("execute")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding flag: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(bubble.GetCommand())
}
