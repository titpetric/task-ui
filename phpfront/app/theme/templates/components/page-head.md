# page-head

Input scope: `data_head`, from `Task\Head::data()`.

| Field     | Type   | What it is                                           |
|-----------|--------|------------------------------------------------------|
| `title`   | string | The heading, which repeats the sidebar row            |
| `icon`    | string | The glyph before the heading, named without the `i-` prefix; empty draws none |
| `badge`   | string | The label beside it: a run's state, a report; empty draws none |
| `tone`    | string | The badge's colour, from `Task\Run::tone()`; empty draws the plain badge |
| `tags`    | list   | The plain badges after it, in order: a job's kind, then that it is interactive |
| `note`    | string | The faint text after the badge: a terminal's title; empty draws none |
| `lead`    | string | The paragraph under it, as HTML; empty draws none     |
| `leads`   | bool   | Whether the paragraph is drawn even when it says nothing |
| `faint`   | string | A second paragraph under the lead, in the faint colour, escaped; empty draws none |
| `live`    | bool   | Whether the actions div carries `data-actions`, which the script redraws |
| `run`     | object | The Run button of a job's page, read by `run-buttons`; empty draws none |
| `actions` | object | The head's own `[...]`, from `Task\Menu::data()`; empty draws none |
| `buttons` | object | The row of actions on the right, from `Task\Buttons::data()`; empty draws none |
| `show`    | bool   | Whether anything sits on the right at all, which is the one div that holds it |

The Run button, the buttons and the `[...]` sit in one `<div class="actions">`
in that order, and a head with none of them draws no div: `Task\Head::data()`
answers `show`, so no page decides that for itself. A page whose markup opens
the div before it knows what is in it says so with `Task\Head::always()`,
which is a folder's page and a job's.

`run` is apart from `buttons` because it is a split button: the Run of a job
that Claude summarizes, or one in a workspace with hours, carries a caret with
the other ways to run it, which no row of plain buttons can hold.

`lead` is drawn raw, so a sentence built out of a run's own values goes
through `Task\Head::said()`, which escapes it.

`leads` is for a page whose markup writes the paragraph without asking whether
there is anything in it, which is a job's: it draws the job's description, and
a job with none draws the empty paragraph. Every other page leaves the
paragraph out where it says nothing, so `Task\Head::leads()` is asked for.

`faint` is text and stays text: a page whose second line carries markup of its
own — the run page's state badge, its stamp in a `title` and its duration —
draws that head itself rather than hand PHP a string of HTML to pass through.
PHP builds scopes and a template draws, so a field that took markup would move
the markup into the module.

A reader who may not press a button is told why in the button's place, as an
item of the row with `faint` set rather than a line beside it: the workflow
page says so where Run would stand and keeps Edit to the right of it.
