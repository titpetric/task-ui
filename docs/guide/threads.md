# Threads

Every prompt is a thread. A question from a folder's **Prompt** box, a run's page, the orb, a workflow step or `ask_prompt` starts one, and the page at `/threads/<id>` holds the question, what Claude did and its answer.

## Asking

The Threads page is a list and a conversation side by side: the threads the reader may see down the leading side, the selected thread on the other. A row opens `/threads/<id>` without leaving the page, and the address follows, so Back returns to the box that starts a thread.

The box is one component on every page that starts a message: the Threads page, a folder, a run, a workspace and a workflow. It asks Claude by default, with a **Note** switch on the Threads page, names the place it asks in and sets the agent a reply is posted as. The agent is a user with **Agent reply** and **MCP** set; it names who answers, and the thread keeps it for every later question. The run still works as the user who asked, with that user's permissions and token, so picking an agent never grants the agent's rights.

Every form marked `data-send` posts one message. An ask posts `/api/prompts` with `folder`, `dir`, `prompt`, `agent` and `origin`; a note posts `/api/notes`; a chat posts `/api/threads`. The page goes to the note the API answers, which is the thread's page. A prompt box links to the threads of its place with **Threads in this place**, and lists none inline.

## The Claude run

`@claude-prompt` writes what Claude does to the file `TASK_UI_TRANSCRIPT` names, one JSON object a line: `status`, `text`, `thinking`, `tool_use`, `tool_result`, `system`, `result` and `error`. The board reads the file as the run goes and stores each record, so the thread page shows the work without waiting for the end: a pinned line with what Claude is doing now and how long it has run, a card per tool call with its input and its result, and thinking behind a toggle. Claude's bubble is made when the run starts and its work streams into it. The run's final message fills it, drawn on the agent's side with an **agent** chip, even when the agent answers as the asker's own user.

The status line is the last state of the run, and it has **Stop** and **Kill** for a reader who may end it. A working thread's row in the list shows **Stop** too, so a run is ended without opening its thread. Stop sends SIGHUP and SIGTERM to the run's process group; Kill sends the same and then SIGKILL after the grace period. The process group is the whole tree: a script, the job it starts and the commands under that. The jobs table and a run's own page have the same two actions.

The page follows the thread over Server-Sent Events at `/api/sse/thread/<id>`. Each work event writes an `id:`, the page opens the stream at the newest event it rendered (`data-last-event`), and the handler replays what the store holds after that cursor before subscribing. A short sweep of the store covers a bus that dropped events. A reload mid-run renders the stored transcript, and a reconnect picks up where the page left off.

## The list

`/threads` lists every thread the reader may see, newest activity first: board notes with `note/view`, prompts with `prompt/view` in a place the reader sees, and chats they are in. The search box narrows the list by words of the body, and **+ Filters** adds a place, a state and a user. Press `/` to reach the search box. A selected row is marked in the list, and **All threads** on a thread's own page returns to it. Recents and the dashboard show a working thread at the top with a pulsing marker, and they do not also list its Claude run.

`GET /threads?id=<id>` answers a redirect to `/threads/<id>`, so an old link and a stale history entry both land on the conversation.

## API

| Request | What it answers |
|---------|-----------------|
| `GET /api/threads?threads=1&workspace=&folder=&user=&state=&q=&live=1&limit=` | `{unread, chats, threads}`, the conversations the reader may see; `live=1` keeps open and working threads and the ones finished in the last hour |
| `GET /api/threads/{id}` | a prompt or a note as `{thread, agent, agents, messages, runs}`, `runs` grouping the events under the question that started each one; a chat keeps its shape |
| `POST /api/threads/{id}/agent` | `{agent}`: changes the agent of a thread, by its author or an admin, and refused while it works |

MCP serves `list_threads` and `read_thread` over the same API (see [MCP](mcp.md)).
