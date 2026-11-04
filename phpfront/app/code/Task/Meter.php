<?php

namespace Task;

// Meter is a share of something, 0 to 100, as the bar that draws it reads it.
//
// The share decides the tone and the template compares it, so it goes along
// with the two roundings: a tenth of a percent is the width of the bar, which
// is a style, and a whole percent is the number beside it, which is what a
// reader wants. The Go define does both with printf; MiniTPL has no printf, so
// they are done here.
class Meter {
	/** of is one share as the meter component reads it. */
	static function of($share) {
		return array(
			"share" => $share,
			"width" => sprintf("%.1f", $share),
			"pct" => sprintf("%.0f", $share),
		);
	}
}
