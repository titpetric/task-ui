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

  // The Threads page draws the list and the selected conversation apart, and
  // reads a conversation in place: /threads/<id>?pane=1 answers the pane
  // without the shell. The thread page has no pane element and reads whole.
  var pane = document.querySelector("[data-thread-pane]");
  var list = document.querySelector(".threads-list");
  var blank = pane && pane.querySelector(".threads-new") ? pane.innerHTML : "";
  var stream = null;
  var current = "";

  // follow draws the work of the thread the pane shows and reads its notes.
  // One stream is open at a time: a swap closes the one it replaces.
  function follow(talk) {
    if (stream) {
      stream.close();
      stream = null;
    }
    if (!talk || !talk.dataset.thread || !window.EventSource) {
      return;
    }
    var id = talk.dataset.thread;
    var source = new EventSource("/api/sse/thread:" + encodeURIComponent(id));
    stream = source;
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
        // The answer is a reply now, so the conversation is read again.
        if (pane) {
          load(id, false);
        } else {
          window.location.reload();
        }
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

  // select paints the row of the thread the pane shows.
  function select(id) {
    if (!list) {
      return;
    }
    list.querySelectorAll(".thread-row").forEach(function (row) {
      var link = row.querySelector(".thread-open");
      row.classList.toggle("selected", !!id && !!link && link.dataset.threadId === id);
    });
  }

  // want counts what the pane was last asked to show: a read that answers
  // after the reader moved on is dropped, not drawn over the newer pane.
  var want = 0;

  // load reads one conversation into the pane, dropping the editors of the
  // one it replaces.
  function load(id, push) {
    if (!pane || !id) {
      return Promise.resolve();
    }
    var seq = ++want;
    return fetch("/threads/" + encodeURIComponent(id) + "?pane=1", {
      credentials: "same-origin",
      headers: { Accept: "text/html" }
    }).then(function (res) {
      return res.text();
    }).then(function (html) {
      if (seq !== want) {
        return;
      }
      if (window.taskUI && window.taskUI.unmount) {
        window.taskUI.unmount(pane);
      }
      pane.innerHTML = html;
      if (window.taskUI && window.taskUI.mount) {
        window.taskUI.mount(pane);
      }
      current = id;
      select(id);
      if (window.taskUI && window.taskUI.read) {
        window.taskUI.read();
      }
      follow(pane.querySelector(".talk[data-thread]"));
      var form = pane.querySelector("[data-agent-form]");
      if (form) {
        pick(form);
      }
      if (push) {
        history.pushState({ thread: id }, "", "/threads?id=" + encodeURIComponent(id));
      }
    });
  }

  // newThread restores the box that starts a thread, as the page drew it. A
  // page that opened on a thread has no box to restore, so it reads one.
  function newThread(push) {
    if (!pane) {
      return;
    }
    if (!blank) {
      window.location.assign("/threads");
      return;
    }
    want++;
    if (stream) {
      stream.close();
      stream = null;
    }
    if (window.taskUI && window.taskUI.unmount) {
      window.taskUI.unmount(pane);
    }
    pane.innerHTML = blank;
    if (window.taskUI && window.taskUI.mount) {
      window.taskUI.mount(pane);
    }
    current = "";
    select("");
    if (push) {
      history.pushState({ thread: "" }, "", "/threads");
    }
  }

  if (list && pane) {
    list.addEventListener("click", function (event) {
      var link = event.target.closest(".thread-open");
      if (!link || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
        return;
      }
      event.preventDefault();
      load(link.dataset.threadId, true);
    });
  }

  window.addEventListener("popstate", function () {
    if (!pane) {
      return;
    }
    var id = new URLSearchParams(location.search).get("id") || "";
    if (id === current) {
      return;
    }
    if (id) {
      load(id, false);
    } else {
      newThread(false);
    }
  });

  // A reply in the pane, or a thread started in its box, arrives as an event
  // from the shared note handler.
  document.addEventListener("taskui:pane", function () {
    if (pane && current) {
      load(current, false);
    }
  });
  document.addEventListener("taskui:thread", function (event) {
    if (pane && event.detail && event.detail.id) {
      load(event.detail.id, true);
    }
  });
  // The list is redrawn without a selection of its own, so the row the pane
  // shows is marked again.
  document.addEventListener("taskui:notes", function () {
    if (pane) {
      select(current);
    }
  });

  var talk = pane ? pane.querySelector(".talk[data-thread]") : document.querySelector(".talk[data-thread]");
  if (talk) {
    current = talk.dataset.thread || "";
    follow(talk);
  }
  var form = (pane || document).querySelector("[data-agent-form]");
  if (form) {
    pick(form);
  }
})();
