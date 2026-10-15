// task-ui: run dialog, history playback and live run status.
//
// The run dialog owns the run: the socket is opened when the dialog opens,
// and closing the dialog closes the socket, which stops the process on the
// server. Binary frames are terminal bytes, text frames are JSON status.
(function () {
  "use strict";

  var history_ = window.history;

  var dialog = document.getElementById("run-dialog");
  var els = dialog && {
    kind: document.getElementById("run-kind"),
    title: document.getElementById("run-title"),
    state: document.getElementById("run-state"),
    elapsed: document.getElementById("run-elapsed"),
    host: document.getElementById("run-terminal"),
    notice: document.getElementById("run-notice"),
    hint: document.getElementById("run-hint"),
    force: document.getElementById("run-force"),
    stop: document.getElementById("run-stop"),
    again: document.getElementById("run-again"),
    close: document.getElementById("run-close"),
  };

  var session = null; // { ws, term, fit, timer, spec, finished, changed }

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
      // White maps to dark ink in the light theme: tools print headings in
      // bold white, which has to stay readable on paper.
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

  function setState(state, text) {
    els.state.innerHTML = "";
    if (!state) return;
    var badge = document.createElement("span");
    var cls = { running: "signal", ok: "ok", failed: "bad", killed: "bad", error: "bad", denied: "bad", busy: "signal", stopped: "" }[state];
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

  function teardown() {
    if (!session) return;
    stopClock();
    if (session.ws && session.ws.readyState <= 1) session.ws.close();
    if (session.term) session.term.dispose();
    if (session.onResize) window.removeEventListener("resize", session.onResize);
    var changed = session.changed;
    session = null;
    if (/^#(history|replay)=/.test(location.hash)) {
      history_.replaceState(null, "", location.pathname + location.search);
    }
    // A run changes the history and the run state on the page.
    if (changed) window.location.reload();
  }

  function buttons(opts) {
    els.stop.hidden = !opts.stop;
    els.force.hidden = !opts.force;
    els.again.hidden = !opts.again;
  }

  function open(kind, title) {
    if (session) teardown();
    els.kind.textContent = kind;
    els.title.innerHTML = "";
    var project = document.createElement("span");
    project.textContent = title.project + " / ";
    els.title.appendChild(project);
    els.title.appendChild(document.createTextNode(title.name));
    els.elapsed.textContent = "";
    setState(null);
    notice("");
    if (!dialog.open) dialog.showModal();
    session = newTerminal();
    session.term.focus();
    return session;
  }

  // run opens the dialog and runs a runbook.
  function run(spec) {
    var s = open("Run", { project: spec.title, name: spec.runbook });
    s.spec = spec;
    els.hint.textContent = "Closing this window stops the run.";
    buttons({ stop: true });
    setState("running", "connecting");

    var term = s.term;
    var proto = location.protocol === "https:" ? "wss:" : "ws:";
    var query = "?cols=" + term.cols + "&rows=" + term.rows + (spec.force ? "&force=1" : "");
    var url = proto + "//" + location.host + "/ws/run/" + encodeURIComponent(spec.project) + "/" + encodeURIComponent(spec.runbook) + query;
    var ws = new WebSocket(url);
    ws.binaryType = "arraybuffer";
    s.ws = ws;

    var encoder = new TextEncoder();
    term.onData(function (data) {
      if (spec.interactive && ws.readyState === 1) ws.send(encoder.encode(data));
    });

    s.onResize = function () {
      try { s.fit.fit(); } catch (e) { return; }
      if (ws.readyState === 1) ws.send(JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows }));
    };
    window.addEventListener("resize", s.onResize);

    ws.onmessage = function (event) {
      if (typeof event.data !== "string") {
        term.write(new Uint8Array(event.data));
        return;
      }
      var msg;
      try { msg = JSON.parse(event.data); } catch (e) { return; }
      if (msg.type !== "status") return;
      status(s, msg);
    };

    ws.onclose = function () {
      if (session !== s || s.finished) return;
      s.finished = true;
      stopClock();
      setState("error", "disconnected");
      notice("The connection closed before the run reported a result.", "bad");
      buttons({ again: true });
    };
  }

  function status(s, msg) {
    switch (msg.state) {
      case "running":
        s.changed = true;
        setState("running");
        startClock(msg.started ? Date.parse(msg.started) : Date.now());
        notice(spec(s).force ? "Started with Run force." : "");
        break;
      case "busy":
        s.finished = true;
        setState("busy", "already running");
        notice(msg.message, "signal");
        buttons({ force: true });
        break;
      case "denied":
      case "error":
        s.finished = true;
        setState(msg.state);
        notice(msg.message, "bad");
        buttons({});
        break;
      default: // ok, failed, stopped, killed
        s.finished = true;
        s.changed = true;
        stopClock();
        s.term.write("\x1b[?25l");
        setState(msg.state, msg.state === "ok" ? "ok" : msg.state + (msg.exit !== undefined ? " " + msg.exit : ""));
        if (msg.state === "killed") notice("This run was stopped from another window, or started again with Run force.", "bad");
        else if (msg.state === "failed") notice("The task exited with status " + msg.exit + ".", "bad");
        else notice("");
        els.hint.textContent = "Finished. The run is in the history.";
        buttons({ again: true });
    }
  }

  function spec(s) {
    return s.spec || {};
  }

  // history plays a recording in the dialog. View writes it out at once,
  // Replay keeps the recorded timing, with long pauses shortened.
  function history(btn) {
    var s = open(btn.hasAttribute("data-replay") ? "Replay" : "History", { project: btn.dataset.title, name: btn.dataset.runbook });
    els.hint.textContent = "Recorded " + (btn.dataset.when || "") + ".";
    els.elapsed.textContent = "";
    buttons({});
    setState("running", "loading");
    var replay = btn.hasAttribute("data-replay");
    // The open recording is in the address, so the link can be shared.
    var id = btn.dataset.history.split("/").pop();
    history_.replaceState(null, "", "#" + (replay ? "replay" : "history") + "=" + id);

    fetch(btn.dataset.history, { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) throw new Error("HTTP " + res.status);
        return res.text();
      })
      .then(function (text) {
        if (session !== s) return;
        var lines = text.split("\n").filter(Boolean);
        var events = [];
        for (var i = 1; i < lines.length; i++) {
          try {
            var ev = JSON.parse(lines[i]);
            if (ev[1] === "o") events.push(ev);
          } catch (e) { /* torn last line */ }
        }
        if (!replay) {
          s.term.write(events.map(function (ev) { return ev[2]; }).join("") + "\x1b[?25l");
          setState("ok", "recorded");
          return;
        }
        setState("running", "replaying");
        var idx = 0, clock = 0;
        var step = function () {
          if (session !== s) return;
          if (idx >= events.length) {
            s.term.write("\x1b[?25l");
            setState("ok", "done");
            return;
          }
          var ev = events[idx++];
          var wait = Math.min(Math.max(ev[0] - clock, 0), 1.0);
          clock = ev[0];
          s.replay = setTimeout(function () {
            if (session !== s) return;
            s.term.write(ev[2]);
            els.elapsed.textContent = elapsedText(ev[0] * 1000);
            step();
          }, wait * 1000);
        };
        step();
      })
      .catch(function (err) {
        if (session !== s) return;
        setState("error");
        notice("Could not load the recording: " + err.message, "bad");
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
    els.force.addEventListener("click", function () {
      var spec = session && session.spec;
      if (spec) run(Object.assign({}, spec, { force: true }));
    });
    els.again.addEventListener("click", function () {
      var spec = session && session.spec;
      if (spec) {
        session.changed = false;
        run(Object.assign({}, spec, { force: false }));
      }
    });
  }

  document.addEventListener("click", function (event) {
    var btn = event.target.closest("[data-run], [data-history], [data-copy]");
    if (!btn) return;

    if (btn.hasAttribute("data-copy")) {
      var value = btn.getAttribute("data-copy");
      var done = function () {
        var label = btn.textContent;
        btn.textContent = "Copied";
        setTimeout(function () { btn.textContent = label; }, 1200);
      };
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(value).then(done);
      } else {
        window.prompt("Copy the public key:", value);
      }
      return;
    }

    if (!dialog || !window.Terminal) return;
    event.preventDefault();

    if (btn.hasAttribute("data-history")) {
      history(btn);
      return;
    }
    run({
      project: btn.dataset.project,
      runbook: btn.dataset.runbook,
      title: btn.dataset.title || btn.dataset.project,
      interactive: btn.dataset.interactive === "true",
      force: btn.hasAttribute("data-force"),
    });
  });

  // #history=<id> and #replay=<id> open a recording on load.
  function openFromHash() {
    var match = /^#(history|replay)=([0-9A-Za-z-]+)$/.exec(location.hash);
    if (!match || !dialog || !window.Terminal) return;
    var selector = "[data-history$='/" + match[2] + "']" + (match[1] === "replay" ? "[data-replay]" : ":not([data-replay])");
    var btn = document.querySelector(selector);
    if (btn) history(btn);
  }
  openFromHash();

  document.addEventListener("submit", function (event) {
    var form = event.target.closest("form[data-confirm]");
    if (form && !window.confirm(form.getAttribute("data-confirm"))) event.preventDefault();
  });

  // Live run status: the sidebar dots follow the runs started anywhere on
  // the instance. Polling is cheap, the list holds active runs only.
  var known = null;
  function poll() {
    if (document.hidden) return;
    fetch("/api/runs", { credentials: "same-origin", headers: { Accept: "application/json" } })
      .then(function (res) { return res.ok ? res.json() : null; })
      .then(function (runs) {
        if (!runs) return;
        var keys = {};
        runs.forEach(function (r) { keys[r.key] = r; });
        document.querySelectorAll(".side [data-key] .dot").forEach(function (dot) {
          dot.classList.toggle("running", !!keys[dot.parentElement.dataset.key]);
        });
        var signature = Object.keys(keys).sort().join(",");
        // Refresh pages that list runs when the set changes, unless a
        // dialog is open: that run reloads the page when it closes.
        if (known !== null && signature !== known && !session && document.querySelector("[data-live]")) {
          window.location.reload();
        }
        known = signature;
      })
      .catch(function () {});
  }
  if (document.querySelector(".side")) {
    poll();
    setInterval(poll, 4000);
  }
})();
