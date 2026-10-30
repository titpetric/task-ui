name: RefinementPage::flatHeadings
description: >
  A request is drawn as markdown, and its headings go down to the
  smallest one, which the page sets as a bold line. A request is a
  paragraph inside a thread and not a document: a flattened one carries
  "## The design" and "## The rules", and in the page's own h2 those read
  as though the thread had been cut into chapters.

  Inside a fence nothing is a heading. A shell snippet's comment is a
  comment, and rewriting it would rewrite the code the request is
  showing.
---
<?php

echo RefinementPage::flatHeadings(implode("\n", array(
	"# The design",
	"",
	"A schema.",
	"",
	"## The rules ##",
	"",
	"1. Use ULID.",
	"",
	"```sh",
	"# not a heading, a comment",
	"mig lint",
	"```",
	"",
	"###### already the smallest",
	"",
	"not # a heading",
))), "\n";
---
###### The design

A schema.

###### The rules

1. Use ULID.

```sh
# not a heading, a comment
mig lint
```

###### already the smallest

not # a heading
