package main

import (
	"fmt"
	"os"

	"github.com/driangle/md-task-tracker/apps/cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
