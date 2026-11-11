# hours

Input scope: `data_hours`, from `Task\Place::hours()`.

| Field  | Type   | What it is                                                |
|--------|--------|-----------------------------------------------------------|
| `text` | string | The rule as it is written; empty means there is none       |
| `line` | string | The rule read, which is what the cell prints               |
| `open` | bool   | Whether a run is allowed at the moment of the reading      |

The hours a run of a place is allowed in, as the Folder and the Job panel draw
them: the rule, and a badge saying whether it is open now.

A place with no rule is open always and says so. The raw text decides which of
the two the cell draws, and not the line: a rule the parser refuses prints as
itself and is never refused a run, and a rule of nothing but whitespace reads
as an empty line while still being a rule.
