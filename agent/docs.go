// Package agent holds task-ui's host commands, the commands of the Claude
// built-ins every folder has: Claude Code on a note from the Notes board,
// @claude-note, on a prompt, @claude-prompt, and in a terminal, @claude.
// The server runs them as `task-ui agent note|prompt|shell`.
//
// The board starts the note command as the agent's user for a note or a
// reply naming the agent. The thread comes in TASK_UI_NOTE, with
// TASK_UI_NOTE_ID and TASK_UI_NOTE_AUTHOR, and Claude's final message
// says what changed, or why it declined. It works in TASK_UI_WORKSPACE_DIR,
// the default workspace with the projects, with the context in its system
// prompt. A workspace's home, TASK_UI_HOME, is a directory every Claude
// session here reads, and where a workspace's prompt works.
//
// The shell command is a development tool with someone at the keyboard.
package agent
