package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var manCmd = &cobra.Command{
	Use:    "man [output-dir]",
	Short:  "Generate man pages",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE:   runMan,
}

func init() {
	rootCmd.AddCommand(manCmd)
}

func runMan(_ *cobra.Command, args []string) error {
	header := &doc.GenManHeader{
		Title:   "TASKMD",
		Section: "1",
		Date:    &time.Time{},
	}

	if err := doc.GenManTree(rootCmd, header, args[0]); err != nil {
		return fmt.Errorf("failed to generate man pages: %w", err)
	}
	fmt.Printf("Man pages generated in %s\n", args[0])
	return nil
}
