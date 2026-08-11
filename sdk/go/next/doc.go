// Package next scores and ranks actionable tasks to recommend what to work on.
//
// Tasks are scored based on priority, critical-path position, downstream
// impact, and effort. Only tasks whose dependencies are satisfied are considered.
//
// Effort points scale with the value's position in the project's configured
// vocabulary (an [github.com/driangle/taskmd/sdk/go/effort.Scale]): the lowest
// value earns the most and the highest earns none. The lowest value is also what
// Options.QuickWins selects.
package next
