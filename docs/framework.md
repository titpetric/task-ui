# The front end's shared markup

task-ui serves one stylesheet and one script, `frontend/view/assets/task-ui.css`
and `task-ui.js`, to every page it draws, and a page of a [module](modules.md)
is one of those pages. Between them they carry behaviour a page gets by
writing an attribute rather than by writing JavaScript, which is what makes
a plugin's page behave like the instance's own without shipping a framework.

This chapter is the contract of that markup. A page, a partial, a Go
template or a PHP module writes it and the script does the rest.

## The run status label

A run's state, live, anywhere in the interface:

```html
<span class="state" data-run-id="01K8Z3Q0GJ9S2V7W4X5Y6Z7A8B"></span>
```

That is all of it. The script fills the span with the state the instance
draws everywhere else — the coloured dot, the word, and the exit code of a
run that ended badly — and rewrites it when the run ends. Every label on a
page shares one event stream, so a page may carry fifty.

| Attribute | What it does |
|-----------|--------------|
| `data-run-id` | the run to follow, a ULID. Required, and the span must be a `span`. |
| `data-finished` | `true` for a run that was already over when the page was drawn: never hydrated, never subscribed. |

A span the server leaves empty is filled once over `GET /api/runs/{id}`,
because the stream carries a run when it starts and when it ends and says
nothing in between: a page drawn in the middle of a run would otherwise
show an empty label until the run finished. A span the server already
filled is left alone until the stream has something newer. Either way, the
script sets `data-finished="true"` on a label whose run has ended, and
stops.

**Write `data-finished` when you know.** Most runs are over. A page that
lists a hundred of them and says so for the ninety-nine that ended asks the
server once instead of a hundred times, and subscribes to nothing. A Go
template has the run in hand:

```gotemplate
<span class="state" data-run-id="{{ .ID }}"{{ if ended .State }} data-finished="true"{{ end }}>{{ template "state" . }}</span>
```

A module has whatever its dispatch recorded. `Task::run($note)` answers the
run a dispatch turned into, or `""` for one that has not started yet, so
the refinement module draws its label as:

```php
{if $comment.run}<span class="state" data-run-id="{$comment.run}"></span>{/if}
```

### What it is not

It is not the runs table. `/runs` keys its rows by `tr[data-run-id]` and
updates each row whole — the state, the duration, the actions — from the
same stream. The label claims `span[data-run-id]` and nothing else, so the
two do not meet.

It is not a progress bar. A run reports two things over the stream, that it
started and that it ended, and the label says which. A page that needs the
output as it arrives attaches to the run's terminal instead.

### Where it comes from

`onRuns(fn)` in `task-ui.js` is the one subscription to `/api/sse/runs`;
every handler on the page shares it. `runLabels()` is the label's handler
and the only thing a page has to know about is the markup above.

---

[Contents](README.md) · [Modules](modules.md) · [Development](development.md)
