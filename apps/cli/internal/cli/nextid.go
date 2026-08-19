package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/nextid"
	"github.com/driangle/taskmd/sdk/go/validator"
)

var nextIDFormat string

var nextIDCmd = &cobra.Command{
	Use:   "next-id [directory]",
	Short: "Show the next available task ID",
	Long: `Next-id scans task files and outputs the next available sequential ID.

It finds the highest numeric ID among existing tasks and returns max + 1,
preserving any common prefix and zero-padding. Archived tasks are included,
so an archived task's ID is never handed out again.

By default outputs just the ID string (ideal for scripting). Use --format json
for structured output with additional metadata.

Examples:
  taskmd next-id
  taskmd next-id ./tasks/cli
  taskmd next-id --format json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNextID,
}

func init() {
	rootCmd.AddCommand(nextIDCmd)

	nextIDCmd.Flags().StringVar(&nextIDFormat, "format", "plain", "output format (plain, json)")
}

func runNextID(_ *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	ids, err := scanIDPool(scanDir, flags)
	if err != nil {
		return err
	}

	cfg := resolveIDConfig()

	switch nextIDFormat {
	case "plain":
		id, genErr := generateID(ids, cfg)
		if genErr != nil {
			return genErr
		}
		fmt.Println(id)
		return nil
	case "json":
		return outputNextIDJSON(ids, cfg)
	default:
		return ValidateFormat(nextIDFormat, []string{"plain", "json"})
	}
}

func outputNextIDJSON(ids []string, cfg validator.IDConfig) error {
	switch cfg.Strategy {
	case "prefixed":
		id := nextid.GeneratePrefixed(ids, cfg.Prefix, cfg.Padding)
		return WriteJSON(os.Stdout, nextid.Result{
			NextID:  id,
			Prefix:  cfg.Prefix,
			Padding: cfg.Padding,
			Total:   len(ids),
		})
	case "random":
		id, err := nextid.GenerateRandom(ids, cfg.Length)
		if err != nil {
			return err
		}
		return WriteJSON(os.Stdout, nextid.Result{
			NextID: id,
			Total:  len(ids),
		})
	case "ulid":
		id, err := nextid.GenerateULID(ids, cfg.Length)
		if err != nil {
			return err
		}
		return WriteJSON(os.Stdout, nextid.Result{
			NextID: id,
			Total:  len(ids),
		})
	default:
		return WriteJSON(os.Stdout, nextid.Calculate(ids))
	}
}
