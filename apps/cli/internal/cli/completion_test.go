package cli

import (
	"testing"
)

func TestCompletion_AllShells(t *testing.T) {
	repo := newTaskRepo(t, nil)
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			res := repo.Run("completion", shell)
			if res.Err != nil {
				t.Fatalf("runCompletion(%q) returned error: %v", shell, res.Err)
			}

			if res.Stdout == "" {
				t.Errorf("runCompletion(%q) produced no output", shell)
			}
		})
	}
}
