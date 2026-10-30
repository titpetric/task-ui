name: RefinementManager::title
description: >
  A thread's title is the first line of the comment that opened it,
  trimmed, cut at 120 characters with an ellipsis, and "a thread" when
  the comment has no first line to take. The thread list shows nothing
  else, so a comment of ten paragraphs still reads as one line.
---
<?php

echo "[", RefinementManager::title("Design the user tables"), "]\n";
echo "[", RefinementManager::title("  Design the user tables  \n\nand the logins."), "]\n";
echo "[", RefinementManager::title("first\nsecond"), "]\n";
echo "[", RefinementManager::title(""), "]\n";
echo "[", RefinementManager::title("\n\n"), "]\n";

$long = RefinementManager::title(str_repeat("x", 200));

echo "length ", strlen($long), "\n";
echo "ends ", substr($long, -3), "\n";
---
[Design the user tables]
[Design the user tables]
[first]
[a thread]
[a thread]
length 120
ends ...
