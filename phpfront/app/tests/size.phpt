name: size
description: Task\Size words a quantity of bytes the way every page words one - a file's size, a container's memory in megabytes, and a second of traffic.
---
<?php

require "ext/autoload.php";

// Whole bytes under a kilobyte, and a tenth of the larger unit above it, each
// unit from the one under it, so a size reads as the Go pages print it.
echo Task\Size::bytes(0), " ", Task\Size::bytes(1023), " ", Task\Size::bytes(1024), "\n";
echo Task\Size::bytes(1536), " ", Task\Size::bytes(1048576), " ", Task\Size::bytes(1073741824), "\n";
echo Task\Size::bytes(1099511627776), " ", Task\Size::bytes(1125899906842624), "\n";

// A column holds a size as text and a host answers one as a float: both read
// as the number they hold, and nothing reads as no bytes at all.
echo Task\Size::bytes("2048"), " ", Task\Size::bytes(1536.9), " ", Task\Size::bytes(""), "\n";

// Beszel reports a container's memory in megabytes, and the traffic in bytes
// a second.
echo Task\Size::megabytes(0), " ", Task\Size::megabytes(1), " ", Task\Size::megabytes(512), "\n";
echo Task\Size::megabytes(1536), " ", Task\Size::megabytes(0.5), "\n";
echo Task\Size::rate(0), " ", Task\Size::rate(2048), " ", Task\Size::rate(1572864), "\n";
---
0 B 1023 B 1.0 KB
1.5 KB 1.0 MB 1.0 GB
1.0 TB 1024.0 TB
2.0 KB 1.5 KB 0 B
0 B 1.0 MB 512.0 MB
1.5 GB 512.0 KB
0 B/s 2.0 KB/s 1.5 MB/s
