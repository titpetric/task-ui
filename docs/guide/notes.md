# Notes and Chat

Two places to leave messages for the people on the instance, and for the
agent: **Notes**, a board of notes with replies, and **Chat**, one channel
for everyone signed in.

## Notes

A note is a message with replies under it. Name someone with `@name`: the
note shows the name marked, and `list_notes` over [MCP](mcp.md) finds it
for that user.

A note naming the agent, `@claude` on most instances, is an
instruction. Its left edge carries its state: **open**, **working**, then
**done** or **failed**.

- With an agent job configured, task-ui starts that job as the agent's
  user the moment the note appears, with the note in its environment
  (`TASK_UI_NOTE`, `TASK_UI_NOTE_ID`, `TASK_UI_NOTE_AUTHOR`). The note
  links the run. When the run ends, the agent replies with how it ended
  and the end of its output. One run of the job goes at a time; a second
  instruction waits for the first.
- Without an agent job, the note stays open until an agent reads the board
  over MCP, replies and closes it with `note_comment`.

A change to finished work is a new note, not a reply: a reply is
conversation, and only a new note starts the agent again. A viewer's note
to the agent stays a message, and so does a note in the chat.

The author of a note, or an admin, deletes it with its replies. The form
posts with **Post note** or Ctrl+Enter; a reply opens under **Reply**.
`` `code` `` and ```` ``` ```` blocks show as code.

## Chat

One channel, the latest messages last, the way a chat room reads. Enter
sends, Shift+Enter starts a new line. Messages from others arrive the
moment someone sends them, on both pages.

## The agent

`TASK_UI_AGENT` names the agent's user and `TASK_UI_AGENT_JOB` its job, as
`folder/job` (see [Running the server](../running.md)). The agent's user
has to exist in the password file, and its role decides what its runs may
reach.

---

Previous: [Scheduled](scheduled.md) · Next: [Projects and jobs](projects.md) · [Contents](../README.md)
