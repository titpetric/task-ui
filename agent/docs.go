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
// The prompt command answers a thread from a folder, a run or a workflow
// run. It reads:
//
//   - TASK_UI_PROMPT, the thread, oldest first, the last message the question
//   - TASK_UI_PROMPT_DIR, the directory to run in; TASK_UI_HOME or
//     TASK_UI_WORKSPACE_DIR stand in for it
//   - TASK_UI_PROMPT_REPLY, a file the reply is written to as well
//   - TASK_UI_PROMPT_RUN, a file describing the run a thread is about
//   - TASK_UI_PROMPT_WORKFLOW, the same for a workflow run's steps
//   - TASK_UI_PROMPT_INSTRUCTIONS, prepended to the system prompt
//   - TASK_UI_PROMPT_SOURCE, the thread to post the reply in over the task
//     MCP server; the final message is then only the fallback
//   - TASK_UI_HOSTS, the inventory Claude may inspect over ssh
//
// The shell command is a development tool with someone at the keyboard.
package agent
