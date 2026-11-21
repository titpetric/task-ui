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

  // scroller is the box the talk scrolls in: the pane where it has its own
  // scrollbar, the page otherwise.
  function scroller() {
    var pane = document.querySelector(".chat-pane");
    if (pane) {
      var overflow = getComputedStyle(pane).overflowY;
      if (overflow === "auto" || overflow === "scroll") {
        return pane;
      }
    }
    return document.scrollingElement || document.documentElement;
  }

  function nearBottom() {
    var box = scroller();
    return box.scrollHeight - box.scrollTop - box.clientHeight < 240;
  }

  function scrollEnd(behavior) {
    var box = scroller();
    if (box.scrollTo) {
      box.scrollTo({ top: box.scrollHeight, behavior: behavior || "auto" });
    } else {
      box.scrollTop = box.scrollHeight;
    }
  }

  // latest shows the button back to the talk while the reader is above it,
  // centred over the column the talk scrolls in.
  function latest() {
    var button = document.querySelector("[data-scroll-latest]");
    if (!button) {
      return;
    }
    var box = scroller();
    button.hidden = nearBottom();
    if (box === document.scrollingElement) {
      button.style.left = "50%";
      return;
    }
    var rect = box.getBoundingClientRect();
    button.style.left = rect.left + rect.width / 2 + "px";
    var composer = document.querySelector(".chat-pane .composer-box");
    if (composer) {
      button.style.bottom = window.innerHeight - composer.getBoundingClientRect().top + 56 + "px";
    }
  }

  function icon(name) {
    var el = document.createElement("i");
    el.className = "ic ic-" + name;
    el.setAttribute("aria-hidden", "true");
    return el;
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

  // statusLine is the state block of the answer that is still being written.
  function statusLine() {
    return document.querySelector(".said-live[data-status]");
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
    var line = statusLine();
    var elapsedText = line && line.querySelector("[data-status-elapsed]");
    if (elapsedText) {
      elapsedText.textContent = elapsed(line.dataset.started);
    }
  }

  // agentList is the status block's list of subagents, made on the first.
  function agentList() {
    var line = statusLine();
    if (!line) {
      return null;
    }
    var list = line.querySelector("[data-status-agents]");
    if (!list) {
      list = node("ul", "status-agents");
      list.dataset.statusAgents = "";
      line.insertBefore(list, line.querySelector(".said-live-foot"));
    }
    return list;
  }

  // addAgent draws the subagent a Task call starts, with its brief behind it.
  function addAgent(event) {
    if (!event.tool_id) {
      return;
    }
    var list = agentList();
    if (!list || list.querySelector('[data-agent="' + event.tool_id + '"]')) {
      return;
    }
    var input = {};
    try {
      input = JSON.parse(event.input || "{}") || {};
    } catch (err) {
      input = {};
    }
    var item = node("li");
    var box = node("details", "status-agent working");
    box.dataset.agent = event.tool_id;
    var summary = node("summary");
    summary.appendChild(icon("users"));
    summary.appendChild(node("span", "", input.description || "A subagent"));
    if (input.subagent_type) {
      summary.appendChild(node("small", "faint", input.subagent_type));
    }
    var badge = node("span", "badge signal", "working");
    badge.dataset.statusAgentState = "";
    summary.appendChild(badge);
    box.appendChild(summary);
    if (input.prompt) {
      box.appendChild(node("pre", "status-agent-prompt", input.prompt));
    }
    item.appendChild(box);
    list.appendChild(item);
    if (nearBottom()) {
      list.scrollIntoView({ block: "nearest" });
    }
  }

  // settleAgent marks the subagent of a finished Task call, with its answer.
  function settleAgent(event) {
    var box = event.tool_id ? document.querySelector('.status-agent[data-agent="' + event.tool_id + '"]') : null;
    if (!box) {
      return;
    }
    box.classList.remove("working");
    box.classList.add(event.error ? "failed" : "done");
    var badge = box.querySelector("[data-status-agent-state]");
    if (badge) {
      badge.className = "badge " + (event.error ? "bad" : "ok");
      badge.textContent = event.error ? "failed" : "done";
    }
    if (event.body) {
      box.appendChild(node("pre", "status-agent-result", event.body));
    }
  }

  // liveBox is the answer of the thread that is still being written.
  function liveBox() {
    var text = document.querySelector("[data-work-progress]");
    return text ? text.closest(".said") : null;
  }

  // addProgress puts the text a live run last printed in its answer, in place of the one before.
  function addProgress(event) {
    var box = liveBox();
    var text = box && box.querySelector("[data-work-progress]");
    if (!text || !event.body) {
      return;
    }
    text.classList.remove("faint");
    text.textContent = event.body;
    if (nearBottom()) {
      scrollEnd();
    }
  }

  // addTool counts a tool call into the live answer's stats.
  function addTool() {
    var box = liveBox();
    var stats = box && box.querySelector("[data-work-stats]");
    if (!stats) {
      return;
    }
    var count = Number(stats.dataset.tools || 0) + 1;
    stats.dataset.tools = String(count);
    stats.textContent = count + (count === 1 ? " tool call" : " tool calls");
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
      if (data.kind === "text") {
        addProgress(data);
        return;
      }
      if (data.kind === "tool_use") {
        addTool();
        if (data.tool === "Task" || data.tool === "Agent") {
          addAgent(data);
        }
      }
      if (data.kind === "tool_result") {
        settleAgent(data);
      }
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
      if (note.id === current && note.state === "working" && !document.querySelector(".said-live .actions")) {
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
    var working = pane && pane.querySelector(".said-live") ? "Working · " : "";
    var at = pageTitle.indexOf(" | ");
    if (head) {
      document.title = working + head.textContent.trim() + (at < 0 ? "" : pageTitle.slice(at));
    }
  }

  // select paints the row of the thread the pane shows.
  function select(id) {
    viewing(!!id);
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
      scrollEnd();
      latest();
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
    var box = scroller();
    var y = box.scrollTop;
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
      if (bottom) {
        scrollEnd();
      } else {
        scroller().scrollTop = y;
      }
      latest();
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
    viewing(true);
    scroller().scrollTop = 0;
    latest();
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
    scrollEnd();
  }
  var form = (pane || document).querySelector("[data-agent-form]");
  if (form) {
    pick(form);
  }
  document.addEventListener("scroll", latest, true);
  window.addEventListener("resize", latest);
  document.addEventListener("click", function (event) {
    if (event.target.closest("[data-scroll-latest]")) {
      scrollEnd("smooth");
    }
  });
  setInterval(tickStatus, 1000);
  latest();
  tickStatus();
})();
