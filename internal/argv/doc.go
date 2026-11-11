// Package argv reads the arguments a user types for a job into a list,
// the way a shell splits words and nothing more: quotes and backslashes
// group, and no variable, glob, pipe or command substitution expands.
// The list reaches the job as argv, never through a shell.
//
// Outside quotes, blanks separate arguments and a backslash takes the
// next character literally. Single quotes take everything up to the next
// one literally; double quotes do too, except that a backslash before "
// or \ takes that character.
package argv
