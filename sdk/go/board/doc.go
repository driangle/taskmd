// Package board groups tasks into columns by a specified field.
//
// Supported grouping fields are status, priority, effort, type, group, and tag.
// Column ordering follows natural conventions for each field type; for effort it
// follows the project's configured vocabulary, supplied as an
// [github.com/driangle/taskmd/sdk/go/effort.Scale].
package board
