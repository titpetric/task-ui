name: thread-title
description: Task\Threads\Conversation::title heads a thread with its opening's first line.
---
<?php

require "ext/autoload.php";

$heading = new Task\Model\Note();
$heading->setId("fx-title-n1");
$heading->setUserName("fx-title-ann");
$heading->setBody("# Placement check\n\na second paragraph that is not the title");

$plain = new Task\Model\Note();
$plain->setId("fx-title-n2");
$plain->setUserName("fx-title-ann");
$plain->setBody("why did the build fail");

echo Task\Threads\Conversation::title($heading, []), "\n";
echo Task\Threads\Conversation::title($plain, []), "\n";
---
Placement check
why did the build fail
