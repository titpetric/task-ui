# Forms

A fenced block whose first word is `form` renders as a form in a note's reply. Its one button runs a job, or asks Claude a sentence built from the values. The same block in a reply to [a dictation session](voice.md#questions-back) becomes the question card the orb asks on, one field at a time.

The fence is three backticks and the word `form`, and between them:

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

| Key      | What it is                                                             |
|----------|------------------------------------------------------------------------|
| `title`  | the line above the fields                                              |
| `fields` | one entry per input, drawn in the order declared; a form needs one     |
| `submit` | what the button does                                                   |
| `table`  | the table an admin page's form edits, which a reply's form never names |

## A field

| Key           | What it is                                                                  |
|---------------|-----------------------------------------------------------------------------|
| `key`         | names it in the submission and in `submit.prompt`; `[a-z0-9_-]`, and unique |
| `label`       | what it is called beside its input, and the question the orb asks           |
| `component`   | which input is drawn, from the list below                                   |
| `help`        | the line under the input                                                    |
| `placeholder` | the empty input's text                                                      |
| `default`     | the value an untouched field submits                                        |
| `required`    | refuses an empty submission                                                 |
| `options`     | the static choices of a `choice` or a `multi`                               |
| `source`      | names the getter that supplies the options instead                          |
| `when`        | `<key>=<value>`: show this field only where that field holds that value     |

Components: `text`, `textarea`, `number`, `bool`, `hidden`, `choice`, `multi`, and the six that read the instance, `folder`, `job`, `role`, `user`, `key` and `workspace`.

An option is a `value`, an optional `label` shown in its place, and an optional `next` naming the field it leads to.

`when` is what makes a long form short. A field whose `when` names a field that is itself hidden stays hidden, so a branch closes with the question above it: declare a hundred and the reader answers the dozen that follow from what they said.

## Submit

| Key      | What it is                                                  |
|----------|-------------------------------------------------------------|
| `label`  | the button's words; `Run` where none is named               |
| `folder` | the folder of the job to run                                |
| `dir`    | the directory under that folder to run in                   |
| `job`    | the job's ID                                                |
| `prompt` | the sentence to ask Claude, `{{key}}` replaced by its value |

A placeholder that names no field of the form is an error. Parsing the block reports it, so the form draws the reason and the button is never there to press.

## Where it is parsed

[`form/`](https://github.com/titpetric/task-ui/tree/main/form) holds the dialect and nothing else: it reads no database, no template and no file. `POST /api/forms` takes a markdown body and answers with the blocks in it, each parsed or refused. A client draws a form off that answer and holds no yaml parser of its own.
