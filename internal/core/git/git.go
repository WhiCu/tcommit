package git

import (
	"os"
	"os/exec"
)

func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	// Перенаправляем стандартные потоки
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
