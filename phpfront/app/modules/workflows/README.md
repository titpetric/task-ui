# Workflows

Generated from the module by `atkins docs`. DO NOT EDIT.

![The tables this module reads](schema.svg)

mod_workflows lists the workflows the reader may see, each with the schedules that start it and how its last run ended, and under them the newest runs of any of them.

A workflow the reader may not see every folder of is not listed at all, so the page and the sidebar hold the same set: Task\Workflow::visible decides it once.

## Pages

| Method | Path                         | Answered by                                      |
|--------|------------------------------|--------------------------------------------------|
| `GET`  | `/workflows/{id}`            | [`mod_workflow.php`](mod_workflow.php)           |
| `GET`  | `/workflows/{id}/edit`       | [`mod_workflow_edit.php`](mod_workflow_edit.php) |
| `GET`  | `/workflows/new`             | [`mod_workflow_new.php`](mod_workflow_new.php)   |
| `GET`  | `/workflows/{id}/runs/{run}` | [`mod_workflow_run.php`](mod_workflow_run.php)   |
| `GET`  | `/workflows`                 | [`mod_workflows.php`](mod_workflows.php)         |

## The data model

### `key_info`

Key Info

| Column               | Type        | Key     |
|----------------------|-------------|---------|
| `name`               | `text`      | primary |
| `purpose`            | `TEXT`      | indexed |
| `algo`               | `text`      |         |
| `bits`               | `integer`   |         |
| `fingerprint`        | `text`      | indexed |
| `scope`              | `text`      | indexed |
| `comment`            | `text`      |         |
| `created_at`         | `timestamp` |         |
| `updated_at`         | `timestamp` |         |
| `backup`             | `text`      |         |
| `backup_fingerprint` | `text`      |         |
| `rotated_at`         | `timestamp` |         |

### `key_role`

Key Role

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `id`         | `text`      | primary |
| `key_name`   | `text`      | indexed |
| `role`       | `TEXT`      | indexed |
| `created_at` | `timestamp` |         |

### `key_user`

Key User

| Column       | Type        | Key     |
|--------------|-------------|---------|
| `id`         | `text`      | primary |
| `key_name`   | `text`      | indexed |
| `user_name`  | `text`      | indexed |
| `created_at` | `timestamp` |         |

### `note`

Note

| Column                  | Type        | Key     |
|-------------------------|-------------|---------|
| `id`                    | `text`      | primary |
| `channel`               | `TEXT`      | indexed |
| `parent_id`             | `text`      | indexed |
| `user_id`               | `text`      | indexed |
| `user_name`             | `text`      |         |
| `body`                  | `text`      |         |
| `mentions`              | `text`      |         |
| `state`                 | `TEXT`      | indexed |
| `run_id`                | `text`      | indexed |
| `created_at`            | `timestamp` |         |
| `updated_at`            | `timestamp` |         |
| `about_run_id`          | `text`      | indexed |
| `about_workflow_run_id` | `text`      | indexed |
| `area`                  | `text`      |         |
| `edited_at`             | `timestamp` |         |
| `deleted_at`            | `timestamp` | indexed |
| `audio_path`            | `text`      |         |
| `spoken`                | `text`      |         |
| `workspace_id`          | `text`      | indexed |
| `dir`                   | `text`      | indexed |

### `run`

Run

| Column        | Type        | Key     |
|---------------|-------------|---------|
| `id`          | `text`      | primary |
| `folder`      | `text`      | indexed |
| `job`         | `text`      | indexed |
| `user_id`     | `text`      | indexed |
| `user_name`   | `text`      |         |
| `state`       | `TEXT`      | indexed |
| `exit_code`   | `integer`   |         |
| `forced`      | `integer`   |         |
| `scheduled`   | `integer`   |         |
| `term_cols`   | `integer`   |         |
| `term_rows`   | `integer`   |         |
| `cast_path`   | `text`      |         |
| `reason`      | `text`      |         |
| `started_at`  | `timestamp` | indexed |
| `finished_at` | `timestamp` |         |
| `created_at`  | `timestamp` |         |
| `updated_at`  | `timestamp` |         |
| `dir`         | `text`      |         |
| `args`        | `text`      |         |
| `target`      | `text`      |         |
| `batch_id`    | `text`      | indexed |
| `message`     | `text`      |         |
| `title`       | `text`      |         |
| `parent_id`   | `text`      | indexed |
| `parent_kind` | `text`      |         |

### `schedule`

Schedule

| Column        | Type        | Key     |
|---------------|-------------|---------|
| `id`          | `text`      | primary |
| `folder`      | `text`      | indexed |
| `job`         | `text`      | indexed |
| `user_id`     | `text`      | indexed |
| `user_name`   | `text`      |         |
| `expression`  | `text`      |         |
| `enabled`     | `integer`   | indexed |
| `last_run_at` | `timestamp` |         |
| `next_run_at` | `timestamp` | indexed |
| `created_at`  | `timestamp` |         |
| `updated_at`  | `timestamp` |         |
| `target`      | `text`      |         |
| `args`        | `text`      |         |
| `triage`      | `integer`   |         |
| `workflow_id` | `text`      | indexed |

### `user`

User

| Column        | Type        | Key     |
|---------------|-------------|---------|
| `id`          | `text`      | primary |
| `name`        | `text`      | indexed |
| `role`        | `TEXT`      |         |
| `folders`     | `text`      |         |
| `created_at`  | `timestamp` |         |
| `updated_at`  | `timestamp` |         |
| `admin`       | `boolean`   |         |
| `mcp`         | `boolean`   |         |
| `agent_reply` | `boolean`   |         |

### `workflow`

Workflow

| Column           | Type        | Key     |
|------------------|-------------|---------|
| `id`             | `text`      | primary |
| `name`           | `text`      | indexed |
| `user_id`        | `text`      | indexed |
| `user_name`      | `text`      |         |
| `created_at`     | `timestamp` |         |
| `updated_at`     | `timestamp` |         |
| `description`    | `text`      |         |
| `prompt`         | `text`      |         |
| `prompt_folder`  | `text`      |         |
| `run_as_user_id` | `text`      | indexed |
| `runs_over`      | `text`      |         |
| `sidebar_hidden` | `boolean`   |         |

### `workflow_run`

Workflow Run

| Column        | Type        | Key     |
|---------------|-------------|---------|
| `id`          | `text`      | primary |
| `workflow_id` | `text`      | indexed |
| `name`        | `text`      |         |
| `user_id`     | `text`      | indexed |
| `user_name`   | `text`      |         |
| `state`       | `TEXT`      | indexed |
| `message`     | `text`      |         |
| `started_at`  | `timestamp` | indexed |
| `finished_at` | `timestamp` |         |
| `created_at`  | `timestamp` |         |
| `updated_at`  | `timestamp` |         |
| `input`       | `text`      |         |
| `answer`      | `text`      |         |

### `workflow_run_step`

Workflow Run Step

| Column             | Type        | Key     |
|--------------------|-------------|---------|
| `id`               | `text`      | primary |
| `workflow_run_id`  | `text`      | indexed |
| `position`         | `integer`   | indexed |
| `folder`           | `text`      |         |
| `job`              | `text`      |         |
| `run_id`           | `text`      | indexed |
| `state`            | `TEXT`      |         |
| `created_at`       | `timestamp` |         |
| `updated_at`       | `timestamp` |         |
| `prompt`           | `text`      |         |
| `answer`           | `text`      |         |
| `workflow_ref`     | `text`      |         |
| `workflow_run_ref` | `text`      |         |
| `kind`             | `TEXT`      |         |
| `markers`          | `text`      |         |

### `workflow_step`

Workflow Step

| Column         | Type      | Key     |
|----------------|-----------|---------|
| `id`           | `text`    | primary |
| `workflow_id`  | `text`    | indexed |
| `position`     | `integer` | indexed |
| `folder`       | `text`    |         |
| `job`          | `text`    |         |
| `disabled`     | `boolean` |         |
| `prompt`       | `text`    |         |
| `dir`          | `text`    |         |
| `workflow_ref` | `text`    |         |
| `target`       | `text`    |         |
| `args`         | `text`    |         |
| `kind`         | `TEXT`    |         |
| `markers`      | `text`    |         |
| `tries`        | `integer` |         |

### `workspace`

Workspace

| Column            | Type        | Key     |
|-------------------|-------------|---------|
| `id`              | `text`      | primary |
| `folder`          | `text`      | indexed |
| `title`           | `text`      |         |
| `description`     | `text`      |         |
| `created_at`      | `timestamp` |         |
| `updated_at`      | `timestamp` |         |
| `home_project_id` | `text`      | indexed |
| `home_url`        | `text`      |         |
| `home_branch`     | `text`      |         |
| `hours`           | `text`      |         |
| `shallow`         | `boolean`   |         |
| `sidebar_limit`   | `integer`   |         |
| `sort`            | `integer`   |         |

## Operations

### `Runs`

Runs reads the workflow_run table and its steps: the runs of every workflow, newest first, one workflow's own a page at a time, and which folders a run touched. The table is this module's, as are the workflow and its steps, which Task\Workflows\Workflows reads.

| Method                                   | What it does                                                                                                                                                                                                                                                                                                                                                                        |
|------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `get($id)`                               | get is one workflow run by ID, or null.                                                                                                                                                                                                                                                                                                                                             |
| `latest($limit = 20)`                    | latest is the runs of every workflow, newest first.                                                                                                                                                                                                                                                                                                                                 |
| `listByWorkflow($id, $filter = array())` | listByWorkflow is one workflow's runs, newest first, a page at a time: limit of them, older than the run `before` names or newer than the one `after` names. A page read after a cursor reads up from it, oldest first, and is turned back over here: the rows either side of a cursor are the newer ones, and LIMIT takes from the top. A cursor on a row that is gone keeps none. |
| `steps($id)`                             | steps is a workflow run's steps, in order.                                                                                                                                                                                                                                                                                                                                          |
| `sees($user, $id, $depth = 0)`           | sees reports whether a user sees every folder a workflow run ran in, through the steps of the workflow runs it started. A user who sees every folder is answered without a query: the rail reads a few hundred runs and the recursion is one statement per run.                                                                                                                     |
| `visible($user, $limit)`                 | visible is the latest runs a user sees, newest first, at most limit of them.                                                                                                                                                                                                                                                                                                        |
| `byUser($viewer, $name, $limit)`         | byUser is the latest workflow runs one user started that another sees, newest first, at most limit of them, which a profile lists. The user is not a column of the query but a test over the window, as the Go page reads it: the window is the latest SCAN runs by anyone, and a user who started none of them has an empty list rather than an older run of theirs.               |
| `rows($found)`                           | rows turns what the database answered into workflow runs.                                                                                                                                                                                                                                                                                                                           |
| `hydrate($row)`                          | hydrate is one row as a workflow run, its timestamps the text the column holds.                                                                                                                                                                                                                                                                                                     |

### `Workflows`

Workflows reads the workflow table and its steps: every workflow by name, one by ID, the steps of one in order, and the workflows a step of one runs. The table is this module's. What a page asks of a workflow - who its runs run as, whether the reader may start it, which step stops them - is Task\Workflow, which asks here for the rows. Nothing here writes. Every form of the workflow pages calls /api/workflows from the browser, so Create, Update and Delete stay in storage/workflow_storage.go.

| Method         | What it does                                                                                                                                                                                                                                                                                                                   |
|----------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `all()`        | all is every workflow, by name.                                                                                                                                                                                                                                                                                                |
| `get($id)`     | get is one workflow by ID, or null.                                                                                                                                                                                                                                                                                            |
| `steps($id)`   | steps is a workflow's step rows, in order.                                                                                                                                                                                                                                                                                     |
| `refs($id)`    | refs is the workflows the steps of one run, in order.                                                                                                                                                                                                                                                                          |
| `stepsIn($id)` | stepsIn is the workflow steps a folder and the folders inside it name, each said as its kind, its folder and the workflow it belongs to: what a move of the project leaves where it was. An ID may hold the underscore LIKE matches any character with, so the match is on the ID itself or on the ID and a slash as a prefix. |
| `allSteps()`   | allSteps is every workflow's steps, in position order, by workflow ID.                                                                                                                                                                                                                                                         |
| `named()`      | named is every workflow by ID: its name, the user its runs run as, and the user row's name where the workflow names one. A workflow with no run as user, or one whose user is gone, runs as whoever starts it, which a schedule names itself.                                                                                  |

---

The tables come from `tables()` on `mod_workflow`, `mod_workflow_edit`, `mod_workflow_new`, `mod_workflow_run`, `mod_workflows`. The operations are the classes of `storage/`, which hold every statement this module runs.
