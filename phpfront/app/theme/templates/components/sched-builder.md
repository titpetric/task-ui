# sched-builder

Input scope: none. The component is static markup and `task-ui.js` is what
makes it work, so a page includes it and hands it nothing.

The fields of a schedule, from Repeat down to the button that saves it: the
controls that compile a cron expression, the expression itself, and the table
that says what each of its five positions means.

The expression is the only field of the request. Every control above it carries
no `name`, so `serialize()` passes over it: a control here is a way of writing
the expression and not a thing the API is told. `#sch-expr` is the one field
with a name and the API's own `expression`.

| Attribute           | On                    | What it is                                  |
|---------------------|-----------------------|---------------------------------------------|
| `data-sched`        | the `<form>`          | What the script scopes its queries to, written by the page and not by this |
| `data-sched-mode`   | the Repeat select     | hourly, daily, weekly, monthly or expression |
| `data-sched-when`   | a field               | The modes the field shows under, space separated |
| `data-sched-minute` | a number              | The minute past the hour, for hourly        |
| `data-sched-time`   | a time                | The hour and the minute, for the other three |
| `data-sched-day`    | a checkbox            | A day of the week, in cron's numbering: Sunday is 0 |
| `data-sched-dom`    | a number              | The day of the month, 1 to 28               |
| `data-sched-days`   | a button              | The days it checks, as a list; empty clears them |
| `data-sched-pick`   | a button              | The day of the month it sets                |
| `data-sched-expr`   | the Expression field  | What the controls compile into and what reads back |
| `data-sched-says`   | a line                | The expression in words                     |
| `data-sched-pos`    | a cell of the table   | The field of the expression it shows, 0 to 4 |
| `data-sched-cancel` | a button              | Leaves an edit and puts the form back to adding |
| `data-sched-label`  | the submit's text     | Add schedule, or Save schedule while editing |
| `data-sched-icon`   | the submit's icon     | Hidden while editing: a Save carries none   |

The page writes `data-sched` on its `<form>`, because the form is what holds
the other fields an add needs, the folder and the job of one page and the
hidden workflow of another, and editing turns those off. A page that includes
this and forgets the attribute draws the fields and drives nothing.

The day of the month stops at 28 and not 31. A schedule of the 31st does not
run in February, and the builder offers no day that skips a month; an
expression that names one is still read and still shown, in Expression mode.

The table is the legend and not a control: nothing in it is typed into. The
`This one` column follows the field as it is written, so an expression read
back from a schedule is explained position by position without being parsed by
the reader.
