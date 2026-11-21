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

  // workEvent draws one event the way components/work.tpl draws it. The
  // fields are the stream's own: a tool card carries tool and input, a result
  // carries error and its cut body.
  function workEvent(event) {
    switch (event.kind) {
      case "text":
        return node("div", "work-text prose", event.body || "");
      case "result":
        return null;
      case "thinking":
        return details("work-thinking", "Thinking", event.body || "");
      case "tool_use":
        return details("work-tool", event.tool || "Tool", event.input || "");
      case "tool_result":
        return details("work-result" + (event.error ? " error" : ""), event.error ? "Error" : "Result", event.body || "");
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

  // workNode is the work block of the message a run writes in: its answer,
  // which stands in the thread while the run goes, or the question itself.
  function workNode(run, question) {
    var answer = run ? document.querySelector('.talk .said.agent[data-run="' + run + '"]') : null;
    var message = answer || document.getElementById("note-" + question);
    if (!message) {
      return null;
    }
    var work = message.querySelector(".work");
    if (!work) {
      work = node("div", "work");
      message.insertBefore(work, message.querySelector(".said-audio"));
    }
    return work;
  }

  function addWork(data) {
    var work = workNode(data.run, data.question);
    if (!work) {
      return;
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

  function elapsed(started) {
    var since = Number(started);
    if (!since) {
      return "";
    }
    var seconds = Math.max(0, Math.floor((Date.now() - since) / 1000));
    if (seconds < 60) {
      return seconds + "s";
    }
    if (seconds < 3600) {
      return Math.floor(seconds / 60) + "m" + String(seconds % 60).padStart(2, "0") + "s";
    }
    return Math.floor(seconds / 3600) + "h" + String(Math.floor(seconds / 60) % 60).padStart(2, "0") + "m";
  }

  // statusLine is the pinned line above the conversation, made where the page
  // opened on a run that had not reported one yet.
  function statusLine() {
    var found = document.querySelector("[data-status]");
    if (found) {
      return found;
    }
    var talk = document.querySelector(".talk[data-thread]");
    var head = document.querySelector(".conversation-head");
    if (!talk || !head) {
      return null;
    }
    var line = node("div", "conversation-status");
    line.dataset.status = "";
    var said = node("p");
    said.setAttribute("aria-live", "polite");
    said.appendChild(node("i", "dot working"));
    said.appendChild(node("span", "", "Working"));
    said.lastChild.dataset.statusText = "";
    var clock = node("span", "faint");
    clock.dataset.statusElapsed = "";
    said.appendChild(clock);
    line.appendChild(said);
    head.parentNode.insertBefore(line, head.nextSibling);
    return line;
  }

  function setStatus(event) {
    var line = statusLine();
    if (!line) {
      return;
    }
    var label = event.body || event.tool || event.state || "Working";
    line.dataset.state = event.state || "working";
    if (!line.dataset.started && event.created_at) {
      line.dataset.started = String(Date.parse(event.created_at));
    }
    var text = line.querySelector("[data-status-text]");
    if (text) {
      text.textContent = label;
    }
  }

  function tickStatus() {
    var elapsedText = document.querySelector("[data-status-elapsed]");
    var line = document.querySelector("[data-status]");
    if (elapsedText && line) {
      elapsedText.textContent = elapsed(line.dataset.started);
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

  var pane = document.querySelector("[data-thread-pane]");
  var shell = document.querySelector(".chat-shell");
  var list = document.querySelector(".chat-list");
  var blank = pane && pane.querySelector(".chat-blank") ? pane.innerHTML : "";
  var stream = null;
  var current = "";

  function viewing(on) {
    if (shell) {
      shell.classList.toggle("viewing", on);
    }
  }

  // queueReload reads the pane once however many events ask for it: a run
  // settles with a note, a result and a status in the same tick.
  var reloadTimer = null;

  function queueReload() {
    if (!pane || !current || reloadTimer) {
      return;
    }
    reloadTimer = setTimeout(function () {
      reloadTimer = null;
      refresh(current);
    }, 150);
  }

  // follow reads the thread's work. The page's newest event is the cursor, so
  // what the bus dropped between the page and the stream is replayed.
  function follow(talk) {
    if (stream) {
      stream.close();
      stream = null;
    }
    if (!talk || !talk.dataset.thread || !window.EventSource) {
      return;
    }
    var id = talk.dataset.thread;
    var cursor = talk.dataset.lastEvent || "";
    var url = "/api/sse/thread:" + encodeURIComponent(id);
    if (cursor) {
      url += "?last_event=" + encodeURIComponent(cursor);
    }
    var source = new EventSource(url);
    stream = source;
    source.addEventListener("event", function (incoming) {
      var data;
      try {
        data = JSON.parse(incoming.data);
      } catch (err) {
        return;
      }
      if (data.kind === "status") {
        setStatus(data);
        return;
      }
      addWork(data);
      if (data.kind === "result" || data.kind === "error") {
        queueReload();
      }
    });
    source.addEventListener("note", function (incoming) {
      var note;
      try {
        note = JSON.parse(incoming.data);
      } catch (err) {
        return;
      }
      mark(note);
      if (!document.getElementById("note-" + note.id) && note.parent_id === current) {
        queueReload();
      }
      // A run that started after the page drew its line gives the line its actions.
      if (note.id === current && note.state === "working" && !document.querySelector(".conversation-status .actions")) {
        queueReload();
      }
      if (note.state === "done" || note.state === "failed" || note.state === "killed" || note.state === "stopped") {
        queueReload();
      }
    });
  }

  function pick(form) {
    var select = form.querySelector("select");
    var button = form.querySelector('button[type="submit"]');
    if (!select) {
      return;
    }
    if (button) {
      button.disabled = true;
      select.addEventListener("change", function () {
        button.disabled = false;
      });
    }
    form.addEventListener("submit", function (event) {
      event.preventDefault();
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
        queueReload();
      }).catch(function () {
        form.classList.remove("saved");
        form.classList.add("failed");
        if (button) {
          button.disabled = false;
        }
      });
    });
  }

  // pageTitle names the page the list is on, which a thread's own name joins.
  var pageTitle = document.title;

  function title(id) {
    if (!id) {
      document.title = pageTitle;
      return;
    }
    var head = pane && pane.querySelector(".conversation-title h2");
    var working = pane && pane.querySelector(".conversation-status") ? "Working · " : "";
    var at = pageTitle.indexOf(" | ");
    if (head) {
      document.title = working + head.textContent.trim() + (at < 0 ? "" : pageTitle.slice(at));
    }
  }

  // select paints the row of the thread the pane shows.
  function select(id) {
    if (!list) {
      return;
    }
    list.querySelectorAll(".thread-row").forEach(function (row) {
      var link = row.querySelector(".thread-open");
      var on = !!id && !!link && link.dataset.threadId === id;
      row.classList.toggle("selected", on);
      if (link) {
        if (on) {
          link.setAttribute("aria-current", "true");
        } else {
          link.removeAttribute("aria-current");
        }
      }
    });
    viewing(!!id);
  }

  // want counts what the pane was last asked to show: a read that answers
  // after the reader moved on is dropped, not drawn over the newer pane.
  var want = 0;

  function paint(id) {
    if (window.taskUI && window.taskUI.mount) {
      window.taskUI.mount(pane);
    }
    current = id;
    select(id);
    title(id);
    if (window.taskUI && window.taskUI.read) {
      window.taskUI.read();
    }
    var form = pane.querySelector("[data-agent-form]");
    if (form) {
      pick(form);
    }
  }

  // load reads one conversation into the pane and follows it.
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
      paint(id);
      follow(pane.querySelector(".talk[data-thread]"));
      window.scrollTo(0, document.documentElement.scrollHeight);
      if (push) {
        history.pushState({ thread: id }, "", "/threads/" + encodeURIComponent(id));
      }
    });
  }

  // refresh reads the same conversation again without dropping its stream,
  // and keeps the reader where they were unless they were at the end. A reply
  // just sent puts the caret back in the box it left.
  function refresh(id, focus) {
    if (!pane || !id) {
      return Promise.resolve();
    }
    var seq = ++want;
    var bottom = nearBottom();
    var y = window.scrollY;
    return fetch("/threads/" + encodeURIComponent(id) + "?pane=1", {
      credentials: "same-origin",
      headers: { Accept: "text/html" }
    }).then(function (res) {
      return res.text();
    }).then(function (html) {
      if (seq !== want || id !== current) {
        return;
      }
      if (window.taskUI && window.taskUI.unmount) {
        window.taskUI.unmount(pane);
      }
      pane.innerHTML = html;
      paint(id);
      window.scrollTo(0, bottom ? document.documentElement.scrollHeight : y);
      if (focus) {
        var area = pane.querySelector("form[data-send] textarea");
        if (area) {
          (area.mdeditor ? area.mdeditor.surface : area).focus();
        }
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
    paint("");
    window.scrollTo(0, 0);
    if (push) {
      history.pushState({ thread: "" }, "", "/threads");
    }
  }

  document.addEventListener("click", function (event) {
    var button = event.target.closest("[data-new-thread]");
    if (!button || !pane) {
      return;
    }
    event.preventDefault();
    newThread(true);
    var area = pane.querySelector("textarea");
    if (area) {
      (area.mdeditor ? area.mdeditor.surface : area).focus();
    }
  });

  // "/" focuses the search, as the list is reached before the pane.
  document.addEventListener("keydown", function (event) {
    if (event.key !== "/" || event.metaKey || event.ctrlKey || event.altKey || document.querySelector("dialog[open]")) {
      return;
    }
    var active = document.activeElement;
    if (active && (/^(INPUT|TEXTAREA|SELECT)$/.test(active.tagName) || active.isContentEditable)) {
      return;
    }
    var search = document.querySelector(".chat-search");
    if (!search) {
      return;
    }
    event.preventDefault();
    search.focus();
    search.select();
  });

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
    var found = location.pathname.match(/\/threads\/([^\/]+)\/?$/);
    var id = found ? decodeURIComponent(found[1]) : "";
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
      refresh(current, true);
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
    select(current);
    title(current);
    window.scrollTo(0, document.documentElement.scrollHeight);
  }
  var form = (pane || document).querySelector("[data-agent-form]");
  if (form) {
    pick(form);
  }
  setInterval(tickStatus, 1000);
  tickStatus();
})();
