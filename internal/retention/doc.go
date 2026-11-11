// Package retention keeps what task-ui records within the rules of the
// Data retention settings: a table with one row per area of data, the
// recordings and the rows of runs, prompt threads, workflow runs, the
// history of the artifact tree and the secret read log, each kept for an
// age and, where items have a size, within a size, the oldest going
// first. task-ui cleanup, the root's @cleanup, applies the rules; the
// settings page measures what each area holds.
package retention
