# Notes

A board of notes with replies, for the people on the instance and for the
agent. Name someone with `@name`: the note shows the name marked, and
`list_notes` over [MCP](mcp.md) finds it for that user.

![Notes](../assets/notes-dark.png)

## Instructions to the agent

A note naming the agent, `@claude` on most instances, is an instruction,
and so is a reply that names it. The note carries the state on its left
edge: **open**, **working**, then **done** or **failed**.

- With an agent job, `internal/claude-note` unless configured otherwise,
  task-ui starts that job as the agent's user the moment the note or the
  reply appears, with the whole thread in
  its environment: `TASK_UI_NOTE`, every message of the thread under its
  author, `TASK_UI_NOTE_ID` and `TASK_UI_NOTE_AUTHOR`. The note links the
  run for the users who see it. The job writes its answer to the file `TASK_UI_NOTE_REPLY` names,
  and when the run ends, the agent replies with that answer alone, with a
  link to the run for the output. A job that writes no answer replies with
  how the run ended and the end of its output. One run of the job goes at
  a time; another instruction waits for it.
- The agent may decline an instruction: its reply says why, and the note
  ends there.
- With `TASK_UI_AGENT_JOB` set empty, the note stays open until an agent
  reads the board over MCP, replies and closes it with `note_comment`.

To follow up, reply on the note and name the agent again: it works from
the thread, its own replies included. A viewer's note to the agent stays a
message.

To change what you asked, click the text of the note, or of your reply,
and edit it in place: **Save** or Ctrl+Enter puts it, Esc cancels. An
instruction runs again on the edited text, with the thread up to the
agent's answer to it, and that answer changes in place: its old text
stays, dimmed and marked `working`, until the new one lands. The thread
gets no new reply, and the run keeps its ID, so the run link opens the latest
attempt; the attempt before goes with its recording, summary and
artifacts. An edit while the run goes waits for it, then runs again. The
agent's own answers change only that way. A plain message just
changes; an edited note shows `edited` beside its time.

```sh
curl -s -X PUT https://task.example.com/api/notes/$ID \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "@claude update the weather job, and add Lisbon"}'
```

The author of a note, or an admin, edits it and deletes it with its replies, and the author of a reply, or an admin, deletes that reply alone. The form
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

### The internal folder

task-ui builds the folder `internal` itself, with its host commands as
jobs. Only admins see it, at `/f/internal`, and with it the runs of its
jobs, the agent's included. No checkout or project takes the name.

| Job             | Runs                         | What it does                                                                                     |
|-----------------|------------------------------|--------------------------------------------------------------------------------------------------|
| `claude-note`   | `task-ui agent note`         | Claude Code on the thread in `TASK_UI_NOTE`, in the root; its answer goes to `TASK_UI_NOTE_REPLY` |
| `claude-prompt` | `task-ui agent prompt`       | Claude Code on the thread in `TASK_UI_PROMPT`, in `TASK_UI_PROMPT_DIR`, or else the workspace checkout `TASK_UI_WORKSPACE_DIR`; its answer goes to `TASK_UI_PROMPT_REPLY` |
| `claude`        | `task-ui agent shell [args]` | Claude Code in a terminal, in the root, for an admin at the keyboard                             |

Each signs in with `CLAUDE_CODE_OAUTH_TOKEN` from the Claude Code
integration and stops without it. They run `claude` from the `PATH` with
`IS_SANDBOX=1` and `--dangerously-skip-permissions`, so Claude skips the
permission prompts. The board starts `claude-note` as the agent's user,
whose role does not need to see the folder, and `claude-prompt` as the
admin who asks on a folder page; that prompt and its answer are a thread
of the folder's page, never of the board.

---

Previous: [Scheduled](scheduled.md) · Next: [Projects and jobs](projects.md) · [Contents](../README.md)
