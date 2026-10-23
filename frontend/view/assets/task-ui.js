// task-ui: the run dialog, the forms, and the live parts of the pages.
//
// A run is a session on the server; the dialog is a window on it. Opening
// the dialog attaches, closing it detaches, and the run goes on either way.
// Binary frames are terminal bytes, text frames are JSON control messages.
// Every change a page makes goes through the JSON API under /api.
(function () {
  "use strict";

  var dialog = document.getElementById("run-dialog");
  var els = dialog && {
    kind: document.getElementById("run-kind"),
    title: document.getElementById("run-title"),
    who: document.getElementById("run-who"),
    state: document.getElementById("run-state"),
    elapsed: document.getElementById("run-elapsed"),
    host: document.getElementById("run-terminal"),
    notice: document.getElementById("run-notice"),
    hint: document.getElementById("run-hint"),
    attach: document.getElementById("run-attach"),
    force: document.getElementById("run-force"),
    again: document.getElementById("run-again"),
    stop: document.getElementById("run-stop"),
    kill: document.getElementById("run-kill"),
    close: document.getElementById("run-close"),
    page: document.getElementById("run-page"),
    reason: document.getElementById("run-reason"),
    reasonText: document.getElementById("run-reason-text"),
    rec: document.getElementById("run-rec"),
    replay: document.getElementById("run-replay"),
    pause: document.getElementById("run-pause"),
    restart: document.getElementById("run-restart"),
  };

  var session = null; // { run, spec, ws, term, fit, timer, onResize, finished, changed }

  // ------------------------------------------------------------------ api

  function api(method, url, body) {
    var init = { method: method, credentials: "same-origin", headers: { Accept: "application/json" } };
    if (body !== undefined) {
      init.headers["Content-Type"] = "application/json";
      init.body = JSON.stringify(body);
    }
    return fetch(url, init).then(function (res) {
      if (res.status === 204) return { status: res.status, data: {} };
      return res.text().then(function (text) {
        var data = {};
        try { data = text ? JSON.parse(text) : {}; } catch (e) { data = { error: text }; }
        return { status: res.status, data: data };
      });
    });
  }

  function escapeHTML(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }

  function initials(name) {
    var words = String(name || "?").trim().split(/[\s._@-]+/).filter(Boolean);
    var text = words.length >= 2 ? words[0].charAt(0) + words[1].charAt(0) : (words[0] || "?").slice(0, 2);
    return text.toUpperCase();
  }

  // The role of each user by name, which the page carries on its body.
  var roles = (function () {
    try {
      return JSON.parse(document.body.dataset.roles || "{}") || {};
    } catch (e) {
      return {};
    }
  })();

  // The named workspaces, which the page carries on its body.
  var workspaces = (function () {
    try {
      return JSON.parse(document.body.dataset.workspaces || "{}") || {};
    } catch (e) {
      return {};
    }
  })();

  // folderHref returns the path of a folder's page, as Page.Href does: a
  // folder of a named workspace by its ID, a project at the root under
  // /f/_/.
  function folderHref(id) {
    var top = String(id).split("/")[0];
    if (id === "_") return "/f/_";
    if (Object.prototype.hasOwnProperty.call(workspaces, top)) return "/f/" + id;
    return "/f/_/" + id;
  }

  // avatarHTML draws the initials of a user in the colour of their role,
  // as the avatar template does.
  function avatarHTML(name) {
    var role = Object.prototype.hasOwnProperty.call(roles, name) ? roles[name] : "";
    return '<span class="avatar sm' + (role ? " role-" + escapeHTML(role) : "") + '" title="' + escapeHTML(name) + '">' + escapeHTML(initials(name)) + "</span>";
  }

  // userHTML draws a user's avatar and name with the [...] menu next to
  // them, as the user template does: the profile, for yourself or with
  // user/view; the runs by the user; their permissions, with user/edit.
  // The body says who is signed in and what they may see of users; a
  // name no user holds, such as the agent a note names, has no profile.
  function userHTML(name) {
    var me = document.body.dataset.me || "", may = document.body.dataset.userMenu || "";
    var q = encodeURIComponent(name), who = escapeHTML(name);
    var known = Object.prototype.hasOwnProperty.call(roles, name);
    var item = function (href, icon, text) {
      return '<a class="more-item" role="menuitem" href="' + href + '"><svg class="ic"><use href="#i-' + icon + '"/></svg>' + text + "</a>";
    };
    return '<span class="user-ref">' + avatarHTML(name) + ' <span class="user-name">' + who + "</span>" +
      '<details class="more-menu user-more"><summary class="btn-icon" title="More on ' + who + '" aria-label="More on ' + who + '"><svg class="ic"><use href="#i-more"/></svg></summary>' +
      '<div class="more-list" role="menu">' +
      (known && (name === me || may) ? item("/users/" + q, "users", "Profile") : "") +
      item("/runs?user=" + q, "list", "Runs by this user") +
      (known && may === "edit" ? item("/users/" + q + "#permissions", "key", "Permissions") : "") +
      "</div></details></span>";
  }

  // A user's menu sits in rows that scroll sideways, which would clip
  // it: it opens fixed under its button, kept on the screen, and closes
  // when the page scrolls.
  document.addEventListener("toggle", function (event) {
    var menu = event.target;
    if (!menu.classList || !menu.classList.contains("user-more") || !menu.open) return;
    var list = menu.querySelector(".more-list"), at = menu.querySelector("summary").getBoundingClientRect();
    list.style.top = at.bottom + 4 + "px";
    list.style.left = Math.max(8, Math.min(at.left, window.innerWidth - list.offsetWidth - 8)) + "px";
  }, true);
  window.addEventListener("scroll", function () {
    document.querySelectorAll("details.user-more[open]").forEach(function (menu) { menu.removeAttribute("open"); });
  }, true);

  // A row with data-href opens that page on a click anywhere in it, but
  // not on a link, a button or a menu of its own, nor when text is
  // selected.
  document.addEventListener("click", function (event) {
    var row = event.target.closest("tr[data-href]");
    if (!row || event.target.closest("a, button, summary, details, input, label, select, textarea")) return;
    if (String(window.getSelection() || "") !== "") return;
    window.location.href = row.dataset.href;
  });

  // ------------------------------------------------------------- terminal

  function css(name) {
    return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  }

  function theme() {
    var light = window.matchMedia && window.matchMedia("(prefers-color-scheme: light)").matches;
    var base = {
      background: css("--term-bg"),
      foreground: css("--term-fg"),
      cursor: css("--signal"),
      cursorAccent: css("--term-bg"),
      selectionBackground: light ? "rgba(154,93,13,0.25)" : "rgba(224,164,79,0.3)",
    };
    var palette = light
      ? ["#23201a", "#a63232", "#1c6f56", "#7d5800", "#2f5d8a", "#7a3f8a", "#1f6f78", "#3a352c",
         "#857e6e", "#c0392b", "#228b68", "#9a5d0d", "#3a74ad", "#9450a8", "#2a8a95", "#23201a"]
      : ["#10161d", "#e06c6c", "#57b894", "#e0a44f", "#6ea8dc", "#c38fd9", "#5fc1c9", "#d5dee7",
         "#5a6b79", "#f08a8a", "#78d1ae", "#f0c070", "#93c1ea", "#d7aee8", "#86d8de", "#ffffff"];
    var names = ["black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"];
    names.forEach(function (n, i) {
      base[n] = palette[i];
      base["bright" + n.charAt(0).toUpperCase() + n.slice(1)] = palette[i + 8];
    });
    return base;
  }

  function newTerminal() {
    els.host.innerHTML = "";
    var term = new window.Terminal({
      fontFamily: css("--font") || "monospace",
      fontSize: 13,
      lineHeight: 1.15,
      cursorBlink: true,
      convertEol: false,
      scrollback: 10000,
      theme: theme(),
    });
    var fit = new window.FitAddon.FitAddon();
    term.loadAddon(fit);
    term.open(els.host);
    try { fit.fit(); } catch (e) { /* hidden host */ }
    return { term: term, fit: fit };
  }

  // --------------------------------------------------------------- dialog

  function setState(state, text) {
    els.state.innerHTML = "";
    if (!state) return;
    var badge = document.createElement("span");
    var cls = { running: "signal", ok: "ok", failed: "bad", killed: "bad", error: "bad", busy: "signal", stopped: "" }[state];
    badge.className = "badge" + (cls ? " " + cls : "");
    badge.textContent = text || state;
    els.state.appendChild(badge);
  }

  function notice(text, tone) {
    els.notice.textContent = text || "";
    els.notice.className = "run-notice" + (text ? " show" : "") + (tone ? " " + tone : "");
  }

  function elapsedText(ms) {
    var s = Math.floor(ms / 1000);
    var m = Math.floor(s / 60);
    return (m > 0 ? m + "m" : "") + String(s % 60).padStart(m > 0 ? 2 : 1, "0") + "s";
  }

  function startClock(since) {
    stopClock();
    var start = since || Date.now();
    var tick = function () { els.elapsed.textContent = elapsedText(Date.now() - start); };
    tick();
    session.timer = setInterval(tick, 1000);
  }

  function stopClock() {
    if (session && session.timer) {
      clearInterval(session.timer);
      session.timer = null;
    }
  }

  function buttons(opts) {
    els.attach.hidden = !opts.attach;
    els.force.hidden = !opts.force;
    els.again.hidden = !opts.again;
    els.stop.hidden = !opts.stop;
    els.kill.hidden = !opts.kill;
    els.pause.hidden = !opts.pause;
    els.restart.hidden = !opts.restart;
  }

  // teardown closes the window: the socket and the terminal go, the run
  // does not. A page that showed a run that changed is refreshed.
  function teardown() {
    if (!session) return;
    stopClock();
    if (session.stop) session.stop();
    els.rec.hidden = true;
    dialog.classList.remove("replaying");
    if (session.ws && session.ws.readyState <= 1) session.ws.close();
    if (session.term) session.term.dispose();
    if (session.onResize) window.removeEventListener("resize", session.onResize);
    var changed = session.changed;
    session = null;
    if (/^#run=/.test(location.hash)) history.replaceState(null, "", location.pathname + location.search);
    // A page that follows its topic drew the change already.
    if (changed && !follows) window.location.reload();
  }

  function open(kind, title, who) {
    if (session) teardown();
    els.kind.textContent = kind;
    els.title.textContent = title;
    els.who.textContent = who || "";
    els.elapsed.textContent = "";
    els.hint.textContent = "Closing this window doesn't stop the run.";
    setState(null);
    notice("");
    buttons({});
    els.reason.hidden = true;
    els.reasonText.value = "";
    pageLink("");
    els.rec.hidden = true;
    els.replay.hidden = true;
    if (!dialog.open) dialog.showModal();
    session = newTerminal();
    session.term.focus();
    return session;
  }

  // pageLink points Open run page at the run of the window, and hides it
  // while the window has no run yet.
  function pageLink(id) {
    if (!els.page) return;
    els.page.hidden = !id;
    els.page.href = "/runs/" + encodeURIComponent(id || "");
  }

  // ------------------------------------------------------------------ runs

  // start runs a job: the dialog opens first so the terminal size is known,
  // then the run is started and attached to.
  function start(spec) {
    var s = open("Run", spec.title, "");
    s.spec = spec;
    setState("running", "starting");
    var body = { folder: spec.folder, job: spec.job, force: !!spec.force, reason: spec.reason || "", cols: s.term.cols, rows: s.term.rows };
    // A job of a directory inside the folder names it; arguments go as
    // the list a run row keeps, or the line typed.
    if (spec.dir) body.dir = spec.dir;
    if (spec.args && spec.args.length) body.args = spec.args;
    // Run and summarize: Claude summarises the run on its page when it
    // ends.
    if (spec.summarize) body.summarize = true;
    api("POST", "/api/runs", body).then(function (res) {
      if (session !== s) return;
      if (res.status === 201) {
        s.changed = true;
        connect(s, res.data);
        return;
      }
      if (res.data.out_of_hours) {
        // The job's hours do not allow a run now. Force needs a reason,
        // which the run keeps: the log entry of why.
        s.outOfHours = true;
        setState("busy", "out of hours");
        notice("Out of hours: " + spec.title + " runs " + (res.data.hours || "never").replace(/\n/g, ", ") + ". To run it anyway, give a reason and press Run force.", "signal");
        els.hint.textContent = "task-ui saves the reason with the run, and everyone can see it.";
        els.reason.hidden = false;
        els.reasonText.focus();
        buttons({ force: true });
        return;
      }
      if (res.status === 409 && res.data.run) {
        s.run = res.data.run;
        setState("busy", "already running");
        notice("Already running, started by " + (res.data.run.user_name || "someone") + ". Attach to follow it, or press Run force to kill it and start over.", "signal");
        buttons({ attach: true, force: true });
        return;
      }
      setState("error");
      notice(res.data.error || ("HTTP " + res.status), "bad");
    }).catch(function (err) {
      if (session !== s) return;
      setState("error");
      notice(err.message, "bad");
    });
  }

  // attach opens a window on a run: live over the socket, finished from
  // the recording.
  function attach(id) {
    var s = open("Run", "", "");
    s.id = id;
    pageLink(id);
    history.replaceState(null, "", location.pathname + location.search + "#run=" + id);
    setState("running", "loading");
    api("GET", "/api/runs/" + encodeURIComponent(id)).then(function (res) {
      if (session !== s) return;
      if (res.status !== 200) {
        setState("error");
        notice(res.data.error || ("HTTP " + res.status), "bad");
        return;
      }
      var run = res.data;
      els.title.textContent = runTitle(run);
      els.who.textContent = run.user_name || "";
      s.spec = { folder: run.folder, dir: run.dir || "", job: run.job, args: runArgs(run), title: runTitle(run) };
      // A finished run stays live in memory for a while, for the windows
      // on it; opened afterwards it is history, not something to attach to.
      if (run.live && !ended(run.state)) {
        connect(s, run);
      } else {
        playback(s, run);
      }
    });
  }

  // ended reports whether a run's state is one it ended in: neither
  // queued, waiting for the run of its job before it, nor running.
  function ended(state) {
    return !!state && state !== "queued" && state !== "running";
  }

  // runTitle names a run as its window does: folder, the directory
  // inside it when the job is one of that directory's, and the job; a
  // job of the root, _, by its name alone, as Run.Name does.
  function runTitle(run) {
    return run.folder === "_" && !run.dir ? run.job : runPlace(run) + " / " + run.job;
  }

  // runPlace is the folder a run names: folder/dir, a directory of the
  // root without the _/, as Run.Place does.
  function runPlace(run) {
    if (run.folder === "_" && run.dir) return run.dir;
    return run.folder + (run.dir ? "/" + run.dir : "");
  }

  // runArgs reads the arguments a run row keeps, a JSON list.
  function runArgs(run) {
    try { return JSON.parse(run.args || "[]"); } catch (e) { return []; }
  }

  // connect attaches the terminal to a live run over the socket.
  function connect(s, run) {
    s.run = run;
    pageLink(run.id);
    els.title.textContent = runTitle(run);
    els.who.textContent = run.user_name || "";
    var term = s.term;
    var proto = location.protocol === "https:" ? "wss:" : "ws:";
    var ws = new WebSocket(proto + "//" + location.host + "/api/runs/" + encodeURIComponent(run.id) + "/attach");
    ws.binaryType = "arraybuffer";
    s.ws = ws;
    setState("running", "attaching");
    buttons({ stop: run.can_run, kill: run.can_run });
    els.hint.textContent = run.can_run
      ? "Closing this window doesn't stop the run. Stop ends it; Kill ends it right away."
      : "You're watching. You don't have permission to type into this run or end it. Closing this window doesn't stop it.";

    term.onData(function (data) {
      if (ws.readyState === 1) ws.send(JSON.stringify({ type: "input", message: data }));
    });
    s.onResize = function () {
      try { s.fit.fit(); } catch (e) { return; }
      if (ws.readyState === 1) ws.send(JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows }));
    };
    window.addEventListener("resize", s.onResize);

    ws.onopen = function () {
      setState("running");
      startClock(run.started_at ? Date.parse(run.started_at) : Date.now());
      if (run.can_run) ws.send(JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows }));
    };
    ws.onmessage = function (event) {
      if (typeof event.data !== "string") {
        term.write(new Uint8Array(event.data));
        return;
      }
      var msg;
      try { msg = JSON.parse(event.data); } catch (e) { return; }
      if (msg.type === "status") finish(s, msg);
    };
    ws.onclose = function () {
      if (session !== s || s.finished) return;
      s.finished = true;
      stopClock();
      setState("error", "detached");
      notice("The connection closed, but the run continues. Attach again to follow it.", "bad");
      buttons({ attach: true, again: run.can_run });
    };
  }

  // finish shows how a run ended.
  function finish(s, msg) {
    s.finished = true;
    s.changed = true;
    stopClock();
    s.term.write("\x1b[?25l");
    if (msg.state === "error") {
      setState("error");
      notice(msg.message, "bad");
      buttons({});
      return;
    }
    setState(msg.state, msg.state === "ok" ? "ok" : msg.state + (msg.exit_code ? " " + msg.exit_code : ""));
    if (msg.state === "killed") notice("Killed from another window, by Run force, or by the server.", "bad");
    else if (msg.state === "failed") notice(msg.message || ("The job exited with status " + msg.exit_code + "."), "bad");
    else if (msg.state === "stopped") notice("Stopped.", "");
    else notice("");
    els.hint.textContent = s.spec && s.spec.summarize
      ? "Finished. Claude writes a summary on the run's page, under the output."
      : "Finished. You can find the run and its recording in the history.";
    // A run that ended well is run again; one that did not is forced,
    // which also gets past what refused it.
    var can = s.run && s.run.can_run;
    buttons({ again: can && msg.state === "ok", force: can && msg.state !== "ok" });
  }

  // playback writes the recording of a finished run into the terminal.
  function playback(s, run) {
    s.run = run;
    pageLink(run.id);
    els.kind.textContent = "History";
    els.hint.textContent = "Recorded " + (run.started_at ? new Date(run.started_at).toLocaleString() : "") + ".";
    buttons({ again: run.can_run });
    fetch("/api/runs/" + encodeURIComponent(run.id) + "/cast", { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) throw new Error(res.status === 404 ? "no recording" : res.status === 410 ? "the recording expired under the Data retention settings" : "HTTP " + res.status);
        return res.text();
      })
      .then(function (text) {
        if (session !== s) return;
        var lines = text.split("\n").filter(Boolean);
        var out = [];
        for (var i = 1; i < lines.length; i++) {
          try {
            var ev = JSON.parse(lines[i]);
            if (ev[1] === "o") out.push(ev[2]);
          } catch (e) { /* torn last line */ }
        }
        s.term.write(out.join("") + "\x1b[?25l");
        finish(s, { state: run.state, exit_code: run.exit_code, message: run.status && run.status.message });
        s.changed = false;
        els.replay.hidden = false;
      })
      .catch(function (err) {
        if (session !== s) return;
        setState(run.state);
        notice("Couldn't load the recording: " + err.message, "bad");
      });
  }

  // replay plays the recording of a finished run into the window at
  // the pace it printed, idle gaps and all. It asks the server for the
  // recording and nothing else: no run, no socket, nothing in the
  // recents, and what is typed into the terminal goes nowhere. The
  // recording streams in, so a long one starts playing at once.
  function replay(id, title, who) {
    var s = open("Recording", title, who);
    s.replaying = true;
    dialog.classList.add("replaying");
    var term = s.term;
    term.options.disableStdin = true;
    term.options.cursorBlink = false;
    els.rec.hidden = false;
    pageLink(id);
    els.hint.textContent = "A recording, played as the run printed it. Nothing runs, and what you type goes nowhere.";
    buttons({ pause: true, restart: true });
    els.pause.textContent = "Pause";

    var size = null, events = [], at = 0, done = false, timer = null;
    var base = 0, pausedAt = 0, total = 0;
    var ctrl = window.AbortController ? new AbortController() : null;
    var clock = function () { return base ? (pausedAt || performance.now()) - base : 0; };
    var show = function () {
      var now = Math.min(clock(), done ? total : Infinity);
      els.elapsed.textContent = elapsedText(now) + (done ? " / " + elapsedText(total) : "");
    };
    var resize = function (cols, rows) {
      if (cols > 0 && rows > 0) term.write("", function () { term.resize(cols, rows); });
    };
    var step = function () {
      timer = null;
      if (session !== s || pausedAt || !base) return;
      var now = clock(), out = "";
      while (at < events.length && events[at][0] * 1000 <= now) {
        var ev = events[at++];
        if (ev[1] === "o") {
          out += ev[2];
        } else if (ev[1] === "r") {
          var wh = String(ev[2]).split("x");
          term.write(out);
          out = "";
          resize(parseInt(wh[0], 10), parseInt(wh[1], 10));
        }
      }
      if (out) term.write(out);
      if (at < events.length) {
        timer = setTimeout(step, events[at][0] * 1000 - now);
      } else if (done) {
        term.write("\x1b[?25l");
        pausedAt = base + total;
        show();
        notice("End of the recording.");
        buttons({ restart: true });
      }
    };
    var line = function (text) {
      if (!text) return;
      var v;
      try { v = JSON.parse(text); } catch (e) { return; /* torn last line */ }
      if (!size) {
        size = v;
        resize(v.width, v.height);
        base = performance.now();
        return;
      }
      if (Array.isArray(v) && typeof v[0] === "number") {
        events.push(v);
        total = Math.max(total, v[0] * 1000);
      }
    };
    s.stop = function () {
      if (timer) clearTimeout(timer);
      timer = null;
      if (ctrl) ctrl.abort();
    };
    s.pause = function () {
      if (done && at >= events.length) return;
      if (pausedAt) {
        base += performance.now() - pausedAt;
        pausedAt = 0;
        els.pause.textContent = "Pause";
        step();
      } else {
        pausedAt = performance.now();
        if (timer) clearTimeout(timer);
        timer = null;
        els.pause.textContent = "Resume";
      }
      show();
    };
    s.restart = function () {
      if (timer) clearTimeout(timer);
      timer = null;
      term.reset();
      term.options.disableStdin = true;
      if (size) resize(size.width, size.height);
      at = 0;
      pausedAt = 0;
      base = performance.now();
      notice("");
      buttons({ pause: true, restart: true });
      els.pause.textContent = "Pause";
      step();
    };
    s.timer = setInterval(show, 250);

    var buf = "";
    var feed = function (chunk, last) {
      var lines = (buf + chunk).split("\n");
      buf = last ? "" : lines.pop();
      lines.forEach(line);
      if (last) done = true;
      if (!timer) step();
      show();
    };
    fetch("/api/runs/" + encodeURIComponent(id) + "/cast", { credentials: "same-origin", signal: ctrl ? ctrl.signal : undefined })
      .then(function (res) {
        if (!res.ok) throw new Error(res.status === 404 ? "no recording" : res.status === 410 ? "the recording expired under the Data retention settings" : "HTTP " + res.status);
        if (!res.body || !res.body.getReader || !window.TextDecoder) return res.text().then(function (text) { feed(text, true); });
        var reader = res.body.getReader(), dec = new TextDecoder();
        var pump = function () {
          return reader.read().then(function (r) {
            if (session !== s) return reader.cancel();
            if (r.done) return feed(dec.decode(), true);
            feed(dec.decode(r.value, { stream: true }), false);
            return pump();
          });
        };
        return pump();
      })
      .catch(function (err) {
        if (session !== s) return;
        setState("error");
        notice("Couldn't load the recording: " + err.message, "bad");
        buttons({});
      });
  }

  function endRun(id, how) {
    return api("POST", "/api/runs/" + encodeURIComponent(id) + "/" + how).then(function (res) {
      if (res.status >= 300) window.alert(res.data.error || ("HTTP " + res.status));
      else if (!session && !follows) window.location.reload();
    });
  }

  if (dialog) {
    dialog.addEventListener("close", teardown);
    els.close.addEventListener("click", function () { dialog.close(); });
    els.stop.addEventListener("click", function () {
      if (session && session.ws && session.ws.readyState === 1) {
        session.ws.send(JSON.stringify({ type: "stop" }));
        els.stop.hidden = true;
      }
    });
    els.kill.addEventListener("click", function () {
      if (session && session.ws && session.ws.readyState === 1) {
        session.ws.send(JSON.stringify({ type: "kill" }));
        els.kill.hidden = true;
      }
    });
    els.attach.addEventListener("click", function () {
      var run = session && session.run;
      if (run) attach(run.id);
    });
    els.force.addEventListener("click", function () {
      var spec = session && session.spec;
      if (!spec) return;
      var reason = els.reasonText.value.trim();
      if (session.outOfHours && !reason) {
        notice("Enter a reason to run outside its hours.", "bad");
        els.reasonText.focus();
        return;
      }
      start(Object.assign({}, spec, { force: true, reason: reason }));
    });
    els.replay.addEventListener("click", function () {
      var run = session && session.run;
      if (run) replay(run.id, runTitle(run), run.user_name || "");
    });
    els.pause.addEventListener("click", function () { if (session && session.pause) session.pause(); });
    // The terminal keeps Esc for the job; a replay has none, so Esc
    // closes its window.
    dialog.addEventListener("keydown", function (event) {
      if (event.key === "Escape" && session && session.replaying) dialog.close();
    }, true);
    els.restart.addEventListener("click", function () { if (session && session.restart) session.restart(); });
    els.again.addEventListener("click", function () {
      var spec = session && session.spec;
      if (spec) start(Object.assign({}, spec, { force: false }));
    });
  }

  // --------------------------------------------------------------- confirm

  var confirmDialog = document.getElementById("confirm-dialog");
  var confirmResolve = null;

  // confirm asks in a dialog of our own, and resolves to the answer.
  function confirm(text, okLabel) {
    if (!confirmDialog || typeof confirmDialog.showModal !== "function") {
      return Promise.resolve(window.confirm(text));
    }
    document.getElementById("confirm-text").textContent = text;
    document.getElementById("confirm-ok").textContent = okLabel || "Delete";
    confirmDialog.showModal();
    return new Promise(function (resolve) { confirmResolve = resolve; });
  }

  if (confirmDialog) {
    document.getElementById("confirm-cancel").addEventListener("click", function () { confirmDialog.close(); });
    document.getElementById("confirm-ok").addEventListener("click", function () {
      var resolve = confirmResolve;
      confirmResolve = null;
      confirmDialog.close();
      if (resolve) resolve(true);
    });
    confirmDialog.addEventListener("close", function () {
      var resolve = confirmResolve;
      confirmResolve = null;
      if (resolve) resolve(false);
    });
  }

  // ----------------------------------------------------------------- forms

  // argsOf is the args a job's field sends: the line as typed, which
  // the API splits as a shell splits words, or for a job of a host,
  // @host:<host>, the line whole, the command ssh runs; null for none.
  function argsOf(line, job) {
    line = line.trim();
    if (!line) return null;
    return job.indexOf("@host:") === 0 ? [line] : line;
  }

  // serialize reads a form into the JSON the API takes. data-type says
  // what a field is: a list splits on commas, a bool parses, an int too,
  // and args go as argsOf has them for the form's job.
  function serialize(form) {
    var body = {}, joined = {};
    Array.prototype.forEach.call(form.elements, function (el) {
      if (!el.name || el.disabled) return;
      // A group of radios is the value of the checked one.
      if (el.type === "radio" && !el.checked) return;
      // A joined set is the checked values of the boxes that share a
      // name, as one comma separated value: a setting of a "map.key" form.
      if (el.dataset.type === "joined") {
        joined[el.name] = joined[el.name] || [];
        if (el.checked) joined[el.name].push(el.value);
        return;
      }
      // A set is the same as a list, empty when none is checked; under a
      // "map.key" name it is that key's list, left out while empty, so a
      // job without its own inherits the folder's.
      if (el.dataset.type === "set") {
        var at = el.name.indexOf(".");
        if (at > 0) {
          if (!el.checked) return;
          var group = el.name.slice(0, at), item = el.name.slice(at + 1);
          body[group] = body[group] || {};
          body[group][item] = body[group][item] || [];
          body[group][item].push(el.value);
          return;
        }
        body[el.name] = body[el.name] || [];
        if (el.checked) body[el.name].push(el.value);
        return;
      }
      var value = el.type === "checkbox" ? el.checked : el.value;
      switch (el.dataset.type) {
        case "list":
          value = value.split(",").map(function (s) { return s.trim(); }).filter(Boolean);
          break;
        case "bool":
          value = value === "true" || value === "1" || value === "on";
          break;
        case "int":
          value = parseInt(value, 10) || 0;
          break;
        case "args":
          value = argsOf(value, form.elements.job ? form.elements.job.value : "");
          if (value === null) return;
          break;
      }
      // "map.key" fields build an object; empty entries are left out, so
      // a job without an override inherits the folder's.
      var dot = el.name.indexOf(".");
      if (dot > 0) {
        var map = el.name.slice(0, dot), key = el.name.slice(dot + 1);
        if (value === "" || (Array.isArray(value) && !value.length)) return;
        body[map] = body[map] || {};
        body[map][key] = value;
        return;
      }
      body[el.name] = value;
    });
    Object.keys(joined).forEach(function (name) {
      var dot = name.indexOf("."), value = joined[name].join(", ");
      if (dot < 0) { body[name] = value; return; }
      if (!value) return;
      var map = name.slice(0, dot);
      body[map] = body[map] || {};
      body[map][name.slice(dot + 1)] = value;
    });
    return body;
  }

  // A user's folders: All projects turns the boxes off, so the form sends
  // none and the user sees everything; Custom turns them on.
  document.addEventListener("change", function (event) {
    var radio = event.target.closest("[data-folder-pick] input[name=scope]");
    if (!radio) return;
    var custom = radio.value === "custom" && radio.checked;
    radio.closest("[data-folder-pick]").querySelectorAll("input[name=folders]").forEach(function (box) { box.disabled = !custom; });
  });

  // A project's Workspace select moves it: picking another workspace
  // makes the save ask first, naming the new ID and the rows that stay
  // behind under the old one.
  document.addEventListener("change", function (event) {
    var pick = event.target.closest("select[data-move]");
    if (!pick) return;
    var form = pick.form, id = pick.dataset.move;
    if (pick.value === pick.dataset.moveFrom) {
      delete form.dataset.confirm;
      delete form.dataset.confirmOk;
      form.dataset.next = "/settings/projects?notice=Saved";
      return;
    }
    var last = id.slice(id.lastIndexOf("/") + 1);
    var to = pick.value === "_" ? last : pick.value + "/" + last;
    var left = pick.dataset.moveLeft;
    form.dataset.confirm = "Move " + id + " to " + to + "? The checkout, the project and the users' folder access move with it; earlier runs stay under " + id + ". " +
      (left ? "These still point to " + id + " and won't move: " + left + "." : "Nothing else points to " + id + ".");
    form.dataset.confirmOk = "Move";
    form.dataset.next = "/settings/projects?notice=" + encodeURIComponent("Moved to " + to);
  });

  // The algorithm of a new key carries its size: the bits field takes
  // the chosen option's, 0 for ed25519.
  document.addEventListener("change", function (event) {
    var pick = event.target.closest("select[data-key-algo]");
    if (pick) pick.form.elements.bits.value = pick.selectedOptions[0].dataset.bits || "0";
  });

  // A show/hide switch remembers its choice, in this browser.
  Array.prototype.forEach.call(document.querySelectorAll("input[data-peek]"), function (input) {
    var key = "task-ui.peek." + input.dataset.peek;
    try {
      var saved = localStorage.getItem(key);
      if (saved !== null) input.checked = saved === "1";
    } catch (e) { /* no storage: the default stands */ }
    input.addEventListener("change", function () {
      try { localStorage.setItem(key, input.checked ? "1" : "0"); } catch (e) { /* not remembered */ }
    });
  });

  function formError(form, text) {
    var box = form.querySelector(".form-error");
    if (!box) {
      box = document.createElement("div");
      box.className = "form-error";
      box.setAttribute("role", "alert");
      form.insertBefore(box, form.firstChild);
    }
    box.textContent = text || "";
  }

  function next(target) {
    if (!target || target === "reload") window.location.reload();
    else window.location.href = target;
  }

  // filled puts the answer's fields into a data-next: {rekeyed} becomes
  // the count of secrets a rotation wrote again.
  function filled(target, data) {
    if (!target) return target;
    return target.replace(/\{(\w+)\}/g, function (all, name) {
      return data && data[name] !== undefined ? encodeURIComponent(data[name]) : "";
    });
  }

  // reveal shows what an answer holds once, such as a new token, in the
  // block a form's data-reveal names, and hides what it replaces: the
  // block's section's [data-before-reveal]. Each [data-fill] there takes
  // its template with the answer's fields put in; a copy button takes it
  // as what it copies. Nothing keeps the answer: a reload shows it no more.
  function reveal(selector, data) {
    var target = document.querySelector(selector);
    if (!target) return;
    var section = target.closest("section") || target.parentNode;
    section.querySelectorAll("[data-before-reveal]").forEach(function (el) { el.hidden = true; });
    target.querySelectorAll("[data-fill]").forEach(function (el) {
      var value = el.dataset.fill.replace(/\{(\w+)\}/g, function (all, name) {
        var v = data && data[name] !== undefined && data[name] !== null ? String(data[name]) : "";
        if (/_at$/.test(name) && v) v = new Date(v).toLocaleString();
        return v;
      });
      if (el.hasAttribute("data-copy")) el.setAttribute("data-copy", value);
      else el.textContent = value;
    });
    target.hidden = false;
  }

  function submit(form) {
    var method = (form.dataset.method || "POST").toUpperCase();
    var button = form.querySelector("[type=submit]");
    if (button) button.disabled = true;
    formError(form, "");
    var body = serialize(form);
    if (form.hasAttribute("data-workflow")) body.steps = workflowSteps(form);
    api(method, form.dataset.api, body).then(function (res) {
      if (res.status >= 300) {
        if (button) button.disabled = false;
        formError(form, (res.data.error || ("HTTP " + res.status)) + (res.data.output ? "\n\n" + res.data.output : ""));
        return;
      }
      // A prompt is a note: the form empties and its thread shows at the
      // top of the list under it, where the answer lands.
      if (form.hasAttribute("data-refresh-notes")) {
        if (button) button.disabled = false;
        form.reset();
        refreshNotes();
        return;
      }
      if (form.dataset.reveal) {
        reveal(form.dataset.reveal, res.data);
        return;
      }
      next(filled(form.dataset.next, res.data));
    }).catch(function (err) {
      if (button) button.disabled = false;
      formError(form, err.message);
    });
  }

  // Data retention: a preset fills the size field, and the estimate
  // under it follows the field, at the average size of what the kind
  // holds.
  function parseSize(text) {
    var m = /^\s*([\d.]+)\s*([kmgt]?)i?b?\s*$/i.exec(text || "");
    if (!m) return 0;
    var unit = { "": 1, k: 1024, m: 1048576, g: 1073741824, t: 1099511627776 }[m[2].toLowerCase()];
    return parseFloat(m[1]) * unit;
  }
  function estimate(block) {
    var input = block.querySelector("[data-size-input]"), out = block.querySelector("[data-estimate]");
    var average = parseFloat(block.dataset.average) || 0, limit = parseSize(input.value);
    if (!average || !limit) { out.textContent = ""; return; }
    var n = Math.floor(limit / average);
    out.textContent = "about " + n + " " + (n === 1 ? block.dataset.one : block.dataset.many);
  }
  document.addEventListener("click", function (event) {
    var btn = event.target.closest("[data-retention] [data-size]");
    if (!btn) return;
    var block = btn.closest("[data-retention]");
    block.querySelector("[data-size-input]").value = btn.dataset.size;
    estimate(block);
  });
  document.addEventListener("input", function (event) {
    if (!event.target.matches("[data-retention] [data-size-input]")) return;
    estimate(event.target.closest("[data-retention]"));
  });

  document.addEventListener("submit", function (event) {
    var form = event.target.closest("form[data-api]");
    if (!form) return;
    event.preventDefault();
    if (form.dataset.confirm) {
      confirm(form.dataset.confirm, form.dataset.confirmOk).then(function (ok) { if (ok) submit(form); });
      return;
    }
    submit(form);
  });

  // A workspace's Export posts the passphrase, which keeps it out of the
  // URL, and saves the bundle the answer carries under its file name.
  document.addEventListener("submit", function (event) {
    var form = event.target.closest("form[data-export]");
    if (!form) return;
    event.preventDefault();
    var button = form.querySelector("[type=submit]");
    button.disabled = true;
    formError(form, "");
    fetch(form.dataset.export, {
      method: "POST", credentials: "same-origin",
      headers: { "Content-Type": "application/json", Accept: "application/octet-stream" },
      body: JSON.stringify({ passphrase: form.elements.passphrase.value })
    }).then(function (res) {
      if (!res.ok) {
        return res.text().then(function (text) {
          var data = {};
          try { data = JSON.parse(text); } catch (e) { data = { error: text }; }
          throw new Error(data.error || ("HTTP " + res.status));
        });
      }
      var name = (/filename="?([^";]+)"?/.exec(res.headers.get("Content-Disposition") || "") || [])[1] || "workspace.taskui-ws";
      return res.blob().then(function (blob) {
        var link = document.createElement("a");
        link.href = URL.createObjectURL(blob);
        link.download = name;
        document.body.appendChild(link);
        link.click();
        link.remove();
        setTimeout(function () { URL.revokeObjectURL(link.href); }, 1000);
        form.reset();
      });
    }).catch(function (err) {
      formError(form, err.message);
    }).then(function () {
      button.disabled = false;
    });
  });

  // Import sends the bundle as a multipart form. A workspace whose every
  // project cloned opens its configure page; otherwise the form names
  // the projects that did not clone, with git's output, and the users
  // its grants name that this instance does not have.
  document.addEventListener("submit", function (event) {
    var form = event.target.closest("form[data-import]");
    if (!form) return;
    event.preventDefault();
    var button = form.querySelector("[type=submit]");
    button.disabled = true;
    formError(form, "");
    fetch(form.dataset.import, {
      method: "POST", credentials: "same-origin", headers: { Accept: "application/json" }, body: new FormData(form)
    }).then(function (res) {
      return res.text().then(function (text) {
        var data = {};
        try { data = text ? JSON.parse(text) : {}; } catch (e) { data = { error: text }; }
        return { status: res.status, data: data };
      });
    }).then(function (res) {
      button.disabled = false;
      if (res.status >= 300) {
        formError(form, (res.data.error || ("HTTP " + res.status)) + (res.data.output ? "\n\n" + res.data.output : ""));
        return;
      }
      var folder = res.data.workspace.folder;
      var failed = (res.data.projects || []).filter(function (p) { return !p.cloned && p.output; });
      var unknown = res.data.unknown_users || [];
      if (!failed.length && !unknown.length) {
        next("/settings/workspaces/" + encodeURIComponent(folder) + "?notice=Imported");
        return;
      }
      var lines = ["Imported " + folder + "."];
      failed.forEach(function (p) { lines.push(p.folder + " didn't clone; it stays listed:\n" + p.output); });
      if (unknown.length) lines.push("Some grants name users this instance doesn't have: " + unknown.join(", ") + ".");
      form.reset();
      formError(form, lines.join("\n\n"));
    }).catch(function (err) {
      button.disabled = false;
      formError(form, err.message);
    });
  });

  // keepFolder keeps the options of a folder in a job select, each in
  // the group it came in, such as Hosts, and drops the rest.
  function keepFolder(select, folder) {
    if (!select.allOptions) {
      select.allOptions = Array.prototype.map.call(select.options, function (option) {
        return { option: option, group: option.parentNode.tagName === "OPTGROUP" ? option.parentNode.label : "" };
      });
    }
    select.innerHTML = "";
    var groups = {};
    select.allOptions.forEach(function (item) {
      if (item.option.dataset.folder !== folder) return;
      if (!item.group) {
        select.appendChild(item.option);
        return;
      }
      if (!groups[item.group]) {
        groups[item.group] = document.createElement("optgroup");
        groups[item.group].label = item.group;
        select.appendChild(groups[item.group]);
      }
      groups[item.group].appendChild(item.option);
    });
  }

  // A select that narrows another keeps only the options of its folder:
  // the job list follows the folder picked for a schedule.
  function narrow(select) {
    var target = document.getElementById(select.dataset.narrows);
    if (!target) return;
    keepFolder(target, select.value);
    if (target.dropdown) target.dropdown();
  }
  document.querySelectorAll("select[data-narrows]").forEach(function (select) {
    narrow(select);
    select.addEventListener("change", function () { narrow(select); });
  });

  // A switch turns a part of a form on or off: a workflow's prompt, or
  // one of its steps. Off collapses the part to its switch and heading,
  // and its fields leave the form, so a prompt that is off saves as none;
  // the text stays, for the switch to bring back.
  document.addEventListener("change", function (event) {
    if (!event.target.matches("[data-switch]")) return;
    var part = event.target.closest("[data-switchable]");
    if (!part) return;
    var on = event.target.checked;
    part.classList.toggle("off", !on);
    part.querySelectorAll(".switch-body textarea, .switch-body select, .switch-body input").forEach(function (field) {
      field.disabled = !on;
    });
    if (on) {
      var first = part.querySelector(".switch-body textarea");
      if (first) first.focus();
    }
  });

  // A select with data-submit saves its form on a change: who a
  // schedule runs as.
  document.querySelectorAll("select[data-submit]").forEach(function (select) {
    select.addEventListener("change", function () { select.form.requestSubmit(); });
  });

  // ------------------------------------------------------------ cc-menus

  // A control centre's Other menu opens over the page, placed against
  // its button, so the table's scrolling box cannot clip it; under the
  // button, or over it when the screen ends first. It closes on a click
  // outside it, on Escape, on a scroll, and once one of its jobs is
  // picked.
  document.addEventListener("toggle", function (event) {
    var menu = event.target;
    if (!menu.matches || !menu.matches("details.cc-more") || !menu.open) return;
    var list = menu.querySelector(".cc-menu");
    var at = menu.querySelector("summary").getBoundingClientRect();
    list.style.left = Math.max(8, at.right - list.offsetWidth) + "px";
    var below = window.innerHeight - at.bottom;
    list.style.top = (below < list.offsetHeight + 8 && at.top > below ? at.top - list.offsetHeight - 4 : at.bottom + 4) + "px";
  }, true);
  window.addEventListener("scroll", function () {
    document.querySelectorAll("details.cc-more[open]").forEach(function (menu) { menu.removeAttribute("open"); });
  }, true);
  document.addEventListener("click", function (event) {
    document.querySelectorAll("details.cc-more[open]").forEach(function (menu) {
      if (!menu.contains(event.target) || event.target.closest(".cc-menu .cc, .cc-menu [data-run], .cc-menu [data-attach]")) menu.removeAttribute("open");
    });
  });
  // A row of a Run menu's table runs on Enter or Space, as a button does.
  document.addEventListener("keydown", function (event) {
    if (event.key !== "Enter" && event.key !== " ") return;
    var row = event.target.closest && event.target.closest('.cc-menu tr[role="menuitem"]');
    if (!row) return;
    event.preventDefault();
    row.click();
  });
  document.addEventListener("keydown", function (event) {
    if (event.key !== "Escape") return;
    document.querySelectorAll("details.cc-more[open]").forEach(function (menu) {
      menu.removeAttribute("open");
      menu.querySelector("summary").focus();
    });
  });

  // A ⋯ menu closes on a click outside it, on Escape, and once one of
  // its items is picked, before the item's confirm opens.
  document.addEventListener("click", function (event) {
    document.querySelectorAll("details.more-menu[open]").forEach(function (menu) {
      if (!menu.contains(event.target) || event.target.closest(".more-item")) menu.removeAttribute("open");
    });
  });
  document.addEventListener("keydown", function (event) {
    if (event.key !== "Escape") return;
    document.querySelectorAll("details.more-menu[open]").forEach(function (menu) {
      menu.removeAttribute("open");
      menu.querySelector("summary").focus();
    });
  });

  // A permission matrix: a cell cycles inherit, allow (✓) and deny (✕) on
  // a click. Its hidden field turns on while it differs from what the
  // page loaded with, so Save puts only the cells that changed.
  document.addEventListener("click", function (event) {
    var cell = event.target.closest("form[data-matrix] button[data-tri]");
    if (!cell) return;
    var next = { "": "allow", allow: "deny", deny: "" }[cell.dataset.tri || ""];
    var field = cell.parentNode.querySelector("input[type=hidden]");
    var label = cell.dataset.title + ": " + (next || "inherit");
    cell.dataset.tri = next;
    cell.textContent = next === "allow" ? "\u2713" : next === "deny" ? "\u2715" : "";
    cell.title = label;
    cell.setAttribute("aria-label", label);
    field.value = next;
    field.disabled = next === field.dataset.was;
    var form = cell.closest("form"), changed = form.querySelectorAll("input[type=hidden]:not([disabled])").length;
    var note = form.querySelector("[data-matrix-changed]");
    if (note) note.textContent = changed ? changed + (changed === 1 ? " change" : " changes") + " to save." : "No changes yet.";
  });

  // ------------------------------------------------------------- dropdowns

  // Every select gets a dropdown drawn in the page's own style: a button
  // showing the choice, and a list under it. The select stays in the
  // form, hidden, so forms, validation and change handlers read it as
  // before; the dropdown follows its options, its value and whether it is
  // hidden, as scripts change them. data-native keeps a select as it is.
  var openDropdown = null;

  function enhanceSelect(select) {
    if (select.multiple || select.hasAttribute("data-native") || select.closest(".dd")) return;
    var wrap = document.createElement("div");
    wrap.className = "dd";
    select.parentNode.insertBefore(wrap, select);
    wrap.appendChild(select);
    select.tabIndex = -1;
    select.setAttribute("aria-hidden", "true");

    var button = document.createElement("button");
    button.type = "button";
    button.className = "dd-button";
    button.setAttribute("aria-haspopup", "listbox");
    button.setAttribute("aria-expanded", "false");
    var label = select.getAttribute("aria-label") || (select.id && document.querySelector('label[for="' + select.id + '"]') || {}).textContent;
    if (label) button.setAttribute("aria-label", label.trim());
    var text = document.createElement("span");
    text.className = "dd-text";
    button.appendChild(text);
    var chev = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    chev.setAttribute("class", "ic dd-chev");
    var use = document.createElementNS("http://www.w3.org/2000/svg", "use");
    use.setAttribute("href", "#i-chev");
    chev.appendChild(use);
    button.appendChild(chev);
    wrap.appendChild(button);

    var list = document.createElement("div");
    list.className = "dd-list";
    list.setAttribute("role", "listbox");
    list.hidden = true;
    wrap.appendChild(list);

    var active = -1;
    var typed = "", typedAt = 0;

    function render() {
      var picked = select.options[select.selectedIndex];
      text.textContent = picked ? picked.textContent : "";
      wrap.hidden = select.hidden;
      button.disabled = select.disabled;
      if (!list.hidden) fill();
    }
    function fill() {
      list.innerHTML = "";
      Array.prototype.forEach.call(select.children, function (child) {
        if (child.tagName === "OPTGROUP") {
          var head = document.createElement("div");
          head.className = "dd-group";
          head.textContent = child.label;
          list.appendChild(head);
          Array.prototype.forEach.call(child.children, addOption);
        } else {
          addOption(child);
        }
      });
      mark();
    }
    function addOption(option) {
      var item = document.createElement("div");
      item.className = "dd-option";
      item.setAttribute("role", "option");
      item.dataset.index = option.index;
      item.textContent = option.textContent;
      if (option.disabled) item.setAttribute("aria-disabled", "true");
      list.appendChild(item);
    }
    function items() {
      return Array.prototype.slice.call(list.querySelectorAll(".dd-option:not([aria-disabled])"));
    }
    function mark() {
      Array.prototype.forEach.call(list.querySelectorAll(".dd-option"), function (item) {
        var index = Number(item.dataset.index);
        item.setAttribute("aria-selected", index === select.selectedIndex ? "true" : "false");
        item.classList.toggle("active", index === active);
        if (index === active) item.scrollIntoView({ block: "nearest" });
      });
    }
    function open() {
      if (openDropdown && openDropdown !== close) openDropdown();
      fill();
      active = select.selectedIndex;
      list.hidden = false;
      wrap.classList.add("open");
      button.setAttribute("aria-expanded", "true");
      // Opens upward when the space under it is short of the list.
      var rect = button.getBoundingClientRect();
      wrap.classList.toggle("up", window.innerHeight - rect.bottom < Math.min(list.scrollHeight, 288) + 8 && rect.top > window.innerHeight - rect.bottom);
      mark();
      openDropdown = close;
    }
    function close() {
      list.hidden = true;
      wrap.classList.remove("open", "up");
      button.setAttribute("aria-expanded", "false");
      if (openDropdown === close) openDropdown = null;
    }
    function choose(index) {
      if (index < 0) return;
      var changed = select.selectedIndex !== index;
      select.selectedIndex = index;
      close();
      render();
      button.focus();
      if (changed) {
        select.dispatchEvent(new Event("input", { bubbles: true }));
        select.dispatchEvent(new Event("change", { bubbles: true }));
      }
    }
    function move(step) {
      var all = items();
      if (!all.length) return;
      var at = all.findIndex(function (item) { return Number(item.dataset.index) === active; });
      at = at < 0 ? 0 : Math.max(0, Math.min(all.length - 1, at + step));
      active = Number(all[at].dataset.index);
      mark();
    }

    button.addEventListener("click", function () {
      if (list.hidden) open(); else close();
    });
    list.addEventListener("mousedown", function (event) { event.preventDefault(); });
    list.addEventListener("click", function (event) {
      var item = event.target.closest(".dd-option");
      if (item && !item.hasAttribute("aria-disabled")) choose(Number(item.dataset.index));
    });
    list.addEventListener("mousemove", function (event) {
      var item = event.target.closest(".dd-option");
      if (item && Number(item.dataset.index) !== active) { active = Number(item.dataset.index); mark(); }
    });
    button.addEventListener("keydown", function (event) {
      var key = event.key;
      if (key === "ArrowDown" || key === "ArrowUp") {
        event.preventDefault();
        if (list.hidden) { open(); return; }
        move(key === "ArrowDown" ? 1 : -1);
      } else if (key === "Home" || key === "End") {
        if (list.hidden) return;
        event.preventDefault();
        var all = items();
        if (all.length) { active = Number(all[key === "Home" ? 0 : all.length - 1].dataset.index); mark(); }
      } else if (key === "Enter" || key === " ") {
        if (list.hidden) return;
        event.preventDefault();
        choose(active);
      } else if (key === "Escape") {
        if (list.hidden) return;
        event.preventDefault();
        event.stopPropagation();
        close();
      } else if (key === "Tab") {
        close();
      } else if (key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
        // Typing jumps to the first option that starts with what was typed.
        var now = Date.now();
        typed = (now - typedAt > 700 ? "" : typed) + key.toLowerCase();
        typedAt = now;
        var hit = Array.prototype.find.call(select.options, function (o) { return !o.disabled && o.textContent.trim().toLowerCase().indexOf(typed) === 0; });
        if (!hit) return;
        if (list.hidden) choose(hit.index);
        else { active = hit.index; mark(); }
      }
    });
    // A label pointing at the select, or a script focusing it, focuses
    // the dropdown; the form's own reset puts the shown choice back.
    select.addEventListener("focus", function () { button.focus(); });
    select.addEventListener("change", render);
    if (select.form) select.form.addEventListener("reset", function () { setTimeout(render, 0); });
    new MutationObserver(render).observe(select, { childList: true, subtree: true, characterData: true, attributes: true, attributeFilter: ["hidden", "disabled"] });
    select.dropdown = render;
    render();
  }

  document.addEventListener("mousedown", function (event) {
    if (openDropdown && !event.target.closest(".dd.open")) openDropdown();
  });
  document.querySelectorAll("select").forEach(enhanceSelect);
  // Selects that arrive later, a cloned form row or a window's content,
  // get theirs on arrival.
  new MutationObserver(function (records) {
    records.forEach(function (record) {
      record.addedNodes.forEach(function (node) {
        if (node.nodeType !== 1) return;
        if (node.tagName === "SELECT") enhanceSelect(node);
        else node.querySelectorAll && node.querySelectorAll("select").forEach(enhanceSelect);
      });
    });
  }).observe(document.body, { childList: true, subtree: true });

  // ------------------------------------------------------------- workflows

  // A step's folder select also lists the workflows, as workflow:<id>.
  var WORKFLOW = "workflow:";

  // workflowSteps reads a workflow form's steps, in their order: a job of
  // a folder, or a workflow, each enabled or not.
  function workflowSteps(form) {
    return Array.prototype.map.call(form.querySelectorAll("[data-step]"), function (row) {
      var folder = row.querySelector("[data-step-folder]").value;
      var enabled = row.querySelector("[data-switch]").checked;
      if (folder.indexOf(WORKFLOW) === 0) return { workflow: folder.slice(WORKFLOW.length), enabled: enabled };
      var step = { folder: folder, job: row.querySelector("[data-step-job]").value, enabled: enabled };
      var args = argsOf(row.querySelector("[data-step-args]").value, step.job);
      if (args !== null) step.args = args;
      return step;
    });
  }

  // narrowStep keeps the jobs of the step's folder in its job select,
  // and the job picked when it still belongs. A step that runs a
  // workflow has no job to pick.
  function narrowStep(row) {
    var folder = row.querySelector("[data-step-folder]");
    var job = row.querySelector("[data-step-job]");
    var flow = folder.value.indexOf(WORKFLOW) === 0;
    job.hidden = flow;
    job.required = !flow;
    row.querySelector("[data-step-args]").hidden = flow;
    if (flow) return;
    var picked = job.value;
    keepFolder(job, folder.value);
    if (picked && Array.prototype.some.call(job.options, function (o) { return o.value === picked; })) job.value = picked;
    if (job.dropdown) job.dropdown();
  }

  // stepHead names a step by what it runs, the heading it keeps when
  // it is off: folder / job, or the workflow.
  function stepHead(row) {
    var folder = row.querySelector("[data-step-folder]");
    var job = row.querySelector("[data-step-job]");
    var head = row.querySelector("[data-step-head]");
    var picked = folder.options[folder.selectedIndex];
    if (!picked) return;
    head.textContent = folder.value.indexOf(WORKFLOW) === 0 ? picked.text : folder.value + " / " + job.value;
  }

  // stepTools enables the move buttons that have somewhere to go, and
  // keeps one step at least.
  function stepTools(list) {
    var rows = list.querySelectorAll("[data-step]");
    Array.prototype.forEach.call(rows, function (row, i) {
      row.querySelector("[data-step-up]").disabled = i === 0;
      row.querySelector("[data-step-down]").disabled = i === rows.length - 1;
    });
  }

  document.querySelectorAll("form[data-workflow]").forEach(function (form) {
    var list = form.querySelector("[data-steps]");
    list.querySelectorAll("[data-step]").forEach(narrowStep);
    stepTools(list);
    form.addEventListener("change", function (event) {
      if (event.target.matches("[data-step-folder], [data-step-job]")) {
        var step = event.target.closest("[data-step]");
        if (event.target.matches("[data-step-folder]")) narrowStep(step);
        stepHead(step);
      }
    });
    form.addEventListener("click", function (event) {
      // Summarize fills the prompt with the short summary instruction.
      var preset = event.target.closest("[data-workflow-preset]");
      if (preset) {
        event.preventDefault();
        var text = form.querySelector("[data-workflow-prompt-text]");
        text.value = preset.dataset.workflowPreset;
        text.focus();
        return;
      }
      var btn = event.target.closest("[data-step-add], [data-step-up], [data-step-down], [data-step-remove]");
      if (!btn) return;
      event.preventDefault();
      var row = btn.closest("[data-step]");
      if (btn.hasAttribute("data-step-add")) {
        var fresh = form.querySelector("[data-step-template]").content.firstElementChild.cloneNode(true);
        var last = list.querySelector("[data-step]:last-child");
        // A new step starts in the folder of the one before it.
        if (last) fresh.querySelector("[data-step-folder]").value = last.querySelector("[data-step-folder]").value;
        list.appendChild(fresh);
        fresh.querySelectorAll("select").forEach(enhanceSelect);
        narrowStep(fresh);
        stepHead(fresh);
        fresh.querySelector("[data-step-job]").focus();
      } else if (btn.hasAttribute("data-step-up") && row.previousElementSibling) {
        list.insertBefore(row, row.previousElementSibling);
      } else if (btn.hasAttribute("data-step-down") && row.nextElementSibling) {
        list.insertBefore(row.nextElementSibling, row);
      } else if (btn.hasAttribute("data-step-remove")) {
        // The last step goes too: a workflow of a prompt alone has none.
        row.remove();
      }
      stepTools(list);
    });
  });

  // Move up and Move down on the Workspaces page: the workspace trades
  // places with its neighbour, and every workspace whose sort is not its
  // place in the new order gets it, which is the two of them once the
  // sorts are 0, 1, 2 and on.
  document.addEventListener("click", function (event) {
    var btn = event.target.closest("[data-workspace-move]");
    if (!btn) return;
    event.preventDefault();
    btn.disabled = true;
    api("GET", "/api/workspaces").then(function (res) {
      if (res.status >= 300) throw new Error(res.data.error || ("HTTP " + res.status));
      // The default workspace keeps its place, first; the rest swap.
      var list = res.data.workspaces.filter(function (ws) { return ws.folder !== "_"; });
      var i = list.findIndex(function (ws) { return ws.folder === btn.dataset.folder; });
      var j = btn.dataset.workspaceMove === "up" ? i - 1 : i + 1;
      if (i < 0 || j < 0 || j >= list.length) return null;
      var moved = list[i];
      list[i] = list[j];
      list[j] = moved;
      return Promise.all(list.map(function (ws, sort) {
        if (ws.sort === sort) return null;
        return api("PUT", "/api/workspaces/" + encodeURIComponent(ws.folder), { title: ws.title, description: ws.description, sort: sort });
      }));
    }).then(function (results) {
      var failed = (results || []).find(function (res) { return res && res.status >= 300; });
      if (failed) throw new Error(failed.data.error || ("HTTP " + failed.status));
      next("reload");
    }).catch(function (err) {
      btn.disabled = false;
      window.alert(err.message);
    });
  });

  // The workflow page follows its topic; live says so, below. Any page
  // that follows a topic draws its changes and is never reloaded.
  var live = document.querySelector('[data-sse^="workflow:"]');
  var follows = document.querySelector("[data-sse]");

  // workflowInput asks for the input of a workflow with a prompt, in a
  // dialog: resolves to the text, empty for none, or null on Cancel.
  var inputDialog = document.getElementById("workflow-input-dialog");
  var inputResolve = null;
  function workflowInput(name) {
    if (!inputDialog || typeof inputDialog.showModal !== "function") return Promise.resolve("");
    var form = inputDialog.querySelector("form");
    form.reset();
    document.getElementById("workflow-input-title").textContent = "Run " + name;
    inputDialog.showModal();
    form.elements.input.focus();
    return new Promise(function (resolve) { inputResolve = resolve; });
  }
  if (inputDialog) {
    var inputForm = inputDialog.querySelector("form");
    var settle = function (value) {
      var resolve = inputResolve;
      inputResolve = null;
      if (resolve) resolve(value);
    };
    inputForm.addEventListener("submit", function () { settle(inputForm.elements.input.value.trim()); });
    inputForm.querySelector("[data-workflow-input-cancel]").addEventListener("click", function () { inputDialog.close(); });
    inputForm.elements.input.addEventListener("keydown", function (event) {
      if (event.key === "Enter" && (event.ctrlKey || event.metaKey)) {
        event.preventDefault();
        inputForm.requestSubmit();
      }
    });
    inputDialog.addEventListener("close", function () { settle(null); });
  }

  document.addEventListener("click", function (event) {
    var btn = event.target.closest("[data-workflow-run], [data-workflow-stop]");
    if (!btn) return;
    event.preventDefault();
    var run = btn.hasAttribute("data-workflow-run");
    // A workflow with a prompt asks for its input first.
    var asked = run && btn.hasAttribute("data-workflow-prompt") ? workflowInput(btn.dataset.workflowPrompt) : Promise.resolve("");
    asked.then(function (input) {
      if (input === null) return;
      startOrStop(btn, run, input);
    });
  });

  function startOrStop(btn, run, input) {
    btn.disabled = true;
    var path = run
      ? "/api/workflows/" + encodeURIComponent(btn.dataset.workflowRun) + "/runs"
      : "/api/workflow-runs/" + encodeURIComponent(btn.dataset.workflowStop) + "/stop";
    api("POST", path, input ? { input: input } : undefined).then(function (res) {
      // A workflow already running answers 409; its page shows that run.
      if (res.status >= 300 && res.status !== 409) {
        btn.disabled = false;
        window.alert(res.data.error || ("HTTP " + res.status));
        return;
      }
      // The workflow's own page hears of the run from its events and
      // draws it; elsewhere, go there.
      if (live) return;
      if (run) window.location.href = "/workflows/" + encodeURIComponent(btn.dataset.workflowRun);
      else window.location.reload();
    });
  }

  // The workflow page is drawn on from the events of its topic, no
  // fetch and no reload: `run` resets the pipeline to pending, swaps Run
  // for Stop and adds the run on top of the runs; `step` sets a step's
  // state and links it to its run; `done` ends the run.
  function el(tag, className, text) {
    var node = document.createElement(tag);
    if (className) node.className = className;
    if (text) node.textContent = text;
    return node;
  }

  function icon(name) {
    var svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    svg.setAttribute("class", "ic");
    var use = document.createElementNS("http://www.w3.org/2000/svg", "use");
    use.setAttribute("href", "#" + name);
    svg.appendChild(use);
    return svg;
  }

  function stateSpan(state, message) {
    var span = el("span", "state");
    span.appendChild(el("i", "dot " + state));
    span.appendChild(document.createTextNode(state));
    if (message) {
      span.appendChild(document.createTextNode(" "));
      span.appendChild(el("span", "faint", message));
    }
    return span;
  }

  // pagedBack reports whether the page shows older runs, through a
  // cursor in its query: a run that starts belongs on the first page.
  function pagedBack() {
    var q = new URLSearchParams(location.search);
    return q.has("before") || q.has("after");
  }

  function workflowEvent(name, data) {
    var nodes = Array.prototype.slice.call(document.querySelectorAll("[data-pipeline] > [data-step]"));
    var latest = document.querySelector("[data-latest]");
    var runs = document.querySelector("[data-runs]");
    var actions = document.querySelector("[data-actions]");
    var row = runs && runs.querySelector('[data-run-id="' + data.run_id + '"]');

    if (name === "run") {
      nodes.forEach(function (node) {
        setStep(node, "pending", null, null);
        // A step that ran a workflow starts over: its chain comes with the run.
        var sub = node.querySelector(".pipeline.sub");
        if (sub) sub.innerHTML = "";
      });
      if (latest) {
        latest.innerHTML = "";
        latest.appendChild(document.createTextNode("Latest run "));
        latest.appendChild(stateSpan("running"));
        latest.appendChild(document.createTextNode(" just now by " + (data.user_name || "")));
      }
      // A page further down the runs leaves the new run to the first page.
      if (runs && !row && !pagedBack()) {
        var empty = runs.querySelector(".empty");
        if (empty) empty.remove();
        row = el("tr");
        row.dataset.runId = data.run_id;
        var started = el("td", "nowrap", "just now");
        started.dataset.label = "Started";
        var by = el("td");
        if (data.user_name) by.innerHTML = userHTML(data.user_name);
        by.dataset.label = "By";
        var result = el("td");
        result.dataset.label = "Result";
        result.dataset.result = "";
        result.appendChild(stateSpan("running"));
        var steps = el("td");
        steps.dataset.label = "Steps";
        var chain = el("span", "chain");
        nodes.forEach(function (node, i) {
          if (i) chain.appendChild(el("span", "faint then"));
          var link = node.querySelector("a");
          var a = el("a", "nowrap");
          a.href = link.dataset.target || link.getAttribute("href");
          a.dataset.pos = i;
          a.appendChild(el("i", "dot pending"));
          if (link.querySelector(".ic")) a.appendChild(icon("i-flow"));
          a.appendChild(document.createTextNode(link.textContent.trim()));
          chain.appendChild(a);
        });
        steps.appendChild(chain);
        var view = el("td", "actions");
        var viewLink = el("a", "btn btn-sm", "View");
        viewLink.href = location.pathname.replace(/\/$/, "") + "/runs/" + encodeURIComponent(data.run_id);
        viewLink.dataset.workflowView = "";
        view.appendChild(viewLink);
        row.appendChild(started);
        row.appendChild(by);
        row.appendChild(result);
        row.appendChild(steps);
        row.appendChild(view);
        runs.insertBefore(row, runs.firstChild);
      }
      if (actions) {
        var runBtn = actions.querySelector("[data-workflow-run]");
        if (runBtn) {
          runBtn.removeAttribute("data-workflow-run");
          runBtn.dataset.workflowStop = data.run_id;
          runBtn.className = "btn";
          runBtn.disabled = false;
          runBtn.innerHTML = "";
          runBtn.appendChild(icon("i-stop"));
          runBtn.appendChild(document.createTextNode("Stop"));
        }
      }
      return;
    }

    if (name === "step") {
      var node = nodes[data.position];
      var href = data.step_run_id ? "/runs/" + data.step_run_id : null;
      if (node) setStep(node, data.state, href, data.step_run_id || null);
      if (node && data.step_run_id) stepFrame(node, data.step_run_id, data.state === "running");
      if (node && data.workflow_ref && data.workflow_run_ref) subChain(node, data.workflow_ref, data.workflow_run_ref);
      // A window open on this page follows the workflow: as a step
      // starts, the window attaches to its run.
      if (data.state === "running" && data.step_run_id && session && dialog && dialog.open) attach(data.step_run_id);
      if (row) {
        var a = row.querySelector('[data-pos="' + data.position + '"]');
        if (a) {
          a.querySelector(".dot").className = "dot " + data.state;
          if (href) a.href = href;
        }
      }
      return;
    }

    if (name === "done") {
      if (session && dialog && dialog.open) {
        notice("The workflow ended: " + data.state + (data.message ? ", " + data.message : "") + ".", data.state === "ok" ? "" : "bad");
      }
      if (latest) {
        latest.innerHTML = "";
        latest.appendChild(document.createTextNode("Latest run "));
        latest.appendChild(stateSpan(data.state, data.message));
        latest.appendChild(document.createTextNode(" just now by " + (data.user_name || "")));
      }
      if (row) {
        var result = row.querySelector("[data-result]");
        result.innerHTML = "";
        result.appendChild(stateSpan(data.state, data.message));
        // A run that ended takes Summary, where the page offers it.
        var cell = row.querySelector("td.actions");
        if (cell && runs.hasAttribute("data-summarize") && !cell.querySelector("[data-summary]")) {
          var sum = el("button", "btn btn-sm", "Summary");
          sum.type = "button";
          sum.dataset.summary = "/api/workflow-runs/" + encodeURIComponent(data.run_id) + "/summary";
          sum.dataset.summaryPage = location.pathname.replace(/\/$/, "") + "/runs/" + encodeURIComponent(data.run_id);
          var heading = document.querySelector(".page-head h2");
          sum.dataset.summaryTitle = heading ? heading.textContent : "";
          cell.appendChild(document.createTextNode(" "));
          cell.appendChild(sum);
        }
      }
      if (actions) {
        var stopBtn = actions.querySelector("[data-workflow-stop]");
        if (stopBtn) {
          stopBtn.removeAttribute("data-workflow-stop");
          stopBtn.dataset.workflowRun = data.workflow_id;
          stopBtn.className = "btn btn-primary";
          stopBtn.disabled = false;
          stopBtn.innerHTML = "";
          stopBtn.appendChild(icon("i-play"));
          stopBtn.appendChild(document.createTextNode("Run"));
        }
      }
    }
  }

  // subChain follows the workflow a step runs: its steps are drawn under
  // the step as its topic reports them, with their terminals, and a
  // window open on the page follows into them.
  var subSources = {};
  function subChain(node, workflowId, runId) {
    var sub = node.querySelector(".pipeline.sub");
    if (!sub) {
      sub = el("ol", "pipeline sub");
      node.querySelector(".what").appendChild(sub);
    }
    if (sub.dataset.sub !== runId) {
      sub.innerHTML = "";
      sub.dataset.sub = runId;
      sub.dataset.workflow = workflowId;
    }
    if (subSources[workflowId]) return;
    var source = new EventSource("/api/sse/workflow:" + encodeURIComponent(workflowId));
    subSources[workflowId] = source;
    source.addEventListener("step", function (event) {
      var data;
      try { data = JSON.parse(event.data); } catch (e) { return; }
      var list = document.querySelector('.pipeline.sub[data-sub="' + data.run_id + '"]');
      if (!list) return;
      var item = list.querySelector('[data-step="' + data.position + '"]');
      if (!item) {
        item = el("li", "node");
        item.dataset.step = data.position;
        var rail = el("span", "rail");
        rail.appendChild(el("i", "dot"));
        var what = el("div", "what");
        var a = el("a");
        if (data.workflow_ref) {
          a.href = "/workflows/" + data.workflow_ref;
          a.appendChild(icon("i-flow"));
          a.appendChild(document.createTextNode("workflow"));
        } else {
          a.href = folderHref(data.folder) + "/" + data.job;
          a.textContent = data.folder + " / " + data.job;
        }
        var head = el("div", "step-head");
        head.appendChild(a);
        what.appendChild(head);
        what.appendChild(el("small"));
        item.appendChild(rail);
        item.appendChild(what);
        list.appendChild(item);
      }
      var href = data.step_run_id ? "/runs/" + data.step_run_id : null;
      setStep(item, data.state, href, data.step_run_id || null);
      if (data.step_run_id) stepFrame(item, data.step_run_id, data.state === "running");
      if (data.state === "running" && data.step_run_id && session && dialog && dialog.open) attach(data.step_run_id);
    });
  }

  // stepFrame puts the step's terminal under its name, with Expand on the
  // name's row: polled while the run goes, fetched once more when it
  // ended, when Expand goes if the screen shows the whole output.
  function stepFrame(node, runId, running) {
    var what = node.querySelector(":scope > .what");
    var head = what.querySelector(":scope > .step-head");
    var frame = what.querySelector(":scope > .frame");
    if (!frame) {
      frame = el("a", "frame");
      frame.setAttribute("role", "img");
      what.insertBefore(frame, what.querySelector(":scope > .pipeline.sub"));
    }
    var button = head.querySelector("[data-expand]");
    if (!button) {
      button = el("button", "btn btn-sm");
      button.type = "button";
      button.setAttribute("aria-expanded", "false");
      button.appendChild(icon("i-expand"));
      button.appendChild(el("span", "", "Expand"));
      head.appendChild(button);
    }
    button.dataset.expand = runId;
    frame.href = "/runs/" + runId;
    frame.dataset.attach = runId;
    if (running) {
      frame.dataset.frame = runId;
      watchFrames();
      return;
    }
    delete frame.dataset.frame;
    fetch(frameURL(runId, frame.dataset.all), { credentials: "same-origin" })
      .then(function (res) {
        var lines = res.ok && res.headers ? res.headers.get("X-Output-Lines") : null;
        if (res.ok && res.headers && res.headers.get("X-Output-Complete") === "true" && !frame.dataset.all) button.remove();
        else if (lines && !frame.dataset.all) {
          button.dataset.label = "Expand " + lines + " lines";
          button.querySelector("span").textContent = button.dataset.label;
        }
        return res.ok ? res.text() : null;
      })
      .then(function (svg) { if (svg) frame.innerHTML = svg; })
      .catch(function () {});
  }

  // setStep draws a pipeline node in a state, linked to its run when it
  // has one.
  function setStep(node, state, href, runId) {
    node.className = "node " + state;
    node.querySelector(".dot").className = "dot " + state;
    node.querySelector("small").textContent = state === "pending" ? "step " + (Number(node.dataset.step) + 1) : state;
    var a = node.querySelector("a");
    if (href) {
      a.href = href;
      a.dataset.runLink = runId;
    } else if (state === "pending" && a.dataset.runLink) {
      a.href = a.dataset.target || a.href;
      delete a.dataset.runLink;
    }
  }

  if (live && window.EventSource) {
    // Each link keeps where it pointed before a run, to point back to.
    document.querySelectorAll("[data-pipeline] [data-step] > .what > .step-head > a").forEach(function (a) {
      if (!a.dataset.attach) a.dataset.target = a.getAttribute("href");
    });
    // The workflows the steps run are followed too, for their chains.
    document.querySelectorAll("[data-pipeline] > [data-step] .pipeline.sub[data-workflow]").forEach(function (sub) {
      subChain(sub.closest("[data-step]"), sub.dataset.workflow, sub.dataset.sub);
    });
    var source = new EventSource("/api/sse/" + encodeURIComponent(live.dataset.sse));
    ["run", "step", "done"].forEach(function (name) {
      source.addEventListener(name, function (event) {
        var data;
        try { data = JSON.parse(event.data); } catch (e) { return; }
        workflowEvent(name, data);
      });
    });
  }

  // Show beside a file a folder's pipeline reads opens it, coloured, in
  // the source window, as large as the run window: one request, kept.
  var sourceDialog = document.getElementById("source-dialog");
  function openSource(title, html) {
    if (!sourceDialog) return;
    document.getElementById("source-title").textContent = title;
    document.getElementById("source-code").innerHTML = html;
    sourceDialog.showModal();
  }
  if (sourceDialog) {
    document.getElementById("source-close").addEventListener("click", function () { sourceDialog.close(); });
  }
  document.addEventListener("click", function (event) {
    var btn = event.target.closest("[data-source][data-file]");
    if (!btn) return;
    event.preventDefault();
    var title = btn.dataset.source + "/" + btn.dataset.file;
    if (btn.sourceHTML !== undefined) {
      openSource(title, btn.sourceHTML);
      return;
    }
    btn.disabled = true;
    api("GET", "/api/sources/" + btn.dataset.source.split("/").map(encodeURIComponent).join("/") + "?file=" + encodeURIComponent(btn.dataset.file))
      .then(function (res) {
        btn.disabled = false;
        if (res.status !== 200) {
          openSource(title, escapeHTML(res.data.error || ("HTTP " + res.status)));
          return;
        }
        btn.sourceHTML = res.data.html;
        openSource(title, btn.sourceHTML);
      });
  });

  // The preview window: a file of the artifact tree, an image scaled to
  // fit, markdown rendered, or text, with a pick of the versions the runs
  // wrote, or a file of a checkout the same way, without versions. The link still opens the file on its own
  // with Ctrl or a middle click.
  var previewDialog = document.getElementById("preview-dialog");
  if (previewDialog) {
    var previewImage = document.getElementById("preview-image");
    var previewDoc = document.getElementById("preview-doc");
    var previewVersion = document.getElementById("preview-version");
    var previewDownload = document.getElementById("preview-download");
    var previewPath = "";
    document.getElementById("preview-close").addEventListener("click", function () { previewDialog.close(); });
    previewDialog.addEventListener("close", function () { previewImage.removeAttribute("src"); previewDoc.innerHTML = ""; });

    var fileURL = function (rel, version) {
      return "/artifacts/files/" + rel.split("/").map(encodeURIComponent).join("/") + (version ? "?version=" + encodeURIComponent(version) : "");
    };
    var showImage = function (src) {
      previewDoc.hidden = true;
      previewImage.hidden = false;
      previewImage.src = src;
    };
    var showDoc = function (html, kind) {
      previewImage.hidden = true;
      previewImage.removeAttribute("src");
      previewDoc.hidden = false;
      previewDoc.className = "preview-doc " + kind;
      previewDoc.innerHTML = html;
    };
    var previewLabel = document.getElementById("preview-label");
    var showView = function (view, src) {
      previewDownload.setAttribute("href", src + (src.indexOf("?") < 0 ? "?" : "&") + "download");
      if (view.kind === "image") showImage(src);
      else if (view.kind === "markdown") showDoc(view.html, "markdown");
      else if (view.kind === "text") {
        var cut = view.truncated ? '<p class="faint">The first ' + escapeHTML(String(view.text.length)) + " characters of " +
          escapeHTML(String(view.byte_size)) + ' bytes. Download it for the rest.</p>' : "";
        showDoc(cut + "<pre>" + (view.html || escapeHTML(view.text)) + "</pre>", "text");
      }
      else showDoc('<p class="faint">No preview for this file, ' + escapeHTML(view.content_type) + ". Download it to open it.</p>", "text");
    };
    // previewActions carries a file's Run, Run force or Attach from its
    // row of the Files panel into the window's foot; the window closes for
    // the run window to open.
    var previewActions = document.getElementById("preview-actions");
    previewActions.addEventListener("click", function (event) {
      if (event.target.closest("[data-run], [data-attach]")) previewDialog.close();
    });
    var showActions = function (href) {
      previewActions.innerHTML = "";
      var link = Array.prototype.find.call(document.querySelectorAll("a[data-file]"), function (a) {
        return a.getAttribute("href") === href;
      });
      var row = link && link.closest("tr");
      if (!row) return;
      row.querySelectorAll(".actions > [data-run], .actions > [data-attach], .actions > .runsplit").forEach(function (btn) {
        var copy = btn.cloneNode(true);
        copy.classList.remove("btn-sm");
        copy.querySelectorAll(".btn-sm").forEach(function (part) { part.classList.remove("btn-sm"); });
        previewActions.appendChild(copy);
      });
    };
    // openCheckout opens a file of a folder's checkout: its link is where
    // it opens on its own, and the API answers the same path.
    var openCheckout = function (href, rel) {
      previewPath = href;
      previewLabel.textContent = "File";
      previewVersion.hidden = true;
      showActions(href);
      document.getElementById("preview-title").textContent = rel;
      showDoc('<p class="faint">Loading…</p>', "text");
      if (!previewDialog.open) previewDialog.showModal();
      api("GET", "/api" + href).then(function (res) {
        if (href !== previewPath) return;
        if (res.status !== 200) {
          showDoc('<p class="form-error">' + escapeHTML(res.data.error || ("HTTP " + res.status)) + "</p>", "text");
          return;
        }
        showView(res.data, href);
      });
    };
    var openFile = function (rel, version) {
      previewPath = rel;
      previewLabel.textContent = "Artifact";
      previewActions.innerHTML = "";
      document.getElementById("preview-title").textContent = rel;
      showDoc('<p class="faint">Loading…</p>', "text");
      if (!previewDialog.open) previewDialog.showModal();
      api("GET", "/api" + fileURL(rel, version)).then(function (res) {
        if (rel !== previewPath) return;
        if (res.status !== 200) {
          showDoc('<p class="form-error">' + escapeHTML(res.data.error || ("HTTP " + res.status)) + "</p>", "text");
          return;
        }
        var view = res.data;
        previewVersion.innerHTML = view.versions.map(function (v, i) {
          var when = v.created_at ? new Date(v.created_at).toLocaleString() : v.commit.slice(0, 7);
          return '<option value="' + escapeHTML(v.commit) + '"' + (v.commit === view.version ? " selected" : "") + ">" +
            escapeHTML((i === 0 ? "latest · " : "") + when + " · " + (v.run_id ? v.folder + "/" + v.job : "by " + v.user_name)) + "</option>";
        }).join("");
        previewVersion.hidden = view.versions.length < 2;
        showView(view, fileURL(rel, view.version));
      });
    };
    previewVersion.addEventListener("change", function () { openFile(previewPath, previewVersion.value); });

    document.addEventListener("click", function (event) {
      if (event.ctrlKey || event.metaKey || event.shiftKey || event.button !== 0) return;
      var artifact = event.target.closest("a[data-artifact]");
      if (artifact) {
        event.preventDefault();
        openFile(artifact.dataset.artifact, artifact.dataset.version || "");
        return;
      }
      var checkout = event.target.closest("a[data-file]");
      if (!checkout) return;
      event.preventDefault();
      openCheckout(checkout.getAttribute("href"), checkout.dataset.file);
    });

    // #file=<path> opens a file of the Files panel on load, so a preview
    // can be shared.
    var fileHash = /^#file=(.+)$/.exec(location.hash);
    if (fileHash) {
      var listed = Array.prototype.find.call(document.querySelectorAll("a[data-file]"), function (a) {
        return a.dataset.file === decodeURIComponent(fileHash[1]);
      });
      if (listed) openCheckout(listed.getAttribute("href"), listed.dataset.file);
    }
  }

  // View on a workflow run opens its pipeline in a window: the run's page,
  // fetched, its pipeline part shown. Ctrl or a middle click opens the
  // page itself.
  var pipelineDialog = document.getElementById("pipeline-dialog");
  if (pipelineDialog) {
    var pipelineBody = document.getElementById("pipeline-body");
    document.getElementById("pipeline-close").addEventListener("click", function () { pipelineDialog.close(); });
    pipelineDialog.addEventListener("close", function () { pipelineBody.innerHTML = ""; });
    document.addEventListener("click", function (event) {
      var link = event.target.closest("a[data-workflow-view]");
      if (!link || event.ctrlKey || event.metaKey || event.shiftKey || event.button !== 0) return;
      event.preventDefault();
      var href = link.getAttribute("href");
      var title = document.querySelector(".page-head h2");
      document.getElementById("pipeline-title").textContent = title ? title.textContent : "";
      document.getElementById("pipeline-page").setAttribute("href", href);
      pipelineBody.innerHTML = '<p class="faint">Loading…</p>';
      pipelineDialog.showModal();
      fetch(href, { credentials: "same-origin" }).then(function (res) { return res.text(); }).then(function (text) {
        var view = new DOMParser().parseFromString(text, "text/html").querySelector("[data-pipeline-view]");
        pipelineBody.innerHTML = view ? view.innerHTML : '<p class="form-error">The run is not there.</p>';
      }).catch(function () {
        pipelineBody.innerHTML = '<p class="form-error">The run did not load.</p>';
      });
    });
  }

  // Summary on a run or a workflow run asks Claude for a summary of its
  // output, once: data-summary is the API route, which keeps it, and every
  // later ask returns what it kept; data-summary-page is the run's page.
  var summaryDialog = document.getElementById("summary-dialog");
  if (summaryDialog) {
    var summaryText = document.getElementById("summary-text");
    document.getElementById("summary-close").addEventListener("click", function () { summaryDialog.close(); });
    document.addEventListener("click", function (event) {
      var btn = event.target.closest("[data-summary]");
      if (!btn) return;
      event.preventDefault();
      var meta = document.getElementById("summary-meta");
      document.getElementById("summary-title").textContent = btn.dataset.summaryTitle || "";
      document.getElementById("summary-run").setAttribute("href", btn.dataset.summaryPage || "#");
      summaryText.innerHTML = '<p class="faint">Claude is reading the output…</p>';
      meta.textContent = "";
      summaryDialog.showModal();
      btn.disabled = true;
      api("POST", btn.dataset.summary, {}).then(function (res) {
        btn.disabled = false;
        if (res.status !== 200) {
          summaryText.innerHTML = '<p class="form-error">' + escapeHTML(res.data.error || ("HTTP " + res.status)) + "</p>";
          return;
        }
        var when = res.data.created_at ? new Date(res.data.created_at).toLocaleString() : "";
        summaryText.innerHTML = res.data.html;
        meta.textContent = (res.data.kept ? "Kept from " : "Written ") + when + " for " + res.data.user_name + ".";
      });
    });
  }

  // Notes: a form posts a note or a reply, and the list follows the notes
  // topic, fetched again when a note arrives or changes. A reply form
  // being written keeps its text.
  var notesList = document.querySelector("[data-notes]");
  function refreshNotes() {
    if (!notesList) return Promise.resolve();
    return fetch(location.pathname + location.search, { credentials: "same-origin" })
      .then(function (res) { return res.text(); })
      .then(function (text) {
        var fresh = new DOMParser().parseFromString(text, "text/html").querySelector("[data-notes]");
        if (!fresh) return;
        var drafts = {};
        notesList.querySelectorAll("form[data-parent]").forEach(function (form) {
          var value = form.elements.body.value;
          if (value) drafts[form.dataset.parent] = value;
        });
        var edits = {};
        notesList.querySelectorAll("form[data-note-edit]").forEach(function (form) {
          edits[form.dataset.noteEdit] = form.elements.body.value;
        });
        notesList.innerHTML = fresh.innerHTML;
        // A hidden box, a running run's, shows once a thread comes.
        var box = notesList.closest(".promptbox");
        if (box && box.hidden && notesList.children.length) box.hidden = false;
        Object.keys(drafts).forEach(function (parent) {
          var form = notesList.querySelector('form[data-parent="' + parent + '"]');
          if (!form) return;
          form.elements.body.value = drafts[parent];
          form.closest("details").open = true;
        });
        mentionFields(notesList);
        Object.keys(edits).forEach(function (id) {
          var body = notesList.querySelector('[data-edit="' + id + '"]');
          if (body) editNote(body, edits[id]);
        });
      });
  }
  document.addEventListener("submit", function (event) {
    var form = event.target.closest("form[data-note-form]");
    if (!form) return;
    event.preventDefault();
    var body = form.elements.body.value.trim();
    if (!body) return;
    var button = form.querySelector("button[type=submit]");
    button.disabled = true;
    api("POST", "/api/notes", { parent_id: form.dataset.parent || "", body: body })
      .then(function (res) {
        button.disabled = false;
        if (res.status >= 300) {
          window.alert(res.data.error || ("HTTP " + res.status));
          return;
        }
        form.elements.body.value = "";
        if (form.elements.body.hasAttribute("data-mentions")) paintMentions(form.elements.body);
        refreshNotes();
      });
  });
  // Its author, or an admin, clicks a note's text to edit it in place:
  // Ctrl+Enter or Save puts it, Esc or Cancel leaves it as it was. A note
  // to the agent runs again, and its answer follows over the notes topic.
  function editNote(body, value) {
    var form = document.createElement("form");
    form.className = "note-form note-edit";
    form.dataset.noteEdit = body.dataset.edit;
    form.innerHTML = '<textarea name="body" rows="3" required data-mentions aria-label="The note"></textarea>' +
      '<div class="note-form-foot"><span class="faint">Ctrl+Enter saves, Esc cancels.</span>' +
      '<span><button class="btn" type="button" data-edit-cancel>Cancel</button> <button class="btn btn-primary" type="submit">Save</button></span></div>';
    form.elements.body.value = value === undefined ? body.dataset.body : value;
    body.hidden = true;
    body.after(form);
    mentionField(form.elements.body);
    form.elements.body.focus();
  }
  function closeEdit(form) {
    var body = form.previousElementSibling;
    if (body && body.hasAttribute("data-edit")) body.hidden = false;
    form.remove();
  }
  document.addEventListener("click", function (event) {
    var cancel = event.target.closest("[data-edit-cancel]");
    if (cancel) {
      closeEdit(cancel.closest("form[data-note-edit]"));
      return;
    }
    var body = event.target.closest(".note-body[data-edit]");
    if (!body || event.target.closest("a, button, summary")) return;
    if (String(window.getSelection && window.getSelection()) !== "") return;
    editNote(body);
  });
  document.addEventListener("submit", function (event) {
    var form = event.target.closest("form[data-note-edit]");
    if (!form) return;
    event.preventDefault();
    var value = form.elements.body.value.trim();
    if (!value) return;
    var button = form.querySelector("button[type=submit]");
    button.disabled = true;
    api("PUT", "/api/notes/" + encodeURIComponent(form.dataset.noteEdit), { body: value })
      .then(function (res) {
        button.disabled = false;
        if (res.status >= 300) {
          window.alert(res.data.error || ("HTTP " + res.status));
          return;
        }
        closeEdit(form);
        refreshNotes();
      });
  });
  document.addEventListener("keydown", function (event) {
    var editing = event.target.closest && event.target.closest("form[data-note-edit] textarea");
    if (editing && event.key === "Escape") {
      event.preventDefault();
      closeEdit(editing.form);
      return;
    }
    if (editing && event.key === "Enter" && (event.ctrlKey || event.metaKey)) {
      event.preventDefault();
      editing.form.requestSubmit();
      return;
    }
    var area = event.target.closest && event.target.closest("form[data-note-form] textarea");
    if (!area || event.key !== "Enter") return;
    if (event.ctrlKey || event.metaKey) {
      event.preventDefault();
      area.form.requestSubmit();
    }
  });
  // A note's field, data-mentions, lists the users whose name follows an
  // @ as model.Mentions reads one, and draws a mention of a user in the
  // colour of a link on a layer behind its clear text. The list opens
  // under the caret: up and down move, Enter or Tab takes a name, Esc
  // closes it.
  var mentionAt = /(?:^|[\s(\[{,;])@([A-Za-z0-9._-]*)$/;
  // A colon and a character other than a space after the name make a
  // job ID, @host:dev1, no mention, as model.Mentions reads it.
  var mentionAll = /(^|[\s(\[{,;])@([A-Za-z0-9][A-Za-z0-9._-]*)(?![A-Za-z0-9._-]|:\S)/g;
  var mentionCopy = ["boxSizing", "width", "fontFamily", "fontSize", "fontWeight", "fontStyle", "lineHeight", "letterSpacing", "wordSpacing", "tabSize", "textIndent", "textTransform",
    "paddingTop", "paddingRight", "paddingBottom", "paddingLeft", "borderTopWidth", "borderRightWidth", "borderBottomWidth", "borderLeftWidth", "borderStyle", "borderRadius"];
  var mention = { list: null, area: null, start: 0, names: [], at: 0 };

  function knownUser(name) {
    return Object.prototype.hasOwnProperty.call(roles, name);
  }

  // mentionField wraps a field once with its layer and draws it.
  function mentionField(area) {
    if (!area.parentNode.classList.contains("mention-field")) {
      var wrap = document.createElement("div");
      wrap.className = "mention-field";
      var back = document.createElement("div");
      back.className = "mention-backdrop";
      back.setAttribute("aria-hidden", "true");
      area.before(wrap);
      wrap.append(back, area);
      area.setAttribute("aria-autocomplete", "list");
      area.addEventListener("scroll", function () { back.scrollTop = area.scrollTop; back.scrollLeft = area.scrollLeft; });
      if (window.ResizeObserver) new ResizeObserver(function () { paintMentions(area); }).observe(area);
    }
    paintMentions(area);
  }
  function mentionFields(root) {
    root.querySelectorAll("textarea[data-mentions]").forEach(mentionField);
  }

  // paintMentions draws the text of a field on its layer, a mention of a
  // user marked, in the field's own font, box and scroll.
  function paintMentions(area) {
    var back = area.previousElementSibling;
    if (!back || !back.classList.contains("mention-backdrop")) return;
    var style = getComputedStyle(area);
    mentionCopy.forEach(function (prop) { if (prop !== "width") back.style[prop] = style[prop]; });
    // A scroll bar narrows the field's text; the layer leaves its room.
    var bar = area.offsetWidth - area.clientWidth - parseFloat(style.borderLeftWidth) - parseFloat(style.borderRightWidth);
    back.style.paddingRight = parseFloat(style.paddingRight) + Math.max(0, bar) + "px";
    var text = escapeHTML(area.value).replace(mentionAll, function (m, lead, name) {
      var bare = name.replace(/[.-]+$/, "");
      if (!knownUser(bare)) return m;
      return lead + '<span class="mention-hl">@' + bare + "</span>" + name.slice(bare.length);
    });
    back.innerHTML = text + "\n";
    back.scrollTop = area.scrollTop;
    back.scrollLeft = area.scrollLeft;
  }

  // caretPoint returns where the caret of a field is on the screen, from
  // a copy of the field's text up to it in a hidden element of its box.
  function caretPoint(area, at) {
    var style = getComputedStyle(area), copy = document.createElement("div");
    mentionCopy.forEach(function (prop) { copy.style[prop] = style[prop]; });
    copy.style.position = "absolute";
    copy.style.visibility = "hidden";
    copy.style.top = "0";
    copy.style.left = "-9999px";
    copy.style.whiteSpace = "pre-wrap";
    copy.style.overflowWrap = "break-word";
    copy.style.width = area.clientWidth + parseFloat(style.borderLeftWidth) + parseFloat(style.borderRightWidth) + "px";
    copy.textContent = area.value.slice(0, at);
    var mark = document.createElement("span");
    mark.textContent = area.value.slice(at) || ".";
    copy.append(mark);
    document.body.append(copy);
    var box = area.getBoundingClientRect(), line = parseFloat(style.lineHeight) || parseFloat(style.fontSize) * 1.4;
    var point = { left: box.left + mark.offsetLeft - area.scrollLeft, top: box.top + mark.offsetTop - area.scrollTop + line };
    copy.remove();
    return point;
  }

  function closeMentions() {
    if (mention.area) {
      mention.area.removeAttribute("aria-activedescendant");
      mention.area.removeAttribute("aria-controls");
    }
    if (mention.list) mention.list.hidden = true;
    mention.area = null;
  }

  // findMentions lists the users a word after an @ names: those whose
  // name starts with it first, then those holding it, eight at most.
  function findMentions(query) {
    var q = query.toLowerCase(), names = Object.keys(roles).sort();
    var starts = names.filter(function (n) { return n.toLowerCase().indexOf(q) === 0; });
    var holds = names.filter(function (n) { return n.toLowerCase().indexOf(q) > 0; });
    return starts.concat(holds).slice(0, 8);
  }

  function showMentions(area) {
    var at = area.selectionStart;
    var found = area.selectionStart === area.selectionEnd && mentionAt.exec(area.value.slice(0, at));
    var names = found ? findMentions(found[1]) : [];
    if (!names.length) {
      closeMentions();
      return;
    }
    if (!mention.list) {
      mention.list = document.createElement("ul");
      mention.list.id = "mention-list";
      mention.list.className = "mention-list";
      mention.list.setAttribute("role", "listbox");
      mention.list.setAttribute("aria-label", "Users");
      mention.list.addEventListener("mousedown", function (event) {
        var item = event.target.closest("li");
        event.preventDefault();
        if (item) takeMention(Number(item.dataset.at));
      });
      document.body.append(mention.list);
    }
    var same = mention.area === area && mention.names.join(" ") === names.join(" ");
    mention.area = area;
    mention.start = at - found[1].length - 1;
    mention.names = names;
    if (!same) mention.at = 0;
    mention.list.innerHTML = names.map(function (name, i) {
      var role = roles[name] || "";
      return '<li role="option" id="mention-opt-' + i + '" data-at="' + i + '" aria-selected="' + (i === mention.at) + '">' +
        avatarHTML(name) + " <span>" + escapeHTML(name) + '</span><span class="role">' + escapeHTML(role) + "</span></li>";
    }).join("");
    mention.list.hidden = false;
    var point = caretPoint(area, mention.start);
    mention.list.style.top = point.top + 4 + "px";
    mention.list.style.left = Math.max(8, Math.min(point.left, window.innerWidth - mention.list.offsetWidth - 8)) + "px";
    area.setAttribute("aria-controls", "mention-list");
    area.setAttribute("aria-activedescendant", "mention-opt-" + mention.at);
  }

  function moveMention(by) {
    var n = mention.names.length;
    mention.at = (mention.at + by + n) % n;
    mention.list.querySelectorAll("li").forEach(function (li, i) { li.setAttribute("aria-selected", String(i === mention.at)); });
    mention.area.setAttribute("aria-activedescendant", "mention-opt-" + mention.at);
  }

  // takeMention puts @name and a space in place of the word being typed.
  function takeMention(i) {
    var area = mention.area, name = mention.names[i];
    if (!area || !name) return;
    var end = area.selectionStart, rest = /^[A-Za-z0-9._-]*/.exec(area.value.slice(end))[0];
    area.setRangeText("@" + name + " ", mention.start, end + rest.length, "end");
    closeMentions();
    paintMentions(area);
    area.focus();
  }

  mentionFields(document);
  document.addEventListener("input", function (event) {
    var area = event.target;
    if (!area.matches || !area.matches("textarea[data-mentions]")) return;
    paintMentions(area);
    showMentions(area);
  });
  document.addEventListener("click", function (event) {
    if (event.target.matches && event.target.matches("textarea[data-mentions]")) showMentions(event.target);
  });
  document.addEventListener("focusout", function (event) {
    if (event.target === mention.area) closeMentions();
  });
  window.addEventListener("scroll", function (event) { if (mention.area && event.target !== mention.area) closeMentions(); }, true);
  window.addEventListener("resize", closeMentions);
  // The list takes its keys before the field's own: Enter does not post,
  // Esc does not cancel an edit, while it is open.
  document.addEventListener("keydown", function (event) {
    if (!mention.area || event.target !== mention.area || mention.list.hidden) return;
    var key = event.key;
    if (key === "ArrowDown" || key === "ArrowUp") {
      moveMention(key === "ArrowDown" ? 1 : -1);
    } else if ((key === "Enter" && !event.ctrlKey && !event.metaKey) || key === "Tab") {
      takeMention(mention.at);
    } else if (key === "Escape") {
      closeMentions();
    } else {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
  }, true);
  document.addEventListener("keyup", function (event) {
    var area = event.target;
    if (event.key !== "ArrowLeft" && event.key !== "ArrowRight" && event.key !== "Home" && event.key !== "End") return;
    if (area.matches && area.matches("textarea[data-mentions]")) showMentions(area);
  });

  // The board follows its own notes, a prompt box the notes of its
  // place, data-dir, as the note's dir names it.
  if (notesList && window.EventSource) {
    var pending;
    var place = notesList.dataset.dir || "";
    new EventSource("/api/sse/notes").addEventListener("note", function (event) {
      var changed = {};
      try { changed = JSON.parse(event.data); } catch (e) { /* refresh anyway */ }
      if (changed.id && (changed.dir || "") !== place) return;
      clearTimeout(pending);
      pending = setTimeout(refreshNotes, 150);
    });
  }

  // The header's prompt asks Claude where the page is, as the header's
  // data-prompt-* name it: a folder or a directory in it, a run, or the
  // root, _. A panel under the header follows the thread over the notes
  // topic: the place, the question, working, then the answer, rendered as
  // the thread's page renders it. The logo opens and closes the panel;
  // Esc or a click outside closes it. The tab keeps the panel's thread
  // and whether it is open, so another page brings the same thread back.
  var headerPrompt = document.getElementById("header-prompt");
  var headerAnswer = document.getElementById("header-answer");
  var headerToggle = document.getElementById("header-toggle");
  var headerTop = document.querySelector("header.top[data-prompt-folder]");
  var headerOpen = function () { return !!headerAnswer && !headerAnswer.hidden; };
  if (headerPrompt && headerAnswer && headerTop) {
    var here = {
      folder: headerTop.dataset.promptFolder, dir: headerTop.dataset.promptDir || "", run: headerTop.dataset.promptRun || "",
      label: headerTop.dataset.promptLabel, href: headerTop.dataset.promptHref
    };
    var me = document.body.dataset.me || "";
    var stateKey = "task-ui.header-prompt";
    var follow = null;
    var shown = "";
    var saved = function () {
      try { return JSON.parse(sessionStorage.getItem(stateKey)) || {}; } catch (e) { return {}; }
    };
    var save = function (state) {
      try { sessionStorage.setItem(stateKey, JSON.stringify(state)); } catch (e) { /* the tab forgets */ }
    };
    var typed = function () { headerPrompt.classList.toggle("typed", headerPrompt.value !== ""); };
    var unfollow = function () {
      if (follow) follow.close();
      follow = null;
    };
    var threadsURL = function (place) {
      if (place.run) return "/api/prompts?limit=20&run=" + encodeURIComponent(place.run);
      return "/api/prompts?limit=20&folder=" + encodeURIComponent(place.folder) + (place.dir ? "&dir=" + encodeURIComponent(place.dir) : "");
    };
    var setOpen = function (open) {
      headerAnswer.hidden = !open;
      if (headerToggle) headerToggle.setAttribute("aria-expanded", open ? "true" : "false");
      var state = saved();
      state.open = open;
      save(state);
      if (!open) unfollow();
    };
    // frame draws the panel for a thread in a place: the place, linked to
    // the thread, the question, its state, and room for the answer.
    var frame = function (place, question, id) {
      shown = id || "";
      headerAnswer.innerHTML = '<p class="place">@claude-prompt · <a></a></p><p class="question"></p><p class="working" data-state>open</p><div class="answer note-body"></div>';
      var link = headerAnswer.querySelector(".place a");
      link.textContent = place.label;
      link.href = place.href + (id ? "#note-" + encodeURIComponent(id) : "");
      headerAnswer.querySelector(".question").textContent = question;
    };
    var setState = function (state, text) {
      var line = headerAnswer.querySelector("[data-state]");
      if (!line) return;
      if (state === "done") {
        line.remove();
        return;
      }
      line.className = state === "failed" ? "failed" : "working";
      line.textContent = text || state;
    };
    // answer puts the last reply of the thread in the panel: as the
    // thread's page renders it, or as the summaries render markdown.
    var answer = function (place, id, reply) {
      var put = function (html) {
        var box = headerAnswer.querySelector(".answer");
        if (shown === id && box && html) box.innerHTML = html;
      };
      put(markdownHTML(reply.body));
      fetch(place.href, { credentials: "same-origin" })
        .then(function (res) { return res.ok ? res.text() : ""; })
        .then(function (text) {
          var doc = new DOMParser().parseFromString(text, "text/html");
          var bodies = doc.querySelectorAll("#note-" + CSS.escape(id) + " .replies > li .note-body");
          if (bodies.length) put(bodies[bodies.length - 1].innerHTML);
        })
        .catch(function () {});
    };
    // load draws a thread from its place's list, and follows it while it
    // runs; false when the place no longer lists it.
    var load = function (place, id) {
      return api("GET", threadsURL(place)).then(function (res) {
        var thread = ((res.data && res.data.notes) || []).filter(function (n) { return n.id === id; })[0];
        if (!thread) return false;
        frame(place, thread.body, thread.id);
        setState(thread.state || "done", thread.state === "failed" ? "failed" : "");
        var replies = thread.replies || [];
        if (replies.length && thread.state !== "open" && thread.state !== "working") answer(place, id, replies[replies.length - 1]);
        else watch(place, id);
        return true;
      });
    };
    // watch follows a thread over the notes topic until it ends.
    var watch = function (place, id) {
      unfollow();
      if (!window.EventSource) return;
      follow = new EventSource("/api/sse/notes");
      follow.addEventListener("note", function (event) {
        var changed = {};
        try { changed = JSON.parse(event.data); } catch (e) { return; }
        if (shown !== id || changed.id !== id) return;
        if (changed.state === "done" || changed.state === "failed") {
          unfollow();
          load(place, id);
        } else if (changed.state) {
          setState(changed.state);
        }
      });
    };
    // reopen shows the tab's thread, or else the user's latest here, or
    // an empty panel with the field focused.
    var reopen = function () {
      var state = saved();
      var empty = function () {
        frame(here, "", "");
        headerAnswer.querySelector(".question").remove();
        headerAnswer.querySelector("[data-state]").remove();
        headerPrompt.focus();
      };
      if (state.thread && state.place) {
        return load(state.place, state.thread).then(function (found) {
          if (found) return;
          save({ open: true });
          return reopen();
        });
      }
      return api("GET", threadsURL(here)).then(function (res) {
        var mine = ((res.data && res.data.notes) || []).filter(function (n) { return n.user_name === me; })[0];
        if (!mine) return empty();
        save({ open: true, thread: mine.id, place: here });
        return load(here, mine.id);
      });
    };
    var ask = function (text) {
      unfollow();
      frame(here, text, "");
      setOpen(true);
      var body = here.run ? { run: here.run, prompt: text } : { folder: here.folder, dir: here.dir, prompt: text };
      api("POST", "/api/prompts", body).then(function (res) {
        if (res.status >= 300) {
          setState("failed", res.data.error || ("HTTP " + res.status));
          return;
        }
        var id = res.data.id;
        save({ open: true, thread: id, place: here });
        frame(here, text, id);
        setState(res.data.state || "open");
        watch(here, id);
      });
    };
    if (headerToggle) {
      headerToggle.addEventListener("click", function (event) {
        event.stopPropagation();
        if (headerOpen()) {
          setOpen(false);
          return;
        }
        setOpen(true);
        reopen();
      });
    }
    headerPrompt.addEventListener("input", typed);
    headerPrompt.addEventListener("keydown", function (event) {
      if (event.key === "Escape") {
        event.preventDefault();
        headerPrompt.value = "";
        typed();
        headerPrompt.blur();
        return;
      }
      if (event.key !== "Enter" || event.isComposing) return;
      event.preventDefault();
      var text = headerPrompt.value.trim();
      if (!text) return;
      headerPrompt.value = "";
      typed();
      ask(text);
    });
    document.addEventListener("keydown", function (event) {
      if (event.key === "Escape" && headerOpen() && event.target !== headerPrompt) setOpen(false);
    });
    document.addEventListener("click", function (event) {
      if (!headerOpen() || event.target === headerPrompt || headerAnswer.contains(event.target)) return;
      setOpen(false);
    });
    if (saved().open && saved().thread) {
      setOpen(true);
      reopen();
    }
  }

  // The inventory runs a command on the hosts checked, all at once, and
  // shows each host's output and exit code. A quick button fills in a
  // command.
  var inventoryRun = document.querySelector("form[data-inventory-run]");
  if (inventoryRun) {
    var results = document.querySelector("[data-inventory-results]");
    inventoryRun.addEventListener("click", function (event) {
      var quick = event.target.closest("[data-command]");
      if (!quick) return;
      inventoryRun.elements.command.value = quick.dataset.command;
      inventoryRun.elements.command.focus();
    });
    inventoryRun.elements.command.addEventListener("keydown", function (event) {
      if (event.key === "Enter" && (event.ctrlKey || event.metaKey)) {
        event.preventDefault();
        inventoryRun.requestSubmit();
      }
    });
    inventoryRun.addEventListener("submit", function (event) {
      event.preventDefault();
      var summarize = !!(event.submitter && event.submitter.name === "summarize");
      var hosts = Array.prototype.filter.call(inventoryRun.querySelectorAll("input[name=hosts]"), function (box) { return box.checked; })
        .map(function (box) { return box.value; });
      if (!hosts.length) return;
      var buttons = inventoryRun.querySelectorAll("button[type=submit]");
      buttons.forEach(function (button) { button.disabled = true; });
      // Each host's run is a run of @host:<host>, under one batch: the page
      // shows the batch, the user's last, as it loads.
      api("POST", "/api/inventory/run", { command: inventoryRun.elements.command.value, hosts: hosts, summarize: summarize }).then(function (res) {
        buttons.forEach(function (button) { button.disabled = false; });
        if (res.status >= 300) {
          results.innerHTML = '<p class="form-error">' + escapeHTML(res.data.error || ("HTTP " + res.status)) + "</p>";
          return;
        }
        window.location.reload();
      });
    });
    // While a run of the batch runs, the page fetches itself and puts
    // the batch back, the outputs of the runs that ended read from
    // their recordings.
    var followBatch = function () {
      if (!results.querySelector("[data-batch-live]")) return;
      setTimeout(function () {
        if (document.hidden) return followBatch();
        fetch(location.pathname, { credentials: "same-origin" })
          .then(function (res) { return res.ok ? res.text() : null; })
          .then(function (text) {
            var fresh = text && new DOMParser().parseFromString(text, "text/html").querySelector("[data-inventory-results]");
            if (fresh) {
              results.innerHTML = fresh.innerHTML;
              readOutputs(results);
            }
            followBatch();
          })
          .catch(followBatch);
      }, 2000);
    };
    followBatch();
  }

  // The inventory's hosts refresh once a minute, the rate task-ui reads
  // Beszel: the page fetches itself and puts its hosts section back,
  // the time it was drawn with it. A hidden tab waits.
  var hostsLive = document.querySelector("[data-hosts-live]");
  if (hostsLive) {
    setInterval(function () {
      if (document.hidden) return;
      fetch(location.pathname, { credentials: "same-origin" })
        .then(function (res) { return res.ok ? res.text() : null; })
        .then(function (text) {
          if (!text) return;
          var fresh = new DOMParser().parseFromString(text, "text/html").querySelector("[data-hosts-live]");
          if (fresh) {
            hostsLive.innerHTML = fresh.innerHTML;
            reopenHosts();
          }
        })
        .catch(function () {});
    }, 60000);
  }

  // A click on a host's line, or Enter or Space on it, opens its
  // expanded row under it: the host's page, fetched, kept 30 seconds.
  // The open hosts outlast a refresh, and #host=<name> opens one.
  var hostDetails = {};
  var openHosts = function () {
    try { return JSON.parse(localStorage.getItem("inventory.open") || "[]"); } catch (e) { return []; }
  };
  var keepOpen = function (name, open) {
    var names = openHosts().filter(function (n) { return n !== name; });
    if (open) names.push(name);
    try { localStorage.setItem("inventory.open", JSON.stringify(names)); } catch (e) {}
  };
  var showHost = function (row, open) {
    var name = row.getAttribute("data-host-row");
    var next = row.nextElementSibling;
    if (next && next.classList.contains("host-detail-row")) next.remove();
    row.setAttribute("aria-expanded", open ? "true" : "false");
    if (!open) return;
    var detail = document.createElement("tr");
    detail.className = "host-detail-row";
    detail.innerHTML = '<td colspan="' + row.children.length + '"><p class="faint">Reading ' + escapeHTML(name) + "…</p></td>";
    row.after(detail);
    var wrap = row.closest(".table-wrap");
    if (wrap) wrap.style.setProperty("--detail-w", (wrap.clientWidth - 32) + "px");
    var fill = function (html) { detail.firstChild.innerHTML = html; };
    var kept = hostDetails[name];
    if (kept && Date.now() - kept.at < 30000) return fill(kept.html);
    fetch("/inventory/hosts/" + encodeURIComponent(name), { credentials: "same-origin" })
      .then(function (res) { return res.ok ? res.text() : Promise.reject(res.status); })
      .then(function (text) {
        var found = new DOMParser().parseFromString(text, "text/html").querySelector("[data-host-detail]");
        if (!found) return Promise.reject("no detail");
        hostDetails[name] = { html: found.outerHTML, at: Date.now() };
        fill(found.outerHTML);
      })
      .catch(function (why) {
        fill('<p class="form-error">Reading ' + escapeHTML(name) + " failed: " + escapeHTML(typeof why === "number" ? "HTTP " + why : String(why)) + "</p>");
      });
  };
  var reopenHosts = function () {
    if (!hostsLive) return;
    var wanted = openHosts();
    var hash = /^#host=(.+)$/.exec(location.hash);
    if (hash) wanted.push(decodeURIComponent(hash[1]));
    hostsLive.querySelectorAll("[data-host-row]").forEach(function (row) {
      if (wanted.indexOf(row.getAttribute("data-host-row")) >= 0) showHost(row, true);
    });
  };
  if (hostsLive) {
    var toggleHost = function (row) {
      var open = row.getAttribute("aria-expanded") !== "true";
      showHost(row, open);
      keepOpen(row.getAttribute("data-host-row"), open);
    };
    hostsLive.addEventListener("click", function (e) {
      var row = e.target.closest("[data-host-row]");
      if (!row || e.target.closest("button, a, input, label, code")) return;
      toggleHost(row);
    });
    hostsLive.addEventListener("keydown", function (e) {
      if ((e.key !== "Enter" && e.key !== " ") || !e.target.matches("[data-host-row]")) return;
      e.preventDefault();
      toggleHost(e.target);
    });
    reopenHosts();
  }

  // A table with data-sortable sorts by a header with data-sort on a
  // click, by each cell's data-v: numbers busiest first, words A to Z,
  // the next click the other way.
  document.addEventListener("click", function (e) {
    var th = e.target.closest && e.target.closest("table[data-sortable] th[data-sort]");
    if (!th) return;
    var table = th.closest("table");
    var index = Array.prototype.indexOf.call(th.parentNode.children, th);
    var num = th.getAttribute("data-sort") === "num";
    var was = th.getAttribute("aria-sort");
    var down = was ? was !== "descending" : num;
    th.parentNode.querySelectorAll("th").forEach(function (h) { h.removeAttribute("aria-sort"); });
    th.setAttribute("aria-sort", down ? "descending" : "ascending");
    var body = table.tBodies[0];
    var rows = Array.prototype.slice.call(body.rows);
    rows.sort(function (a, b) {
      var x = a.cells[index].getAttribute("data-v") || "", y = b.cells[index].getAttribute("data-v") || "";
      var c = num ? parseFloat(x) - parseFloat(y) : x.localeCompare(y);
      return down ? -c : c;
    });
    rows.forEach(function (row) { body.appendChild(row); });
  });

  // The host picker: All, None and Invert set the boxes, and the count
  // says how many of the hosts a run reaches.
  document.querySelectorAll("[data-host-picker]").forEach(function (picker) {
    var boxes = picker.querySelectorAll("input[name=hosts]");
    var count = picker.querySelector("[data-host-count]");
    var tally = function () {
      var picked = Array.prototype.filter.call(boxes, function (box) { return box.checked; }).length;
      count.textContent = picked + " of " + boxes.length + " host" + (boxes.length === 1 ? "" : "s");
    };
    picker.addEventListener("change", tally);
    picker.addEventListener("click", function (event) {
      var btn = event.target.closest("[data-hosts]");
      if (!btn) return;
      boxes.forEach(function (box) {
        box.checked = btn.dataset.hosts === "all" ? true : btn.dataset.hosts === "none" ? false : !box.checked;
      });
      tally();
    });
    tally();
  });

  // markdownHTML renders the markdown Claude answers in: headings, lists,
  // fences, **bold**, `code` and paragraphs. The text is escaped first.
  function markdownHTML(text) {
    var out = [], list = null, fence = null, para = [];
    function inline(line) {
      return escapeHTML(line)
        .replace(/`([^`]+)`/g, "<code>$1</code>")
        .replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>")
        .replace(/(^|[^*])\*([^*\s][^*]*)\*/g, "$1<em>$2</em>");
    }
    function flush() {
      if (para.length) { out.push("<p>" + para.map(inline).join("<br>") + "</p>"); para = []; }
      if (list) { out.push("</" + list + ">"); list = null; }
    }
    text.replace(/\r/g, "").split("\n").forEach(function (line) {
      if (fence !== null) {
        if (/^```/.test(line)) { out.push("<pre>" + escapeHTML(fence.join("\n")) + "</pre>"); fence = null; }
        else fence.push(line);
        return;
      }
      var m;
      if (/^```/.test(line)) { flush(); fence = []; }
      else if ((m = /^(#{1,6})\s+(.*)$/.exec(line))) { flush(); out.push("<h5>" + inline(m[2]) + "</h5>"); }
      else if ((m = /^\s*(?:[-*+]|(\d+)[.)])\s+(.*)$/.exec(line))) {
        var kind = m[1] ? "ol" : "ul";
        if (para.length) { out.push("<p>" + para.map(inline).join("<br>") + "</p>"); para = []; }
        if (list !== kind) { if (list) out.push("</" + list + ">"); out.push("<" + kind + ">"); list = kind; }
        out.push("<li>" + inline(m[2]) + "</li>");
      }
      else if (!line.trim()) flush();
      else { if (list) { out.push("</" + list + ">"); list = null; } para.push(line); }
    });
    if (fence !== null) out.push("<pre>" + escapeHTML(fence.join("\n")) + "</pre>");
    flush();
    return out.join("");
  }

  // ansiHTML renders text with the colour codes a terminal prints: the
  // 16 colours, the 256, true colour, bold, dim, italic and underline.
  // Every other escape goes, and the text is escaped.
  function ansiHTML(text) {
    var out = "", open = false, state = {};
    function span() {
      if (open) out += "</span>";
      var cls = [], style = [];
      if (state.bold) cls.push("ansi-bold");
      if (state.dim) cls.push("ansi-dim");
      if (state.italic) cls.push("ansi-italic");
      if (state.underline) cls.push("ansi-underline");
      if (typeof state.fg === "number") cls.push("ansi-fg-" + state.fg);
      else if (state.fg) style.push("color:" + state.fg);
      if (typeof state.bg === "number") cls.push("ansi-bg-" + state.bg);
      else if (state.bg) style.push("background:" + state.bg);
      open = cls.length > 0 || style.length > 0;
      if (open) out += '<span class="' + cls.join(" ") + '"' + (style.length ? ' style="' + style.join(";") + '"' : "") + ">";
    }
    function color256(n) {
      if (n < 16) return n;
      if (n >= 232) { var g = 8 + (n - 232) * 10; return "rgb(" + g + "," + g + "," + g + ")"; }
      n -= 16;
      var level = function (v) { return v ? 55 + v * 40 : 0; };
      return "rgb(" + level(Math.floor(n / 36)) + "," + level(Math.floor(n / 6) % 6) + "," + level(n % 6) + ")";
    }
    var parts = text.split(/(\x1b\[[0-9;?]*[A-Za-z]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[()][A-Za-z0-9]|\x1b.)/);
    parts.forEach(function (part) {
      if (part.charAt(0) !== "\x1b") {
        out += escapeHTML(part.replace(/\r/g, ""));
        return;
      }
      var m = /^\x1b\[([0-9;]*)m$/.exec(part);
      if (!m) return;
      var codes = m[1] === "" ? [0] : m[1].split(";").map(Number);
      for (var i = 0; i < codes.length; i++) {
        var c = codes[i];
        if (c === 0) state = {};
        else if (c === 1) state.bold = true;
        else if (c === 2) state.dim = true;
        else if (c === 3) state.italic = true;
        else if (c === 4) state.underline = true;
        else if (c === 22) state.bold = state.dim = false;
        else if (c === 23) state.italic = false;
        else if (c === 24) state.underline = false;
        else if (c >= 30 && c <= 37) state.fg = c - 30;
        else if (c >= 90 && c <= 97) state.fg = c - 90 + 8;
        else if (c === 39) delete state.fg;
        else if (c >= 40 && c <= 47) state.bg = c - 40;
        else if (c >= 100 && c <= 107) state.bg = c - 100 + 8;
        else if (c === 49) delete state.bg;
        else if ((c === 38 || c === 48) && codes[i + 1] === 5) {
          state[c === 38 ? "fg" : "bg"] = color256(codes[i + 2]); i += 2;
        } else if ((c === 38 || c === 48) && codes[i + 1] === 2) {
          state[c === 38 ? "fg" : "bg"] = "rgb(" + codes[i + 2] + "," + codes[i + 3] + "," + codes[i + 4] + ")"; i += 4;
        }
      }
      span();
    });
    if (open) out += "</span>";
    return out;
  }

  // onRuns calls fn with each run of the runs topic, as it starts and as
  // it ends, over one event stream for the page.
  var runsHandlers = [];
  function onRuns(fn) {
    if (!runsHandlers.length) {
      var source = new EventSource("/api/sse/runs");
      source.addEventListener("run", function (event) {
        var run;
        try { run = JSON.parse(event.data); } catch (e) { return; }
        runsHandlers.forEach(function (handler) { handler(run); });
      });
    }
    runsHandlers.push(fn);
  }

  // The runs page follows the runs topic: a run that starts is added on
  // top of the table, one that ends is updated in place. The filters of
  // the page hold for the rows it adds; a page of older runs adds none.
  var runsTable = document.querySelector('[data-sse="runs"] [data-runs-table]');
  if (runsTable && window.EventSource) {
    var params = new URLSearchParams(location.search);
    var matches = function (run) {
      return (!params.get("folder") || params.get("folder") === run.folder) &&
        (!params.get("job") || run.job.indexOf(params.get("job")) >= 0) &&
        (!params.get("state") || params.get("state") === run.state) &&
        (!params.get("user") || params.get("user") === run.user_name);
    };
    var took = function (run) {
      if (!run.started_at) return "";
      var end = run.finished_at ? Date.parse(run.finished_at) : Date.now();
      var s = Math.max(0, (end - Date.parse(run.started_at)) / 1000);
      return s < 60 ? s.toFixed(1) + "s" : Math.floor(s / 60) + "m" + ("0" + Math.floor(s % 60)).slice(-2) + "s";
    };
    var stateHTML = function (run) {
      return '<span class="state"><i class="dot ' + escapeHTML(run.state) + '"></i>' + escapeHTML(run.state) +
        (run.state !== "ok" && ended(run.state) ? ' <span class="faint">(' + (run.exit_code || 0) + ")</span>" : "") + "</span>";
    };
    var actionsHTML = function (run, canRun) {
      var id = escapeHTML(run.id);
      var html = '<a class="btn btn-sm" href="/runs/' + id + '" data-attach="' + id + '">' + (!ended(run.state) ? "Attach" : "View") + "</a>";
      if (ended(run.state) && run.cast_path) {
        html += '<button type="button" class="btn btn-sm replay" data-replay="' + id + '" data-replay-title="' + escapeHTML(runTitle(run)) + '" data-replay-who="' + escapeHTML(run.user_name || "") +
          '" title="Replay the recording" aria-label="Replay the recording">|&gt;</button>';
      }
      if (!ended(run.state) && canRun) {
        html += '<button type="button" class="btn-icon" data-stop="' + id + '" title="Stop the run" aria-label="Stop the run"><svg class="ic"><use href="#i-stop"/></svg></button>' +
          '<button type="button" class="btn-icon danger" data-kill="' + id + '" title="Kill the run" aria-label="Kill the run"><svg class="ic"><use href="#i-trash"/></svg></button>';
      }
      if (ended(run.state) && runsTable.hasAttribute("data-summarize")) {
        html += '<button type="button" class="btn btn-sm" data-summary="/api/runs/' + id + '/summary" data-summary-page="/runs/' + id + '" data-summary-title="' + escapeHTML(runTitle(run)) +
          '" title="Claude\'s summary of the output">Summary</button>';
      }
      return html;
    };
    var whoHTML = function (run) {
      return '<span class="who-ran">' + userHTML(run.user_name) + (run.scheduled ? ' <span class="badge">scheduled</span>' : "") +
        (run.reason ? ' <span class="badge bad" title="' + escapeHTML(run.reason) + '">out of hours</span>' : run.forced ? ' <span class="badge signal">forced</span>' : "") + "</span>";
    };
    onRuns(function (run) {
      var row = runsTable.querySelector('[data-run-id="' + run.id + '"]');
      if (row) {
        if (!matches(run)) { row.remove(); return; }
        row.querySelector("[data-result]").innerHTML = stateHTML(run);
        row.querySelector("[data-took]").textContent = took(run);
        var canRun = !!row.querySelector("[data-stop]");
        row.querySelector("[data-actions]").innerHTML = actionsHTML(run, canRun);
        return;
      }
      if (!matches(run) || pagedBack()) return;
      var empty = runsTable.querySelector(".empty");
      if (empty) empty.remove();
      row = document.createElement("tr");
      row.dataset.runId = run.id;
      var folder = escapeHTML(run.folder), job = escapeHTML(run.job);
      var dir = run.dir ? escapeHTML(run.dir) : "";
      var place = folderHref(folder) + (dir ? "/" + dir : "");
      row.innerHTML =
        '<td class="nowrap" data-label="Started">just now</td>' +
        '<td data-label="Folder"><a href="' + place + '">' + escapeHTML(runPlace(run)) + "</a></td>" +
        '<td class="key"><a href="' + place + "/" + job + '">' + job + "</a></td>" +
        '<td data-label="By">' + whoHTML(run) + "</td>" +
        '<td class="nowrap" data-label="Took" data-took>' + took(run) + "</td>" +
        '<td data-label="Result" data-result>' + stateHTML(run) + "</td>" +
        '<td class="actions" data-actions>' + actionsHTML(run, false) + "</td>";
      runsTable.insertBefore(row, runsTable.firstChild);
      // Whether this user may end it comes with the run's own view.
      api("GET", "/api/runs/" + encodeURIComponent(run.id)).then(function (res) {
        if (res.status === 200 && !ended(res.data.state)) row.querySelector("[data-actions]").innerHTML = actionsHTML(res.data, res.data.can_run);
      });
    });
    // The running rows count their time.
    setInterval(function () {
      runsTable.querySelectorAll("tr[data-run-id]").forEach(function (row) {
        if (!row.querySelector(".dot.running")) return;
        var cell = row.querySelector("[data-took]");
        if (!cell.dataset.since) cell.dataset.since = Date.now() - (parseFloat(cell.textContent) || 0) * 1000;
        cell.textContent = took({ started_at: new Date(Number(cell.dataset.since)).toISOString() });
      });
    }, 1000);
  }

  // A job's control button in the dashboard's control centre follows
  // the runs topic: an eye on the job
  // running, Run again when it ended, in the colour of how it ended. Its
  // slot keeps that colour in data-last.
  function slotHTML(slot, run) {
    var label = slot.dataset.label ? "<span>" + escapeHTML(slot.dataset.label) + "</span>" : "";
    var named = slot.dataset.label ? " named" : "";
    if (run && run.state && run.state !== "running") slot.dataset.last = run.state === "ok" ? "ok" : "bad";
    var last = " " + escapeHTML(slot.dataset.last || "never");
    var what = escapeHTML(slot.dataset.folder + " / " + slot.dataset.name);
    if (run && run.state === "running") {
      return '<a class="cc' + named + ' running" href="/runs/' + escapeHTML(run.id) + '" data-attach="' + escapeHTML(run.id) +
        '" title="Attach to ' + what + ', running"><svg class="ic"><use href="#i-eye"/></svg>' + label + "</a>";
    }
    if (slot.hasAttribute("data-can")) {
      return '<button type="button" class="cc' + named + last + '" data-run data-folder="' + escapeHTML(slot.dataset.folder) + '" data-job="' +
        escapeHTML(slot.dataset.job) + '" data-title="' + what + '" title="Run ' + what + '"><svg class="ic"><use href="#i-play"/></svg>' + label + "</button>";
    }
    return '<a class="cc' + named + " idle" + last + '" href="' + escapeHTML(folderHref(slot.dataset.folder)) + "/" + escapeHTML(slot.dataset.job) + '" title="' + what +
      ': no permission to run it"><i class="shebang" aria-hidden="true">#!</i>' + label + "</a>";
  }
  if (document.querySelector(".cc-slot") && window.EventSource) {
    onRuns(function (run) {
      document.querySelectorAll('.cc-slot[data-cc="' + run.folder + "/" + run.job + '"]').forEach(function (slot) {
        slot.innerHTML = slotHTML(slot, run);
      });
    });
  }
  // The dot on a Files list row's Run follows the runs topic too: it
  // pulses while the job runs and takes how it ended.
  if (document.querySelector(".runsplit[data-key]") && window.EventSource) {
    onRuns(function (run) {
      var key = run.folder + (run.dir ? "/" + run.dir : "") + "/" + run.job;
      document.querySelectorAll(".runsplit[data-key]").forEach(function (split) {
        if (split.dataset.key !== key) return;
        var dot = split.querySelector(":scope > .btn .dot");
        if (dot) dot.className = "dot " + (run.state || "pending");
      });
    });
  }

  // The dashboard follows the runs topic too: a run that starts adds its
  // tile, polled for its frame; one that ends marks its tile and the last
  // frame is fetched. The running count follows.
  var dashboard = document.querySelector('[data-sse="runs"][data-dashboard]');
  if (dashboard && window.EventSource) {
    var tiles = dashboard.querySelector("[data-tiles]");
    var tilesEmpty = dashboard.querySelector("[data-tiles-empty]");
    var runningCount = dashboard.querySelector("[data-running-count]");
    var countRunning = function () {
      var n = tiles.querySelectorAll(".tile-head .dot.running").length;
      if (runningCount) runningCount.textContent = n;
      if (tilesEmpty) tilesEmpty.hidden = tiles.children.length > 0;
    };
    onRuns(function (run) {
      var tile = tiles.querySelector('[data-tile="' + run.id + '"]');
      if (!tile && run.state === "running") {
        tile = document.createElement("a");
        tile.className = "tile";
        tile.href = "/runs/" + run.id;
        tile.dataset.attach = run.id;
        tile.dataset.tile = run.id;
        tile.innerHTML = '<div class="tile-head"><i class="dot running"></i><b>' + escapeHTML(runTitle(run)) + "</b>" +
          '<span class="spacer"></span><span class="faint"></span>' +
          avatarHTML(run.user_name) + "</div>" +
          '<div class="frame" data-frame="' + escapeHTML(run.id) + '" role="img" aria-label="Terminal of ' + escapeHTML(runTitle(run)) + '"></div>';
        tiles.insertBefore(tile, tiles.firstChild);
        watchFrames();
      } else if (tile) {
        tile.querySelector(".tile-head .dot").className = "dot " + run.state;
        var frame = tile.querySelector(".frame");
        if (run.state !== "running" && frame.dataset.frame) {
          delete frame.dataset.frame;
          fetch(frameURL(run.id), { credentials: "same-origin" })
            .then(function (res) { return res.ok ? res.text() : null; })
            .then(function (svg) { if (svg) frame.innerHTML = svg; })
            .catch(function () {});
        }
      }
      countRunning();
    });
  }

  // A table with col-hover lights the column under the pointer, every
  // cell of it, so a job is easy to follow down the folders.
  document.querySelectorAll("table.col-hover").forEach(function (table) {
    var hot = -1;
    var light = function (index) {
      if (index === hot) return;
      hot = index;
      table.querySelectorAll(".col-hot").forEach(function (cell) { cell.classList.remove("col-hot"); });
      if (index < 0) return;
      Array.prototype.forEach.call(table.rows, function (row) {
        var cell = row.cells[index];
        if (cell && row.cells.length > 1) cell.classList.add("col-hot");
      });
    };
    table.addEventListener("mouseover", function (event) {
      var cell = event.target.closest("td, th");
      if (cell && table.contains(cell)) light(cell.cellIndex);
    });
    table.addEventListener("mouseleave", function () { light(-1); });
  });

  // A file picker fills the field it names, so a key can be chosen or pasted.
  document.addEventListener("change", function (event) {
    var input = event.target.closest("input[type=file][data-read-into]");
    if (!input || !input.files[0]) return;
    var target = document.getElementById(input.dataset.readInto);
    var reader = new FileReader();
    reader.onload = function () { target.value = reader.result; };
    reader.readAsText(input.files[0]);
  });

  // --------------------------------------------------------------- clicks

  document.addEventListener("click", function (event) {
    var btn = event.target.closest("[data-run], [data-attach], [data-replay], [data-stop], [data-kill], [data-delete], [data-copy]");
    if (!btn) return;
    // A link inside the element matched is a link: it navigates, it does
    // not run or attach for the element around it.
    var link = event.target.closest("a[href]");
    if (link && link !== btn && btn.contains(link)) return;

    if (btn.hasAttribute("data-copy")) {
      var value = btn.getAttribute("data-copy");
      // A path copies as a whole link, to paste anywhere.
      if (value.charAt(0) === "/") value = location.origin + value;
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(value).then(function () {
          btn.classList.add("copied");
          setTimeout(function () { btn.classList.remove("copied"); }, 1200);
        });
      } else {
        window.prompt("Copy:", value);
      }
      return;
    }

    if (btn.hasAttribute("data-delete")) {
      event.preventDefault();
      confirm(btn.dataset.confirm || "Delete?", btn.dataset.confirmOk).then(function (ok) {
        if (!ok) return;
        api("DELETE", btn.dataset.delete).then(function (res) {
          if (res.status >= 300) window.alert(res.data.error || ("HTTP " + res.status));
          else next(btn.dataset.next);
        });
      });
      return;
    }

    if (btn.hasAttribute("data-stop")) {
      event.preventDefault();
      endRun(btn.dataset.stop, "stop");
      return;
    }
    if (btn.hasAttribute("data-kill")) {
      event.preventDefault();
      confirm("Kill the run? It stops right away and shows as killed.", "Kill").then(function (ok) { if (ok) endRun(btn.dataset.kill, "kill"); });
      return;
    }

    if (!dialog || !window.Terminal) return; // the plain links still work
    event.preventDefault();

    if (btn.hasAttribute("data-attach")) {
      attach(btn.dataset.attach);
      return;
    }
    if (btn.hasAttribute("data-replay")) {
      replay(btn.dataset.replay, btn.dataset.replayTitle || "", btn.dataset.replayWho || "");
      return;
    }
    var spec = {
      folder: btn.dataset.folder,
      dir: btn.dataset.dir || "",
      job: btn.dataset.job,
      title: btn.dataset.title || (btn.dataset.folder + " / " + btn.dataset.job),
      force: btn.hasAttribute("data-force"),
      summarize: btn.hasAttribute("data-summarize"),
    };
    // A job that takes args opens a field for them first, under its
    // card or chips.
    if (btn.hasAttribute("data-args") && !spec.force) {
      argsLine(btn, spec);
      return;
    }
    // A Run that stops task-ui itself, compose up or down of its own
    // project, asks first.
    if (btn.dataset.confirm) {
      confirm(btn.dataset.confirm, btn.dataset.confirmOk || "Run").then(function (ok) { if (ok) start(spec); });
      return;
    }
    start(spec);
  });

  // ------------------------------------------------------------ args line

  // splitArgs reads a line as the server's argv.Split does, to show what
  // runs: blanks separate, quotes and backslashes group, nothing expands.
  // It returns the words, or null for a quote that does not close. The
  // server splits the line again and is the one that counts.
  function splitArgs(line) {
    var args = [], cur = "", started = false, quote = "", escaped = false;
    for (var i = 0; i < line.length; i++) {
      var c = line.charAt(i);
      if (escaped) {
        if (quote === '"' && c !== '"' && c !== "\\") cur += "\\";
        cur += c;
        escaped = false;
      } else if (quote === "'") {
        if (c === "'") quote = ""; else cur += c;
      } else if (c === "\\") {
        escaped = started = true;
      } else if (quote === '"') {
        if (c === '"') quote = ""; else cur += c;
      } else if (c === "'" || c === '"') {
        quote = c;
        started = true;
      } else if (c === " " || c === "\t") {
        if (started) { args.push(cur); cur = ""; started = false; }
      } else {
        cur += c;
        started = true;
      }
    }
    if (quote || escaped) return null;
    if (started) args.push(cur);
    return args;
  }

  // shownArg quotes a word that a shell would read otherwise, so the
  // preview reads as the argv it is.
  function shownArg(word) {
    if (/^[A-Za-z0-9_@%+=:,.\/-]+$/.test(word)) return word;
    return "'" + word.replace(/'/g, "'\\''") + "'";
  }

  // argsLine opens the field for the arguments of a job under its chips,
  // its card's head or the page's head: what the job runs, the words
  // typed after it as they reach it, and Run.
  function argsLine(btn, spec) {
    var anchor = btn.closest(".chips, .jobcard > header, .page-head, .runsplit") || btn.parentNode;
    var open = document.querySelector("form.argsline");
    if (open) {
      var same = open.dataset.job === spec.job && open.dataset.dir === spec.dir && open.dataset.folder === spec.folder;
      open.remove();
      if (same) return;
    }
    var cmd = btn.dataset.args;
    var form = document.createElement("form");
    form.className = "argsline";
    form.dataset.folder = spec.folder;
    form.dataset.dir = spec.dir;
    form.dataset.job = spec.job;
    form.innerHTML =
      '<div class="argsline-field"><code class="argsline-cmd"></code>' +
      '<input type="text" name="args" spellcheck="false" autocomplete="off" placeholder="arguments"></div>' +
      '<div class="argsline-foot"><span class="argsline-preview"></span>' +
      '<button type="button" class="btn btn-sm" data-args-cancel>Cancel</button>' +
      '<button type="submit" class="btn btn-primary btn-sm"><svg class="ic"><use href="#i-play"/></svg>Run</button></div>';
    form.querySelector(".argsline-cmd").textContent = cmd;
    var input = form.querySelector("input");
    input.setAttribute("aria-label", "Arguments for " + spec.title);
    var preview = form.querySelector(".argsline-preview");
    var show = function () {
      var words = splitArgs(input.value);
      preview.classList.toggle("bad", words === null);
      if (words === null) {
        preview.textContent = "A quote isn't closed.";
        return;
      }
      preview.innerHTML = "Runs <code></code>. Each word is one argument; variables, globs and pipes don't expand.";
      preview.querySelector("code").textContent = [cmd].concat(words.map(shownArg)).join(" ");
    };
    input.addEventListener("input", show);
    input.addEventListener("keydown", function (event) {
      if (event.key === "Escape") form.remove();
    });
    form.querySelector("[data-args-cancel]").addEventListener("click", function () { form.remove(); });
    form.addEventListener("submit", function (event) {
      event.preventDefault();
      if (splitArgs(input.value) === null) return;
      var run = Object.assign({}, spec, { args: input.value.trim() });
      form.remove();
      if (btn.dataset.confirm) {
        confirm(btn.dataset.confirm, btn.dataset.confirmOk || "Run").then(function (ok) { if (ok) start(run); });
        return;
      }
      start(run);
    });
    anchor.insertAdjacentElement("afterend", form);
    show();
    input.focus();
  }

  // ------------------------------------------------------------- output

  // The whole output of a finished run, flat: the recording replayed into
  // a terminal no one sees, with scrollback enough for every line, so what
  // redraws itself (a progress line, a pipeline tree) ends as it was last
  // drawn, and read back line by line with its colours into the page.
  function palette256(t) {
    var names = ["black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"];
    var p = names.map(function (n) { return t[n]; }).concat(names.map(function (n) {
      return t["bright" + n.charAt(0).toUpperCase() + n.slice(1)];
    }));
    var steps = [0, 95, 135, 175, 215, 255];
    for (var i = 0; i < 216; i++) {
      p.push("rgb(" + steps[Math.floor(i / 36)] + "," + steps[Math.floor(i / 6) % 6] + "," + steps[i % 6] + ")");
    }
    for (var g = 0; g < 24; g++) {
      var v = 8 + g * 10;
      p.push("rgb(" + v + "," + v + "," + v + ")");
    }
    return p;
  }

  function cellColor(cell, fg, pal) {
    if (fg ? cell.isFgDefault() : cell.isBgDefault()) return "";
    var c = fg ? cell.getFgColor() : cell.getBgColor();
    if (fg ? cell.isFgRGB() : cell.isBgRGB()) {
      return "#" + ("000000" + c.toString(16)).slice(-6);
    }
    return pal[c] || "";
  }

  function outputHTML(term, pal) {
    var buf = term.buffer.active;
    var cell = buf.getNullCell();
    var lines = [];
    for (var y = 0; y < buf.length; y++) {
      var line = buf.getLine(y);
      var html = "", run = "", style = null;
      var flush = function () {
        if (!run) return;
        html += style ? '<span style="' + style + '">' + escapeHTML(run) + "</span>" : escapeHTML(run);
        run = "";
      };
      for (var x = 0; x < line.length; x++) {
        line.getCell(x, cell);
        if (cell.getWidth() === 0) continue;
        var fg = cellColor(cell, true, pal), bg = cellColor(cell, false, pal);
        if (cell.isInverse()) { var tmp = fg; fg = bg || "var(--term-bg)"; bg = tmp || "var(--term-fg)"; }
        var st = (fg ? "color:" + fg + ";" : "") + (bg ? "background:" + bg + ";" : "") +
          (cell.isBold() ? "font-weight:700;" : "") + (cell.isDim() ? "opacity:.6;" : "") +
          (cell.isItalic() ? "font-style:italic;" : "") + (cell.isUnderline() ? "text-decoration:underline;" : "");
        st = st || null;
        if (st !== style) { flush(); style = st; }
        run += cell.getChars() || " ";
      }
      flush();
      lines.push(html.replace(/(\s|&nbsp;)+$/, ""));
    }
    while (lines.length && lines[lines.length - 1] === "") lines.pop();
    return lines.join("\n");
  }

  // readOutputs fills each data-output block under root with the output
  // of its run, read from the recording.
  function readOutputs(root) {
    Array.prototype.forEach.call(root.querySelectorAll("[data-output]"), readOutput);
  }
  readOutputs(document);

  function readOutput(pre) {
    if (!window.Terminal) return;
    var id = pre.dataset.output;
    fetch("/api/runs/" + encodeURIComponent(id) + "/cast", { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) throw new Error(res.status === 404 ? "no recording" : res.status === 410 ? "the recording expired under the Data retention settings" : "HTTP " + res.status);
        return res.text();
      })
      .then(function (text) {
        var out = [];
        text.split("\n").filter(Boolean).slice(1).forEach(function (line) {
          try {
            var ev = JSON.parse(line);
            if (ev[1] === "o") out.push(ev[2]);
          } catch (e) { /* torn last line */ }
        });
        var t = theme();
        var term = new window.Terminal({
          cols: parseInt(pre.dataset.cols, 10) || 120,
          rows: parseInt(pre.dataset.rows, 10) || 30,
          scrollback: 1000000,
          theme: t,
        });
        var host = document.createElement("div");
        host.style.cssText = "position:absolute;left:-10000px;top:0;width:10px;height:10px;overflow:hidden";
        document.body.appendChild(host);
        term.open(host);
        term.write(out.join(""), function () {
          pre.innerHTML = outputHTML(term, palette256(t)) || '<span class="faint">The run printed nothing.</span>';
          term.dispose();
          host.remove();
        });
      })
      .catch(function (err) {
        pre.innerHTML = '<span class="faint">Could not load the recording: ' + escapeHTML(err.message) + "</span>";
      });
  }

  // #run=<id> opens a run on load, so a window can be shared;
  // #start=<folder>/<job> starts one.
  (function openFromHash() {
    var match = /^#run=([0-9a-z]+)$/.exec(location.hash);
    if (match && dialog && window.Terminal) attach(match[1]);
    var start_ = /^#start=([^/]+)\/(.+)$/.exec(location.hash);
    if (start_ && dialog && window.Terminal) {
      history.replaceState(null, "", location.pathname + location.search);
      start({ folder: start_[1], job: start_[2], title: start_[1] + " / " + start_[2] });
    }
  })();

  // #args=<job>[;<line>] opens the arguments field of a job on the page,
  // the line typed in, so the field and what it runs can be shared.
  (function openArgsFromHash() {
    var match = /^#args=([^;]+)(?:;(.*))?$/.exec(location.hash);
    if (!match) return;
    var job = decodeURIComponent(match[1]);
    var btn = Array.prototype.find.call(document.querySelectorAll("[data-run][data-args]"), function (b) {
      return b.dataset.job === job;
    });
    if (!btn) return;
    btn.click();
    var input = document.querySelector("form.argsline input");
    if (input && match[2]) {
      input.value = decodeURIComponent(match[2]);
      input.dispatchEvent(new Event("input"));
    }
  })();

  // The sidebar is a drawer under the breakpoint. #menu opens it on load.
  var menuToggle = document.getElementById("menu-toggle");
  if (menuToggle) {
    var setMenu = function (open) {
      document.body.classList.toggle("menu-open", open);
      menuToggle.setAttribute("aria-expanded", open ? "true" : "false");
    };
    menuToggle.addEventListener("click", function () { setMenu(!document.body.classList.contains("menu-open")); });
    document.getElementById("menu-backdrop").addEventListener("click", function () { setMenu(false); });
    // Opened by the link it is simply open, without the slide: the page
    // is new, there is nothing to slide in from.
    if (location.hash === "#menu") {
      var side = document.getElementById("side");
      side.style.transition = "none";
      setMenu(true);
      requestAnimationFrame(function () { requestAnimationFrame(function () { side.style.transition = ""; }); });
    }
  }

  // The workspaces and projects in the sidebar fold on their chevron,
  // and the folded ones are remembered in this browser by their full
  // ID, so a workspace and a project in it fold apart. The open page's
  // project stays open.
  var FOLDED = "task-ui.folded";
  var folded = [];
  try { folded = JSON.parse(localStorage.getItem(FOLDED) || "[]"); } catch (e) { folded = []; }
  document.querySelectorAll(".folders details[data-project]").forEach(function (project) {
    var id = project.dataset.project;
    if (folded.indexOf(id) >= 0 && !project.classList.contains("active")) project.open = false;
    project.addEventListener("toggle", function () {
      var i = folded.indexOf(id);
      if (project.open && i >= 0) folded.splice(i, 1);
      else if (!project.open && i < 0) folded.push(id);
      try { localStorage.setItem(FOLDED, JSON.stringify(folded)); } catch (e) { /* no storage */ }
    });
  });

  // The user menu closes when the page is clicked elsewhere.
  var userMenu = document.getElementById("user-menu");
  if (userMenu) {
    document.addEventListener("click", function (event) {
      if (userMenu.open && !userMenu.contains(event.target)) userMenu.removeAttribute("open");
    });
  }

  // ------------------------------------------------------------------ live

  // The sidebar follows the runs started anywhere on the instance: the
  // recents, the dots on the folders, and the frames on the dashboard.
  var recents = document.getElementById("recents");
  var known = null;
  // recentsMore is the rail's last slot, as the layout draws it: shown in
  // place of the last entry that fits when some don't.
  var recentsMore = '<a href="/runs" class="recent more" data-recents-more hidden><svg class="ic"><use href="#i-list"/></svg><span class="what"><b>All runs</b><small></small></span></a>';

  // renderRecents draws the runs of jobs and of workflows, newest first.
  function renderRecents(runs, flows) {
    if (!recents) return;
    var entries = runs.map(function (r) { return { run: r, at: r.started_at || "" }; })
      .concat((flows || []).map(function (f) { return { flow: f, at: f.started_at || "" }; }));
    entries.sort(function (a, b) { return a.at < b.at ? 1 : a.at > b.at ? -1 : 0; });
    entries = entries.slice(0, 50);
    if (!entries.length) {
      recents.innerHTML = '<div class="empty">Runs started by anyone show up here.</div>';
      return;
    }
    var now = Date.now(), heading = "";
    recents.innerHTML = entries.map(function (e) {
      var label = hoursAgo(now - Date.parse(e.at));
      var head = label !== heading ? '<div class="recent-group">' + label + "</div>" : "";
      heading = label;
      return head + recentHTML(e);
    }).join("") + recentsMore;
    fitRecents();
  }

  // fitRecents hides the recents that run past the rail's column or the
  // bottom of the screen, whichever comes first, and a heading left with
  // no entry under it: the list is trimmed, not scrolled. When it hides
  // any, the last slot that fits links to Runs and says how many more
  // there are, so the rail never ends silently.
  function fitRecents() {
    if (!recents) return;
    var rail = recents.closest(".rail-inner");
    if (!rail || !rail.offsetParent) return;
    // Where the rail starts on the page, not on the screen, so the trim
    // holds wherever the page is scrolled to.
    var top = rail.getBoundingClientRect().top + window.scrollY;
    var limit = Math.min(rail.clientHeight, window.innerHeight - top);
    var more = recents.querySelector("[data-recents-more]");
    var items = Array.prototype.slice.call(recents.children).filter(function (item) { return item !== more; });
    var overflows = function (item) { return item.offsetTop + item.offsetHeight > limit; };
    if (more) more.hidden = true;
    items.forEach(function (item) { item.hidden = false; });
    items.forEach(function (item) {
      if (overflows(item)) item.hidden = true;
    });
    var hidden = items.filter(function (item) { return item.hidden && item.classList.contains("recent"); });
    if (more && hidden.length) {
      more.hidden = false;
      // The link takes the place of the last entries that fit until it
      // fits itself.
      var shown = items.filter(function (item) { return !item.hidden; });
      while (shown.length && overflows(more)) shown.pop().hidden = true;
      var count = items.filter(function (item) { return item.hidden && item.classList.contains("recent"); }).length;
      more.querySelector("small").textContent = count + " more";
    }
    for (var i = items.length - 1; i >= 0; i--) {
      var item = items[i];
      if (!item.classList.contains("recent-group") || item.hidden) continue;
      var next = item.nextElementSibling;
      if (!next || next.hidden || next === more || next.classList.contains("recent-group")) item.hidden = true;
    }
  }
  fitRecents();
  window.addEventListener("resize", fitRecents);

  // hoursAgo is the heading of a recent that started a while ago, in
  // milliseconds: whole hours, then whole days, as the page draws them.
  function hoursAgo(ms) {
    var hours = Math.floor((ms || 0) / 3600000);
    if (hours < 1) return "In the last hour";
    if (hours < 24) return (hours === 1 ? "1 hour" : hours + " hours") + " ago";
    var days = Math.floor(hours / 24);
    return (days === 1 ? "1 day" : days + " days") + " ago";
  }

  // recentHTML draws one entry of the recents: a workflow run, or a run.
  function recentHTML(e) {
    if (e.flow) {
      var f = e.flow;
      return '<a href="/workflows/' + encodeURIComponent(f.workflow_id) + '" class="recent flow" title="' +
        escapeHTML("Workflow " + f.name + ", " + f.user_name + ", " + f.state) + '">' +
        '<i class="dot ' + escapeHTML(f.state) + '"></i>' +
        '<span class="what"><b>' + escapeHTML(f.name) + '</b><small><svg class="ic"><use href="#i-flow"/></svg>workflow</small></span>' +
        (f.state === "running" ? '<svg class="ic act"><use href="#i-eye"/></svg>' : "") +
        avatarHTML(f.user_name) + "</a>";
    }
    var r = e.run;
    var running = r.state === "running";
    return '<a href="/runs/' + encodeURIComponent(r.id) + '" class="recent" title="' +
      escapeHTML("Open " + runTitle(r) + ", " + r.user_name + ", " + r.state) + '">' +
      '<i class="dot ' + escapeHTML(r.state) + '"></i>' +
      '<span class="what"><b>' + escapeHTML(r.job + (r.target ? " · " + r.target : "")) + '</b><small>' + escapeHTML(runPlace(r)) + '</small></span>' +
      (running ? '<svg class="ic act"><use href="#i-' + (r.can_run ? "play" : "eye") + '"/></svg>' : "") +
      avatarHTML(r.user_name) + "</a>";
  }

  // isPromptRun reports whether a run is a prompt's: @claude-prompt, or
  // internal/claude-prompt of an older thread.
  function isPromptRun(r) {
    return r.job === "@claude-prompt" || (r.folder === "internal" && r.job === "claude-prompt");
  }

  function poll() {
    if (document.hidden) return;
    Promise.all([api("GET", "/api/runs?limit=50"), api("GET", "/api/workflow-runs?limit=50")]).then(function (both) {
      var res = both[0], flows = both[1];
      if (res.status !== 200 || !res.data.runs) return;
      var runs = res.data.runs;
      renderRecents(runs, flows.status === 200 ? flows.data.runs : []);
      var running = {}, runningJobs = {};
      runs.forEach(function (r) {
        if (r.state !== "running") return;
        running[r.folder] = true;
        runningJobs[r.folder + (r.dir ? "/" + r.dir : "") + "/" + r.job] = true;
      });
      document.querySelectorAll(".side [data-folder] > summary .dot").forEach(function (dot) {
        dot.classList.toggle("running", !!running[dot.closest("[data-folder]").dataset.folder]);
      });
      document.querySelectorAll(".side [data-job-key] .shebang").forEach(function (dot) {
        dot.classList.toggle("running", !!runningJobs[dot.parentElement.dataset.jobKey]);
      });
      var any = document.querySelector("[data-any-running]");
      if (any) any.classList.toggle("running", Object.keys(running).length > 0);
      // A prompt's run changes no page: its thread follows it, in the
      // header's panel or under a prompt box.
      var signature = runs.filter(function (r) { return !isPromptRun(r); }).map(function (r) { return r.id + ":" + r.state + ":" + r.attached; }).join(",");
      // Pages that list runs refresh when the set changes, unless a window
      // or the header's panel is open: a window refreshes the page when it
      // closes.
      if (known !== null && signature !== known && !session && !headerOpen() && !document.querySelector("dialog[open]") && document.querySelector("[data-live]")) {
        window.location.reload();
      }
      known = signature;
    }).catch(function () {});
  }

  // frameURL is where a run's frame is drawn: its last screen, or with
  // all its whole output so far.
  function frameURL(runId, all) {
    return "/api/runs/" + encodeURIComponent(runId) + "/frame.svg" + (all ? "?all=1" : "");
  }

  // A job card's "show N more" is a details element, open without a
  // script; its summary says aloud whether the lines past the third show.
  function syncMore(details) {
    details.querySelector(":scope > summary").setAttribute("aria-expanded", details.open ? "true" : "false");
  }
  document.querySelectorAll(".jobcard .code details").forEach(syncMore);
  document.addEventListener("toggle", function (event) {
    if (event.target.matches && event.target.matches(".jobcard .code details")) syncMore(event.target);
  }, true);

  // Expand on a step's row asks the API for the run's whole output and
  // shows it in place of the last screen; Collapse puts back the screen
  // it kept, with no request.
  document.addEventListener("click", function (event) {
    var btn = event.target.closest("[data-expand]");
    if (!btn) return;
    event.preventDefault();
    var frame = btn.closest(".what").querySelector(":scope > .frame");
    if (!frame) return;
    var open = btn.getAttribute("aria-expanded") !== "true";
    frame.classList.toggle("expanded", open);
    btn.setAttribute("aria-expanded", open ? "true" : "false");
    // The label keeps the line count it was drawn with: "Expand 120 lines".
    var label = btn.querySelector("span");
    if (!btn.dataset.label) btn.dataset.label = label ? label.textContent : "Expand";
    btn.innerHTML = "";
    btn.appendChild(icon(open ? "i-collapse" : "i-expand"));
    btn.appendChild(el("span", "", open ? "Collapse" : btn.dataset.label));
    if (!open) {
      delete frame.dataset.all;
      if (frame.screen !== undefined) frame.innerHTML = frame.screen;
      return;
    }
    frame.screen = frame.innerHTML;
    frame.dataset.all = "1";
    fetch(frameURL(btn.dataset.expand, true), { credentials: "same-origin" })
      .then(function (res) { return res.ok ? res.text() : null; })
      .then(function (svg) { if (svg && frame.dataset.all) frame.innerHTML = svg; })
      .catch(function () {});
  });

  // The frames are inline SVG, so they use the page's own faces; each
  // one is fetched again and swapped in place.
  function refreshFrames() {
    if (document.hidden) return;
    document.querySelectorAll(".frame[data-frame]").forEach(function (frame) {
      fetch(frameURL(frame.dataset.frame, frame.dataset.all), { credentials: "same-origin" })
        .then(function (res) { return res.ok ? res.text() : null; })
        .then(function (svg) { if (svg) frame.innerHTML = svg; })
        .catch(function () {});
    });
  }

  // watchFrames starts polling the frames once there is one to poll.
  var framesTimer = null;
  function watchFrames() {
    if (!framesTimer && document.querySelector(".frame[data-frame]")) framesTimer = setInterval(refreshFrames, 2000);
  }

  if (document.querySelector(".side")) {
    poll();
    setInterval(poll, 4000);
    watchFrames();
  }
})();
