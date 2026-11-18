# Threads

Every prompt is a thread. A question from a folder's **Prompt** box, a run's page, the orb, a workflow step or `ask_prompt` starts one, and the page at `/threads/<id>` holds the question, what Claude did and its answer.

## Asking

A prompt box carries the place it asks in and a picker of the agents a reply may be posted as. The agent is a user with **Agent reply** and **MCP** set; it names who answers, and the thread keeps it for every later question. The run still works as the user who asked, with that user's permissions and token, so picking an agent never grants the agent's rights.

A new question lands on the thread page. The box posts to `/api/prompts` with `folder`, `dir`, `prompt`, `agent` and `origin`, and the thread opens on the answer.

## The Claude run

`@claude-prompt` writes what Claude does to the file `TASK_UI_TRANSCRIPT` names, one JSON object a line: `status`, `text`, `thinking`, `tool_use`, `tool_result`, `system`, `result` and `error`. The board reads the file as the run goes and stores each record, so the thread page shows the work without waiting for the end: the status line, the text Claude says, a card per tool call with its input and its result, and thinking behind a toggle. The final message is the answer.

The page follows the thread over Server-Sent Events at `/api/sse/thread/<id>` and appends each record as it arrives. A reload mid-run renders the stored transcript. A run that ends without a final message gets an error event, and a thread whose run was lost in a restart still reads the transcript it wrote.

## The list

`/threads` lists every thread the reader may see, newest activity first: board notes with `note/view`, prompts with `prompt/view` in a place the reader sees, and chats they are in. Recents and the dashboard show a working thread at the top with a pulsing marker, and they do not also list its Claude run.

## API

| Request | What it answers |
|---------|-----------------|
| `GET /api/threads?threads=1&workspace=&folder=&user=&state=&q=&live=1&limit=` | `{unread, chats, threads}`, the conversations the reader may see; `live=1` keeps open and working threads and the ones finished in the last hour |
| `GET /api/threads/{id}` | a prompt or a note as `{thread, agent, agents, messages, runs}`, `runs` grouping the events under the question that started each one; a chat keeps its shape |
| `POST /api/threads/{id}/agent` | `{agent}`: changes the agent of a thread, by its author or an admin, and refused while it works |

MCP offers `list_threads` and `read_thread` over the same API (see [MCP](mcp.md)).
