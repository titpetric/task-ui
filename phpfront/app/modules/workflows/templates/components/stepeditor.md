# stepeditor.tpl

The workflow form, new and edit alike: the name, a description, the sidebar toggle, who its runs run as and the workspace whose projects they run over, then the steps in order on a rail, each a card of its kind, and Add step with the kinds to pick from. One blank card a kind follows them in a `<template>`, which task-ui.js clones when a step is added.

Input scope: `data_editor`, as `Task\Editor::data()` builds it. It also sets `data_options` and `data_step` for the components it includes.

| Key            | Type   | What                                                     |
|----------------|--------|----------------------------------------------------------|
| `api`          | string | where the form posts, no mount                           |
| `method`       | string | POST on a new workflow, PUT on edit                      |
| `next`         | string | the path to go to once it saves                          |
| `label`        | string | the submit button                                        |
| `id`           | string | what the fields are named by, `new` or the workflow      |
| `name`         | string | the workflow's name                                      |
| `description`  | string | its description                                          |
| `sidebar`      | bool   | it shows in the sidebar                                  |
| `users`        | map    | who the runs may run as, the scope of user-options.tpl   |
| `runAs`        | string | that user as a reader reads it, for one who may not pick |
| `runAsValue`   | string | that user, "" for whoever starts it                      |
| `isNew`        | bool   | the workflow is being added                              |
| `overNone`     | bool   | no workspace: the steps run once                         |
| `overs`        | list   | the workspaces the steps may run over                    |
| `asks`         | bool   | the reader may add a prompt or a summary                 |
| `shell`        | bool   | the reader may add a script                              |
| `steps`        | list   | the steps, each the scope of wfstep.tpl                  |
| `emptyHidden`  | bool   | the workflow has steps, so No steps yet is hidden        |
| `blankJob`     | map    | the blank command card, the scope of wfstep.tpl          |
| `blankScript`  | map    | the blank script card                                    |
| `blankPrompt`  | map    | the blank prompt card                                    |
| `blankSummary` | map    | the blank summary card                                   |
| `blankCheck`   | map    | the blank check card                                     |

Each item of `overs`:

| Key        | Type   | What                 |
|------------|--------|----------------------|
| `id`       | string | the workspace        |
| `title`    | string | its name             |
| `selected` | bool   | it is the one picked |
