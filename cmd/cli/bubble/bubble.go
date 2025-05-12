package bubble

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/WhiCu/TCommit/internal/cli/bubble"
	"github.com/WhiCu/TCommit/internal/core/template"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bubbleCmd = &cobra.Command{
	Use:   "bubble",
	Short: "Start interactive commit message editor",
	Long: `Start an interactive TUI editor for creating commit messages.
This mode allows you to fill in template variables interactively.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Open and parse template file
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("failed to open template file: %w", err)
		}
		defer file.Close()

		tmpl, err := template.Parse(file)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		fileName := filepath.Base(args[0])

		replace := map[string]string{}
		// Create and run the program
		program := bubble.NewProgram(fileName, tmpl, replace)
		if _, err := program.Run(); err != nil {
			return fmt.Errorf("program error: %w", err)
		}

		message, err := tmpl.Execute(template.ReplacerFuncFromMap(replace))
		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		if viper.GetBool("execute") {
			viper.Set("message", message)
		}

		return nil
	},
}

// GetCommand returns the interactive command
func GetCommand() *cobra.Command {
	return bubbleCmd
}
