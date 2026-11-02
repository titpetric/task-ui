<?php

// Module is what every mod_ class declares. An interface carries no body and
// no storage here, so each module writes all of them.
interface Module {
	/** name is the slug: a module's assets answer under /assets/<name>/. */
	function name();

	/** title is the heading and the sidebar row. */
	function title();

	/** icon names a symbol of the sprite, without the i- prefix. */
	function icon();

	/** tables are the tables this module reads, for the generated docs. */
	function tables();

	/** serve answers the matched request. */
	function serve();
}
