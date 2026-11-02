<?php

namespace MiniTPL;

interface Hook {
	const POSITION_PRE = "pre";
	const POSITION_POST = "post";

	function execute($filename, $contents);
}
