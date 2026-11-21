# Asks

A fenced block whose first word is `ask` renders as a form in a note's reply. Its one button runs a job, or answers the agent that asked. The same block in a reply to [a dictation session](voice.md#questions-back) becomes the question card the orb asks on, one field at a time.

The fence is three backticks and the word `ask`, and between them:

```yaml
title: Deploy
fields:
  - key: target
    label: Where
    component: choice
    required: true
    options:
      - value: staging
      - value: prod
        label: Production
submit:
  label: Deploy
  folder: malta/deploy
  job: deploy
  prompt: "deploy to {{target}}"
```

The block is yaml and an unknown key is refused. A typo is an error naming the field it sits in, drawn in place of the form; a setting nobody declared never silently does nothing.

## The form

| Key             | What it is                                                              |
|-----------------|--------------------------------------------------------------------------|
| `id`            | names the form in its artifact path; `ask-<n>` where none is named       |
| `project`       | the project the questions sit in, kept in the record                     |
| `workflow_id`   | the workflow the questions sit in, kept in the record                    |
| `workstep_id`   | the step the questions sit in, kept in the record                        |
| `title`         | the line above the fields                                                |
| `subtitle`      | the line under the title                                                 |
| `context`       | the paragraph that says why the questions are asked                      |
| `prerequisites` | what the reader should hold before answering, one line each              |
| `sections`      | groups the fields under a heading, in the order declared                 |
| `fields`        | one entry per input, drawn in the order declared; a form needs one       |
| `questions`     | the askhuman spelling of `fields`; a form declares one or the other      |
| `submit`        | what the button does                                                     |
| `table`         | the table an admin page's form edits, which a reply's form never names   |

`context_summary` is read as `context`. A section is an `id` and, optionally, a `title` and a `description`; a field names its section with `section`.

## A field

| Key                   | What it is                                                                  |
|-----------------------|------------------------------------------------------------------------------|
| `key`                 | names it in the submission and in `submit.prompt`; `[a-z0-9_-]`, and unique  |
| `label`               | what it is called beside its input, and the question the orb asks           |
| `component`           | which input is drawn, from the list below                                   |
| `kind`                | an askhuman question kind, which names the component where it says none     |
| `axis`                | what the question settles, kept in the decision record                      |
| `section`             | the section the field draws under                                           |
| `help`                | the line under the input                                                    |
| `context`             | the askhuman spelling of `help`                                             |
| `placeholder`         | the empty input's text                                                      |
| `default`             | the value an untouched field submits                                        |
| `required`            | refuses an empty submission                                                 |
| `multi_select`        | the askhuman spelling of a `multi`                                          |
| `free_text`           | asks a reason beside the answer, posted as `<key>.text`                     |
| `free_text_label`     | the reason's label                                                          |
| `max_free_text_bytes` | the reason's cap, 32768 unless named                                        |
| `options`             | the static choices of a `choice`, a `multi` or a `checklist`                |
| `source`              | names the getter that supplies the options instead                          |
| `when`                | `<key>=<value>`: show this field only where that field holds that value     |

`id` is read as `key`, and `title` as `label`. Components: `text`, `textarea`, `number`, `bool`, `hidden`, `choice`, `multi`, `checklist`, and the six that read the instance, `folder`, `job`, `role`, `user`, `key` and `workspace`.

Kinds: `free_text` draws a `textarea`, `qcm_free_text` draws a `choice` with a reason beside it and needs two options, `boolean_checklist` draws a `checklist`.

An option is a `value` (`id` is the askhuman spelling), an optional `label` (`title`), an optional `rationale`, `recommended` and `default_checked` marks, and an optional `next` naming the field it leads to.

`when` is what makes a long form short. A field whose `when` names a field that is itself hidden stays hidden, so a branch closes with the question above it: declare a hundred and the reader answers the dozen that follow from what they said.

## Submit

| Key      | What it is                                                  |
|----------|-------------------------------------------------------------|
| `label`  | the button's words; `Run` where none is named               |
| `folder` | the folder of the job to run                                |
| `dir`    | the directory under that folder to run in                   |
| `job`    | the job's ID                                                |
| `prompt` | the sentence to ask the agent, `{{key}}` replaced by its value |

A placeholder that names no field of the form is an error. Parsing the block reports it, so the form draws the reason and the button is never there to press.

**A form that names a job runs it, and the prompt is then never asked.** The values of the fields it shows are the run's arguments, in the order the fields are declared, empty ones left out: a `hidden` field is how a flag goes in front of the value after it. The foot of the form names the job the button starts, `<folder>/<job>`, because the label is whatever the block's author wrote.

A form with no job answers the note's agent. It is asked the rendered `prompt`, or a sentence naming the answers where the form renders none, followed by the two artifact paths and the instruction to continue the work they decide and not ask the same questions again.

## The sink

`POST /api/forms/submit` takes `{note, block, values, note_text}`, where `block` is the ask's index in the note's body. It needs `prompt/edit` or `note/edit` and the reader's view of the thread.

The sink reads the note from the database and re-parses the fence named by `block`; the values a client posts are never taken on trust. It binds them against the form and refuses a wrong answer with `400` and the field each error sits in: a required field with none, an option no field declares, a reason over `max_free_text_bytes`. A submission that binds is written before anything it asks for runs.

The record is two files under the artifact tree:

| Path                                | What it holds                                                |
|-------------------------------------|---------------------------------------------------------------|
| `ask/<note>/<form>.answers.json`    | one answer per question shown, and who submitted it, and when |
| `ask/<note>/<form>.decisions.md`    | the same answers as a decision record an agent reads          |

`answers.json` follows the askhuman schema: `question_id`, `axis`, `section_id`, `kind`, `selected_options`, `boolean_checks`, `explicit_none` and `free_text`, under the project, workflow, step, form, note, submitter and time. The decision record draws each section and question with the chosen labels and rationales, and closes on `general_note` where the reader wrote one.

## Where it is parsed

[`feature/askhuman/`](https://github.com/titpetric/task-ui/tree/main/feature/askhuman) holds the dialect and nothing else: it reads no database, no template and no file. `POST /api/forms` takes a markdown body and answers with the ask blocks in it, each parsed or refused.

The PHP front end reads the same dialect in `code/Task/Form.php`, so a page drawn without the API still renders the block. The two must agree on every key; a key added to one is added to the other in the same change.
