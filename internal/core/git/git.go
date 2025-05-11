package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GitError represents a git-specific error
type GitError struct {
	Command string
	Err     error
}

func (e *GitError) Error() string {
	return fmt.Sprintf("git %s: %v", e.Command, e.Err)
}

// runGitCommand executes a git command and returns its output
func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", &GitError{
			Command: strings.Join(args, " "),
			Err:     fmt.Errorf("%s: %w", string(output), err),
		}
	}
	return strings.TrimSpace(string(output)), nil
}

// Commit executes git commit with the given message
func Commit(message string) error {
	// Validate git state before committing
	if err := ValidateGitState(); err != nil {
		return err
	}

	// Execute commit
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return &GitError{
			Command: "commit",
			Err:     err,
		}
	}

	return nil
}

// IsGitRepository checks if the current directory is a git repository
func IsGitRepository() error {
	_, err := runGitCommand("rev-parse", "--is-inside-work-tree")
	return err
}

// HasStagedChanges checks if there are any staged changes
func HasStagedChanges() (bool, error) {
	output, err := runGitCommand("diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	return output != "", nil
}

// HasUnstagedChanges checks if there are any unstaged changes
func HasUnstagedChanges() (bool, error) {
	output, err := runGitCommand("diff", "--name-only")
	if err != nil {
		return false, err
	}
	return output != "", nil
}

// GetCurrentBranch returns the name of the current branch
func GetCurrentBranch() (string, error) {
	return runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
}

// ValidateGitState checks if git is in a valid state for commit
func ValidateGitState() error {
	// Check if we're in a git repository
	if err := IsGitRepository(); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	// Check for staged changes
	hasStaged, err := HasStagedChanges()
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}
	if !hasStaged {
		return fmt.Errorf("no staged changes to commit")
	}

	// Check for unstaged changes
	hasUnstaged, err := HasUnstagedChanges()
	if err != nil {
		return fmt.Errorf("failed to check unstaged changes: %w", err)
	}
	if hasUnstaged {
		return fmt.Errorf("you have unstaged changes. Please stage them first or use --include")
	}

	// Get current branch
	branch, err := GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	if branch == "HEAD" {
		return fmt.Errorf("detached HEAD state. Please checkout a branch")
	}

	return nil
}
