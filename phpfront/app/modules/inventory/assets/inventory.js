// The Inventory's own forms: the one that enters a host and the one that
// removes it. Both are plain posts, so the page works with no script at all;
// this asks before a removal and says what the probe is doing after an entry.
(function () {
	"use strict";

	// Entering a host waits on an ssh login, up to twenty seconds. The button
	// says what it is waiting for while the post is out.
	function waitOnSubmit(form) {
		form.addEventListener("submit", function () {
			var button = form.querySelector("button[type=submit]");

			if (!button) return;
			button.disabled = true;
			button.textContent = "Signing in over ssh";
		});
	}

	function start() {
		document.querySelectorAll("form.inventory-add").forEach(waitOnSubmit);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", start);
	} else {
		start();
	}
})();
