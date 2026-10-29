<?php

namespace Task;

// Module is a stub of the host class task-ui registers as `Task\Module`, for
// the linter alone. Nothing requires this file at runtime: the real class is
// Go, and `rt.RegisterConstructor("Task\\Module", ...)` publishes the methods
// of the value its constructor returns.
//
// `phpscript lint` resolves a class name against the declarations it can read,
// and a host binding is not one of them, so without this file every
// `new Task\Module` in the module is reported as an undefined class and a real
// undefined class would read as one more line of expected noise.
//
// Keep the methods in step with phpscript/module.go.
class Module {
	/** Sets the module's name, as <vendor>/<plugin>. */
	function setName($name) {
	}

	/** Sets the title of the sidebar row and the page heading. */
	function setTitle($title) {
	}

	/** Sets the line under the heading. */
	function setDescription($description) {
	}

	/** Sets the sprite symbol of the sidebar row, without the i- prefix. */
	function setIcon($icon) {
	}

	/** Sets what the module calls its own version. */
	function setVersion($version) {
	}
}
