(function () {
  "use strict";

  function node(tag, className, text) {
    var el = document.createElement(tag);
    if (className) {
      el.className = className;
    }
    if (text !== undefined && text !== null) {
      el.textContent = text;
    }
    return el;
  }

  function details(className, summary, text) {
    var box = node("details", className);
    box.appendChild(node("summary", "", summary));
    box.appendChild(node("pre", "", text));
    return box;
  }

  function workEvent(event) {
    var data = event.data || {};
    switch (event.kind) {
      case "status":
        return node("div", "work-status", event.body || "working");
      case "text":
      case "result":
        return node("div", "work-text prose", event.body || "");
      case "thinking":
        return details("work-thinking", "Thinking", event.body || "");
      case "tool_use":
        return details("work-tool", data.name || "Tool", JSON.stringify(data.input || {}, null, 2));
      case "tool_result":
        return details("work-result" + (data.is_error ? " error" : ""), data.is_error ? "Error" : "Result", event.body || "");
      case "system":
        return details("work-system", "Session", event.body || "");
      case "error":
        return node("div", "work-error", event.body || "The run failed.");
      default:
        return null;
    }
  }

  function nearBottom() {
    return document.documentElement.scrollHeight - window.scrollY - window.innerHeight < 240;
  }

  function runOf(work, id) {
    var runs = work.querySelectorAll(".work-run");
    for (var i = 0; i < runs.length; i++) {
      if (runs[i].dataset.run === id) {
        return runs[i];
      }
    }
    var run = node("div", "work-run");
    run.dataset.run = id;
    work.appendChild(run);
    return run;
  }

  function addWork(data) {
    var message = document.getElementById("note-" + data.question);
    if (!message) {
      return;
    }
    var work = message.querySelector(".work");
    if (!work) {
      work = node("div", "work");
      message.insertBefore(work, message.querySelector(".said-audio"));
    }
    var run = runOf(work, data.run);
    if (run.querySelector('[data-seq="' + data.seq + '"]')) {
      return;
    }
    var item = workEvent(data);
    if (!item) {
      return;
    }
    item.dataset.seq = data.seq;
    run.appendChild(item);
    if (nearBottom()) {
      item.scrollIntoView({ block: "nearest" });
    }
  }

  function mark(note) {
    var message = document.getElementById("note-" + note.id);
    if (!message) {
      return;
    }
    message.classList.toggle("working", note.state === "working");
    var badge = message.querySelector(".note-state");
    if (badge && !note.state) {
      badge.remove();
      return;
    }
    if (!badge && note.state) {
      badge = node("span", "badge note-state " + note.state, note.state);
      message.querySelector("header").appendChild(badge);
      return;
    }
    if (badge) {
      badge.className = "badge note-state " + note.state;
      badge.textContent = note.state;
    }
  }

  function follow(talk) {
    var source = new EventSource("/api/sse/thread:" + encodeURIComponent(talk.dataset.thread));
    source.addEventListener("event", function (incoming) {
      var data;
      try {
        data = JSON.parse(incoming.data);
      } catch (err) {
        return;
      }
      addWork(data);
    });
    source.addEventListener("note", function (incoming) {
      var note;
      try {
        note = JSON.parse(incoming.data);
      } catch (err) {
        return;
      }
      mark(note);
      if (note.parent_id && (note.state === "done" || note.state === "failed")) {
        window.location.reload();
      }
    });
  }

  function pick(form) {
    form.addEventListener("submit", function (event) {
      event.preventDefault();
      var select = form.querySelector("select");
      var button = form.querySelector("button");
      if (!select) {
        return;
      }
      if (button) {
        button.disabled = true;
      }
      fetch("/api/threads/" + encodeURIComponent(form.dataset.thread) + "/agent", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ agent: select.value })
      }).then(function (response) {
        if (!response.ok) {
          throw new Error("save failed");
        }
        form.classList.remove("failed");
        form.classList.add("saved");
      }).catch(function () {
        form.classList.remove("saved");
        form.classList.add("failed");
      }).finally(function () {
        if (button) {
          button.disabled = false;
        }
      });
    });
  }

  var talk = document.querySelector(".talk[data-thread]");
  if (talk && talk.dataset.thread && window.EventSource) {
    follow(talk);
  }
  var form = document.querySelector("[data-agent-form]");
  if (form) {
    pick(form);
  }
})();
