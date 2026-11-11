// The Inventory's own forms: the one that enters a host and the one that
// removes it. Both are plain posts, so the page works with no script at all;
// this asks before a removal and says what the probe is doing after an entry.
(function () {
	"use strict";

	// A form that removes asks first. The question is on the form, because it
	// names the host and says whether the hub brings it back.
	function confirmOnSubmit(form) {
		form.addEventListener("submit", function (event) {
			if (!window.confirm(form.dataset.confirm)) event.preventDefault();
		});
	}

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
		document.querySelectorAll("form.inventory-drop[data-confirm]").forEach(confirmOnSubmit);
		document.querySelectorAll("form.inventory-add").forEach(waitOnSubmit);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", start);
	} else {
		start();
	}
})();
