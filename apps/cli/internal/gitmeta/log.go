package gitmeta

// Debugf and Warnf are logging hooks, no-ops by default. The CLI wires them
// to its --debug and --verbose output at startup; keeping them as injection
// points lets this package stay free of flag and viper state.
var (
	Debugf = func(_ string, _ ...any) {}
	Warnf  = func(_ string, _ ...any) {}
)
