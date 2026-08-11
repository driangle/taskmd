package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// effortUsageVerbs maps a command name to the leading word of its --effort flag
// usage string, preserving each command's original phrasing.
var effortUsageVerbs = map[string]string{
	"add": "task",
	"set": "new",
}

// refreshEffortUsage rewrites cmd's --effort flag usage to list the project's
// configured effort vocabulary.
//
// This exists because cobra renders help before cobra.OnInitialize runs, so the
// config file is not yet loaded when the usage string registered at init time
// would be printed. Commands register the default phrasing and this replaces it
// with the real vocabulary just before help is displayed.
func refreshEffortUsage(cmd *cobra.Command) {
	verb, ok := effortUsageVerbs[cmd.Name()]
	if !ok {
		return
	}
	flag := cmd.Flags().Lookup("effort")
	if flag == nil {
		return
	}
	flag.Usage = fmt.Sprintf("%s effort (%s)", verb, resolveEffortScale())
}
