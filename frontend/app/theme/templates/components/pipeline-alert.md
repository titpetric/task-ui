# pipeline-alert

Input scope: `data_findings`, from `Task\Findings::alert()`.

| Field       | Type   | What it is                                           |
|-------------|--------|------------------------------------------------------|
| `pipelines` | int    | How many pipeline files the linter refused           |
| `says`      | string | "pipeline has" or "pipelines have", for that count   |
| `shown`     | list   | The findings the alert lists, five at most           |
| `more`      | list   | The rest, folded under Show more                     |
| `rest`      | int    | How many that is; zero draws no fold                 |

The alert of the broken atkins pipelines in a place: a row per finding with its
folder, file and job, the rows past five folded away. atkins lists no job of a
pipeline its linter refuses, so a folder with a broken file shows no jobs and
this says why.

A place with no finding draws nothing at all, which is the null
`Task\Findings::alert()` answers: the page tests it and includes nothing. Each
row is `components/pipeline-error.tpl`, drawn in both lists.
