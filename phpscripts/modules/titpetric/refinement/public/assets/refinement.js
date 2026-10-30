/* The refinement module's own script, loaded from its public tree.
 *
 * task-ui's own task-ui.js holds the form contract a page writes against, but
 * it reaches no global: its api(), toast() and confirm() are inside its
 * function and nothing on window names them. A module's forms are therefore
 * ordinary forms that post and redirect, and what this file adds is what a
 * plain form has no spelling for.
 */
(function () {
	"use strict";

	// Enter sends a comment and Shift with it writes a newline, the way a chat
	// box takes them. A comment is one paragraph, so the submit is the common
	// case and a textarea's default puts it behind the button.
	function sendOnEnter(form) {
		var field = form.querySelector("textarea[name=comment]");
		if (!field) return;
		field.addEventListener("keydown", function (event) {
			if (event.key !== "Enter" || event.shiftKey || event.altKey) return;
			event.preventDefault();
			if (field.value.trim() === "") return;
			form.requestSubmit ? form.requestSubmit() : form.dispatchEvent(new Event("submit", { cancelable: true, bubbles: true }));
		});
	}

	// A run started by a comment takes minutes. The thread reloads while one
	// is going, so the artifacts and the report appear without the reader
	// pressing anything, and stops once none is.
	function followRuns(thread) {
		var going = thread.querySelectorAll("[data-run-state=running]").length;
		if (!going) return;
		window.setTimeout(function () {
			window.location.reload();
		}, 15000);
	}

	// A form that deletes asks first. The question is on the form, because
	// the one on the design says that the thread goes with it and the one on
	// a rule says that only the request does.
	function confirmOnSubmit(form) {
		form.addEventListener("submit", function (event) {
			if (!window.confirm(form.dataset.confirm)) event.preventDefault();
		});
	}

	// The artifact index the run wrote is markdown rendered to HTML by the
	// module, so a link in it is an ordinary link. One that names an artifact
	// of this instance opens in the preview window the page already has.
	function previewArtifacts(report) {
		report.addEventListener("click", function (event) {
			var link = event.target.closest("a[href^='/artifacts/']");
			if (!link || typeof window.api !== "function") return;
			// The page's own artifact preview takes the hash; letting the
			// browser follow the link would leave the thread.
			event.preventDefault();
			window.location.hash = "file=" + encodeURIComponent(link.getAttribute("href"));
		});
	}

	function start() {
		document.querySelectorAll("form.refinement-comment-form").forEach(sendOnEnter);
		document.querySelectorAll("form[data-confirm]").forEach(confirmOnSubmit);
		document.querySelectorAll(".refinement-thread").forEach(followRuns);
		document.querySelectorAll(".refinement-report").forEach(previewArtifacts);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", start);
	} else {
		start();
	}
})();
