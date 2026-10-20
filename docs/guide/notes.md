# Notes

A board of notes with replies, for the people on the instance and for the
agent. Name someone with `@name`: the note shows the name marked, and
`list_notes` over [MCP](mcp.md) finds it for that user.

## Instructions to the agent

A note naming the agent, `@claude` on most instances, is an instruction,
and so is a reply that names it. The note carries the state on its left
edge: **open**, **working**, then **done** or **failed**.

- With an agent job configured, task-ui starts that job as the agent's
  user the moment the note or the reply appears, with the whole thread in
  its environment: `TASK_UI_NOTE`, every message of the thread under its
  author, `TASK_UI_NOTE_ID` and `TASK_UI_NOTE_AUTHOR`. The note links the
  run. The job writes its answer to the file `TASK_UI_NOTE_REPLY` names,
  and when the run ends, the agent replies with that answer alone, with a
  link to the run for the output. A job that writes no answer replies with
  how the run ended and the end of its output. One run of the job goes at
  a time; another instruction waits for it.
- The agent may decline an instruction: its reply says why, and the note
  ends there.
- Without an agent job, the note stays open until an agent reads the board
  over MCP, replies and closes it with `note_comment`.

To follow up, reply on the note and name the agent again: it works from
the thread, its own replies included. A viewer's note to the agent stays a
message.

The author of a note, or an admin, deletes it with its replies. The form
posts with **Post note** or Ctrl+Enter; a reply opens under **Reply**.
A note and a reply are markdown: `**bold**`, `*italic*`, lists, headings,
links, quotes, tables, `` `code` `` and ```` ``` ```` blocks. A line break
stays a line break, and HTML shows escaped, as text. Messages from others
arrive the moment someone sends them.

## The agent

`TASK_UI_AGENT` names the agent's user and `TASK_UI_AGENT_JOB` its job, as
`folder/job` (see [Running the server](../running.md)). The agent's user
has to exist in the password file, and its role decides what its runs may
reach. The Claude Code integration under [Settings](settings.md#integrations)
holds its token.

---

Previous: [Scheduled](scheduled.md) · Next: [Projects and jobs](projects.md) · [Contents](../README.md)
