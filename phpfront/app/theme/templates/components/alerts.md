# alerts

Input scope: `data`.

| Field    | Type   | What it is                                    |
|----------|--------|-----------------------------------------------|
| `notice` | string | What a change that reloaded the page did      |
| `error`  | string | What a form needs, or what failed outside one |

Both arrive in the query, written by `Task\Route::notice()` and
`Task\Route::fail()` before the redirect a form answers with. Either one empty
draws nothing.

What fails inside a form reports in the form, and what fails outside one
reports in the toasts, which the site's own script fills.
