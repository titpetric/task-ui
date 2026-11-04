<?php

namespace Task;

// View renders a module's body into the theme's shell.
//
// The search paths are the module's templates and the theme's, in that order,
// so a module's own index.tpl wins and {include components/table.tpl} resolves
// into the theme. No compile location is set: MiniTPL's default writes beside
// each search path, which is the templates_c of the module and of the theme.
class View {
	/** THEME is the theme's template directory. */
	const THEME = "theme/templates/";

	/** MISSING is what the page of a thing that is not there says. */
	const MISSING = "Someone may have renamed it in the pipeline file, or you don't have access to it. Pick a folder on the left.";

	/** REFUSED is what the page a reader may not see says, which the alerts draw. */
	const REFUSED = "You need permission to open this page.";

	/** GRANTED is the rest of the refusal's lead, after what the alert said. */
	const GRANTED = " An admin grants permissions under Settings. Pick a folder on the left to carry on.";

	var $module;
	var $chrome;

	function __construct($module, $chrome = null) {
		if ($chrome === null) {
			$chrome = new Chrome;
		}

		$this->module = $module;
		$this->chrome = $chrome;
	}

	/** templates is the module's template directory. */
	function templates() {
		return "modules/" . $this->module->name() . "/templates/";
	}

	/**
	 * body renders the module's own template over its scope.
	 *
	 * The chrome is given rather than read: page reads it once and hands it to
	 * both templates. Reading it twice costs the recents rail a second walk of
	 * every workspace, which a trace of /runs put at 110ms of a 2s page.
	 */
	function body($template, $data, $chrome) {
		$paths = array($this->templates(), self::THEME);
		$tpl = new \MiniTPL\Template($paths);

		$tpl->load($template);
		$tpl->assign("base", Host::base());
		$tpl->assign("chrome", $chrome);
		$tpl->assign("data", $data);

		return $tpl->get();
	}

	/**
	 * missing answers for a page that is not there, or one the reader may not
	 * see: the status, the heading, what is missing and the theme's own body
	 * for it.
	 *
	 * A page that is refused answers as one that is not there, because a
	 * reader who may not see a run may not learn that the run exists either.
	 */
	function missing($why) {
		$head = new Head("Not found", self::MISSING);

		Route::missing();
		$this->chrome->heading("Not found", array(array("text" => "Not found", "href" => "")));
		$this->chrome->reports($why);
		$this->page(array("head" => $head->data()), "notfound.tpl");
	}

	/**
	 * refused answers for a page the reader may not open: the status, the
	 * heading and the theme's own body for it.
	 *
	 * A refusal says what the reader needs and no more, because what a page
	 * holds is not a refused reader's to learn. A page that is there but hides
	 * a row answers as missing() instead.
	 */
	function refused() {
		$head = new Head("Not allowed", self::REFUSED . self::GRANTED);

		Route::refused();
		$this->chrome->heading("Not allowed");
		$this->chrome->reports(self::REFUSED);
		$this->page(array("head" => $head->data()), "forbidden.tpl");
	}

	/**
	 * assets is the module's own stylesheet and script, where it has them:
	 * modules/<name>/assets/<name>.css and the same name with .js, which the
	 * asset router serves at /assets/<name>/.
	 *
	 * Most modules have neither. A page drawn out of the theme's own components
	 * is styled by the theme's own stylesheet, and the one script every page
	 * shares holds the form contract; a module ships a file of its own where it
	 * draws markup nothing else draws.
	 *
	 * Asked of the filesystem rather than declared, because declaring it is a
	 * method on the Module interface and an interface here inherits nothing:
	 * every one of the forty mod_ files would write out a list it has no use
	 * for. Two stats of the module that is serving, and none of the rest.
	 */
	function assets($chrome) {
		$name = $this->module->name();
		$dir = "modules/" . $name . "/assets/" . $name;
		$href = $chrome["assets"] . "/" . $name . "/" . $name;
		$version = "?v=" . $chrome["assetVersion"];
		$styles = array();
		$scripts = array();

		if (file_exists($dir . ".css")) {
			$styles[] = $href . ".css" . $version;
		}

		if (file_exists($dir . ".js")) {
			$scripts[] = $href . ".js" . $version;
		}

		return array("styles" => $styles, "scripts" => $scripts);
	}

	/** page renders the module's body inside the theme's shell and echoes it. */
	function page($data, $template = "index.tpl") {
		$chrome = $this->chrome->data();
		$content = $this->body($template, $data, $chrome);
		$assets = $this->assets($chrome);

		$chrome["styles"] = $assets["styles"];
		$chrome["scripts"] = $assets["scripts"];

		$paths = array(self::THEME);
		$shell = new \MiniTPL\Template($paths);

		$shell->load("index.tpl");
		$shell->assign("base", $chrome["base"]);
		$shell->assign("chrome", $chrome);
		$shell->assign("content", $content);

		echo $shell->get();
	}
}
