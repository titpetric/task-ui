name: RefinementPage::split
description: >
  What is looked at against what is read. A line that is nothing but an
  embedded image is the output and goes above everything; the rest is
  the run's prose and folds behind "show thinking". An image inside a
  sentence stays in the sentence, and an image inside a fence is a
  fragment being shown rather than a picture being embedded.
---
<?php

$report = implode("\n", array(
	"# Credentials",
	"",
	"![The schema](diagrams/schema.svg)",
	"",
	"Three tables, kept apart.",
	"",
	"  ![The preview](diagrams/erd.png)  ",
	"",
	"See ![inline](x.png) in a sentence.",
	"",
	"```md",
	"![not embedded](example.png)",
	"```",
));

$out = RefinementPage::split($report);

echo "--- images ---\n", $out["images"], "\n";
echo "--- prose ---\n", $out["prose"], "\n";

$plain = RefinementPage::split("# Just words\n\nNothing to look at.\n");

echo "--- no images ---\n[", $plain["images"], "]\n", $plain["prose"], "\n";
---
--- images ---
![The schema](diagrams/schema.svg)

![The preview](diagrams/erd.png)
--- prose ---
# Credentials

Three tables, kept apart.

See ![inline](x.png) in a sentence.

```md
![not embedded](example.png)
```
--- no images ---
[]
# Just words

Nothing to look at.
