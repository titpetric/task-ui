name: RefinementManager::now
description: >
  The stamp every row carries, to the microsecond. date() in this runtime
  answers with the layout it was handed, which once wrote the literal
  "2006-01-02 15:04:05" into every created_at and left a thread's
  comments with nothing to order by; this goes through DateTime. The
  microseconds are not decoration either: ulid() is not monotonic inside
  a millisecond, so at second resolution two comments of one second have
  no order at all.

  The value moves, so the shape is what a test can hold: the length, the
  separators, and that a later call does not sort before an earlier one.
---
<?php

$now = RefinementManager::now();

echo "length ", strlen($now), "\n";
echo "shape ", preg_match('/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{6}$/', $now), "\n";
echo "century ", substr($now, 0, 2), "\n";
echo "ordered ", (RefinementManager::now() >= $now ? "yes" : "no"), "\n";
---
length 26
shape 1
century 20
ordered yes
