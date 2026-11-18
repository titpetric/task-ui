# Search

The header field and the Search page ask an agent over the documentation and job targets of every project you can see.

![Search](https://cdn.si/titpetric/task-ui/docs/assets/search-dark.png)

The field follows the name at the top of every page. Type a question and press Enter: task-ui starts a Claude run that searches the index and lands on the [thread](notes.md) that answers it, where the working state and the cited answer follow as they do on any thread. The Search page, in the sidebar, is the same box with a line on what it does. Both take `prompt/edit`, `prompt/view` and `note/view`, and a Claude model configured on the Instance page.

## What is indexed

Every start of the service walks the projects of every workspace and reindexes what changed, in the background, while the instance serves. A file whose content hash is the same as the last index is skipped; a changed file replaces its sections; a file that left the tree leaves the index.

- Every `.md`, `.markdown` and `.txt` file, up to 1 MiB, in every project of every workspace. Dot directories, `node_modules`, `vendor`, `ext`, `templates_c`, `dist`, `build`, `target` and `__pycache__` are not read.
- Every job target a project declares, with the description, summary, steps, aliases and command the source gives it: `atkins.yml`, `Makefile`, `Taskfile`, `package.json`, `composer.json` and `scripts/*`. A target is listed by its folder and its job ID, such as `internal` and `ci:daily`.

Markdown is read with the blog's renderer, [blackfriday](https://github.com/russross/blackfriday), and walked as a document: each heading starts a section that carries its ancestors, so `docs/guide/scheduled.md > Scheduling a job` is a place in the index and not a line in a blob. Front matter is skipped, and a fenced block keeps its text and hides its hashes. A text file is one section named by the file.

The index holds no paragraphs in an agent's context until a search reaches one. A question narrows to sections first; the agent reads a document's headings next; and only the section it needs is read whole.

## The agent

The run gets the `task` MCP server as every prompt run does, with four tools:

| Tool             | Arguments        | What it reads                                                                                                    |
|------------------|------------------|------------------------------------------------------------------------------------------------------------------|
| `search_docs`    | `query`, `limit` | The sections that hold the words, best first, with a short excerpt, the path, the heading and the section number |
| `search_outline` | `id`             | A document's headings in order, with the section numbers, and no paragraph text                                  |
| `search_read`    | `id`, `section`  | One section of a document, whole                                                                                 |
| `search_browse`  | `path`           | The folders, documents and job targets directly under a path, which is how the folder structure is read          |

The instructions the Search page dispatches are [`prompts/search.md`](https://github.com/titpetric/task-ui/blob/main/phpfront/app/prompts/search.md): answer from the tools, cite the path and heading, quote under three lines, and say plainly when the index holds nothing.

The tools take `prompt/view`, and every result is filtered to the folders the user sees. `GET /api/search`, `/api/search/browse`, `/api/search/documents/{id}` and `/api/search/documents/{id}/sections/{ordinal}` answer the same, and `POST /api/search/reindex` starts an index run for an admin (see [API](../api.md)).

## Daily CI from source

A job target named `ci:daily` is a project saying how its daily CI runs. When an index ends, task-ui registers one schedule for each such target, `@daily`, disabled and with no user: nothing runs until an admin opens [Scheduled](scheduled.md), picks **Runs as** and enables it. The row is marked `ci:daily`.

The reconciler never replaces an existing schedule for the same folder and job, whatever its expression. An admin's change makes the row an ordinary schedule (`source` is cleared), and the reconciler leaves it alone. A `ci:daily` target that leaves the source disables the row it registered. Jobs in a directory inside a folder are not scheduled: a schedule names a folder and a job.

The run still takes the user's permissions and the folder's allow list, and the job's hours. A target is approved once, as the source of the CI it declares.

---

Previous: [Modules](modules.md) · Next: [Forms](forms.md) · [Contents](../README.md)
