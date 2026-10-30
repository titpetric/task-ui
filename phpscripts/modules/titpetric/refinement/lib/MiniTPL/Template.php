<?php

namespace MiniTPL;

/*

Tit Petric <black@titpetric.com>, MIT license, see LICENSE.

 */

/** Template class */
class Template {
	const E_TEMPLATE_COMPILE = "Template file '%s' doesn't exist! Is the compile dir writable?";
	const E_FILENAME_EMPTY = "Filename can't be empty, tried to render '%s'";

	/** Holds search paths */
	var $_paths;
	/** Disable cache */
	var $_nocache = false;
	/** Auto-escape printed variables */
	var $_escape = true;
	/** Compile location, relative or absolute */
	var $_compile_location;
	var $_compile_absolute;
	/** Defaults */
	var $_defaults;

	/** Hold assigned values, filenames, stack */
	protected $stack = array();
	protected $filename;
	protected $source;
	protected $vars = array();
	protected $hooks = array(Hook::POSITION_PRE => array(), Hook::POSITION_POST => array());

	/** Default constructor */
	function __construct($paths = false) {
		$this->set_compile_location("cache/", false);
		$this->set_paths($paths);
		$this->_defaults = array(array("ldelim", "{"), array("rdelim", "}"));
		$this->_nocache = false;
	}

	function add_default($k, $v = '') {
		$this->_defaults[] = array($k, $v);
	}

	function _default_vars() {
		if (empty($this->stack)) {
			$this->vars = array();
		}

		$this->filename = false;
		foreach ($this->_defaults as $v) {
			$this->assign($v[0], $v[1]);
		}
	}

	function add_hook(Hook $hook, string $position) {
		$this->hooks[$position][] = $hook;
	}

	function clear_hooks() {
		$this->hooks = array(Hook::POSITION_PRE => array(), Hook::POSITION_POST => array());
	}

	function push() {
		$this->stack[] = $this->filename;
	}

	function pop() {
		list($this->filename) = array_splice($this->stack, -1);
	}

	/** Template loader */
	function load($filename) {
		$r = 0;

		$this->_default_vars();
		$path = $this->_find_path($filename);
		if ($path !== false) {
			$f_original = $path . $filename;
			$f_compiled = $this->_compile_path($path) . $filename;
			if (file_exists($f_compiled)) {
				$r = 1;
				if (file_exists($f_original) && (filemtime($f_original) > filemtime($f_compiled))) {
					$r = $this->compile($f_original, $f_compiled);
				}
			} else {
				$r = $this->compile($f_original, $f_compiled);
			}

			$this->filename = $f_compiled;
			if (!$r) {
				throw new \Exception(sprintf(self::E_TEMPLATE_COMPILE, $filename));
			}
		}

		$this->source = $filename;
		return (bool)$r;
	}

	/** Compile template */
	function compile($s, $d) {
		$c = new Compiler;

		$c->set_hooks($this->hooks);
		$c->set_escape($this->_escape);
		return $c->compile($s, $d, array($this, "_find_path"), $this->_nocache);
	}

	/** Enable or disable auto-escaping of printed variables */
	function set_escape($escape) {
		$this->_escape = $escape ? true : false;
	}

	/** Sets searchable template paths */
	function set_paths($paths = false) {
		if ($paths === false) {
			$paths = array("templates/");
		}

		if (is_string($paths)) {
			$paths = func_get_args();
		}

		$this->_paths = $paths;
	}

	/** Sets compile location */
	function set_compile_location($path, $is_absolute) {
		$this->_compile_location = rtrim($path, "/") . "/";
		$this->_compile_absolute = $is_absolute;
	}

	/** Compile path calculation */
	function _compile_path($path) {
		if ($this->_compile_absolute) {
			return $this->_compile_location . $path;
		}

		return $path . $this->_compile_location;
	}

	/** Finds first path with existing template file */
	function _find_path($filename) {
		foreach ($this->_paths as $path) {
			// even if only compiled template exists, it's ok
			if (file_exists($path . $filename) || file_exists($this->_compile_path($path) . $filename)) {
				return $path;
			}
		}

		return false;
	}

	/** Assign data to the template */
	function assign($key, $value = '') {
		if (is_array($key)) {
			// $key is an array, use $value as prefix if set
			if ($value != '') {
				$value .= '_';
			}

			foreach ($key as $k => $v) {
				$this->vars[$value . $k] = $v;
			}
		} else {
			// $key is a string, do stuff depending on value and prefix
			$concat = ($key[0] == '.');
			if ($concat) {
				$key = substr($key, 1);
			}

			$this->vars[$key] = ($concat ? (is_array($value) ? array_merge($this->vars[$key], $value) : $this->vars[$key] . $value) : $value);
		}

		return ""; // {$this->assign} calls, ouch
	}

	/** Get variable */
	function getVar($key) {
		return isset($this->vars[$key]) ? $this->vars[$key] : false;
	}

	/** Render the template to standard output */
	function render() {
		if ($this->filename === false) {
			throw new \Exception(sprintf(self::E_FILENAME_EMPTY, $this->source));
		}

		include($this->filename);
	}

	/** Render the template and return text */
	function get() {
		ob_start();
		$this->render();
		$s = ob_get_contents();

		ob_end_clean();
		return $s;
	}
}
