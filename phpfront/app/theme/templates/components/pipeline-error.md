# pipeline-error

Input scope: `data_finding`, one row of `Task\Findings::alert()`.

| Field    | Type   | What it is                                              |
|----------|--------|---------------------------------------------------------|
| `folder` | string | The folder whose pipeline the linter refused            |
| `href`   | string | Its page                                                |
| `path`   | string | The file the finding is about                           |
| `job`    | string | The job it is about; empty where it is about the file    |
| `issue`  | string | The rule that fired, which titles the detail; empty draws no title |
| `detail` | string | What the linter said                                    |

One finding of the pipeline alert. The alert draws it twice over: once for the
findings it shows, and once for the ones it folds under Show more.
