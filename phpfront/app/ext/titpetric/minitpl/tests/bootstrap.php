<?php

declare(strict_types=1);

error_reporting(E_ALL);

// Fixture paths are relative to the project root, so the suite behaves the
// same no matter where phpunit was invoked from.
chdir(dirname(__DIR__));

require_once 'vendor/autoload.php';

mkdir('tests/compile');
