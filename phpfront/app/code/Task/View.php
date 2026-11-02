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

	var $module;
	var $chrome;

	function __construct($module, $chrome = null) {
		$this->module = $module;
		$this->chrome = $chrome === null ? new Chrome : $chrome;
	}

	/** templates is the module's template directory. */
	function templates() {
		return "modules/" . $this->module->name() . "/templates/";
	}

	/** body renders the module's own template over its scope. */
	function body($template, $data) {
		$tpl = new \MiniTPL\Template(array($this->templates(), self::THEME));

		$tpl->load($template);
		$tpl->assign("base", Host::base());
		$tpl->assign("chrome", $this->chrome->data());
		$tpl->assign("data", $data);

		return $tpl->get();
	}

	/** page renders the module's body inside the theme's shell and echoes it. */
	function page($data, $template = "index.tpl") {
		$content = $this->body($template, $data);
		$chrome = $this->chrome->data();

		$shell = new \MiniTPL\Template(array(self::THEME));

		$shell->load("index.tpl");
		$shell->assign("base", $chrome["base"]);
		$shell->assign("chrome", $chrome);
		$shell->assign("content", $content);

		echo $shell->get();
	}
}
