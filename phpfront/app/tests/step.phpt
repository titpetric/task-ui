name: step
description: A check step looks for its markers in the step before it and sends the workflow's steps round again while it finds one, and the card and the chain link draw it as that.
---
<?php

// A blank card of the kind the Add step menu offers, which the template clones:
// a check starts with the marker a reader most often wants and the default
// passes.
$card = Task\Step::added(Task\Step::CHECK);

echo $card["kind"], " ", $card["markers"], " ", $card["tries"], "\n";
echo Task\Step::kindLabel($card), " ", Task\Step::icon($card), "\n";
echo Task\Step::line($card), "\n";

// A check with nothing to look for has no line, so the folded card falls back
// to naming what it lacks instead of a half sentence.
$bare = $card;
$bare["markers"] = "";

echo "[", Task\Step::line($bare), "]\n";

// Several markers are drawn as the reader wrote them, comma separated.
$many = $card;
$many["markers"] = "Next:,TODO:";

echo Task\Step::line($many), "\n";

// passes holds the field between what a check needs to send anything round and
// what keeps a step writing the marker from looping on: a step naming none
// takes the default.
echo Task\Step::passes(0), " ", Task\Step::passes(1), " ", Task\Step::passes(5), " ", Task\Step::passes(99), "\n";

// A check asks Claude nothing: it reads the step before it, so the thread of a
// run neither names it an agent nor takes its text for the question.
echo var_export(Task\Step::asks($card), true), "\n";

// The chain of a workflow's row draws the check as a link of its own, and the
// icon branch of the template turns on the flag.
$link = Task\Chain::blank();
$link["check"] = true;
$link["text"] = "check for Next:";

$node = Task\Pipeline::node($link, 2);

echo var_export($node["check"], true), " ", $node["text"], " ", $node["small"], "\n";
---
check Next: 3
Check rotate
the steps go round again while the step before it writes Next:
[]
the steps go round again while the step before it writes Next:,TODO:
3 2 5 20
false
true check for Next: step 3
