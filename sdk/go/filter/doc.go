// Package filter applies field-based filters to task collections.
//
// Filters use a "field<op>value" expression syntax with AND logic across
// multiple criteria. The default operator is "=" (exact match). The
// priority and effort fields also support ordering operators: >, >=, <, <=.
//
// Supported fields include status, priority, effort, type, group, tags,
// and assignee.
//
// The effort vocabulary is project-configurable, so Apply takes an
// [github.com/driangle/taskmd/sdk/go/effort.Scale] that supplies both the
// accepted values and their order. Pass the zero Scale for the default
// small, medium, large.
package filter
