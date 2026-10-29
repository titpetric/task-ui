/* The refinement module's own script, loaded from its public tree.
 *
 * It defines nothing the page already has. task-ui's own task-ui.js holds the
 * form contract, so a form with data-api posts itself as JSON, a data-next
 * navigates or reloads, and a failure raises a toast: this file adds only what
 * that contract has no spelling for.
 *
 * The globals it reads, all from task-ui.js: api(method, url, body) for a call
 * the form contract does not cover, and toast(text, tone) for feedback. Each
 * is checked before use, so a page that loads this script without them does
 * nothing at all.
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
		document.querySelectorAll(".refinement-thread").forEach(followRuns);
		document.querySelectorAll(".refinement-report").forEach(previewArtifacts);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", start);
	} else {
		start();
	}
})();
