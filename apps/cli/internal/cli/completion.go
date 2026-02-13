package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for taskmd.

To load completions:

Bash:
  $ source <(taskmd completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ taskmd completion bash > /etc/bash_completion.d/taskmd
  # macOS:
  $ taskmd completion bash > $(brew --prefix)/etc/bash_completion.d/taskmd

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. Execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ taskmd completion zsh > "${fpath[1]}/_taskmd"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ taskmd completion fish | source

  # To load completions for each session, execute once:
  $ taskmd completion fish > ~/.config/fish/completions/taskmd.fish

PowerShell:
  PS> taskmd completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, add the output to your profile:
  PS> taskmd completion powershell >> $PROFILE`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(_ *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return rootCmd.GenBashCompletion(os.Stdout)
	case "zsh":
		return rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		return rootCmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		return nil
	}
}
