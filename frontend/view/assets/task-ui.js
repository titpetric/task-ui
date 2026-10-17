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
    var words = String(name || "?").trim().split(/[\s._-]+/).filter(Boolean);
    var text = words.length >= 2 ? words[0].charAt(0) + words[1].charAt(0) : (words[0] || "?").slice(0, 2);
    return text.toUpperCase();
  }

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
  }

  // teardown closes the window: the socket and the terminal go, the run
  // does not. A page that showed a run that changed is refreshed.
  function teardown() {
    if (!session) return;
    stopClock();
    if (session.ws && session.ws.readyState <= 1) session.ws.close();
    if (session.term) session.term.dispose();
    if (session.onResize) window.removeEventListener("resize", session.onResize);
    var changed = session.changed;
    session = null;
    if (/^#run=/.test(location.hash)) history.replaceState(null, "", location.pathname + location.search);
    if (changed) window.location.reload();
  }

  function open(kind, title, who) {
    if (session) teardown();
    els.kind.textContent = kind;
    els.title.textContent = title;
    els.who.textContent = who || "";
    els.elapsed.textContent = "";
    els.hint.textContent = "Closing this window leaves the run running.";
    setState(null);
    notice("");
    buttons({});
    if (!dialog.open) dialog.showModal();
    session = newTerminal();
    session.term.focus();
    return session;
  }

  // ------------------------------------------------------------------ runs

  // start runs a job: the dialog opens first so the terminal size is known,
  // then the run is started and attached to.
  function start(spec) {
    var s = open("Run", spec.title, "");
    s.spec = spec;
    setState("running", "starting");
    var body = { folder: spec.folder, job: spec.job, force: !!spec.force, cols: s.term.cols, rows: s.term.rows };
    api("POST", "/api/runs", body).then(function (res) {
      if (session !== s) return;
      if (res.status === 201) {
        s.changed = true;
        connect(s, res.data);
        return;
      }
      if (res.status === 409 && res.data.run) {
        s.run = res.data.run;
        setState("busy", "already running");
        notice("Already running, started by " + (res.data.run.user_name || "someone") + ". Attach to follow it, or Run force to kill it and start again.", "signal");
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
      els.title.textContent = run.folder + " / " + run.job;
      els.who.textContent = run.user_name || "";
      s.spec = { folder: run.folder, job: run.job, title: run.folder + " / " + run.job };
      if (run.live) {
        connect(s, run);
      } else {
        playback(s, run);
      }
    });
  }

  // connect attaches the terminal to a live run over the socket.
  function connect(s, run) {
    s.run = run;
    els.title.textContent = run.folder + " / " + run.job;
    els.who.textContent = run.user_name || "";
    var term = s.term;
    var proto = location.protocol === "https:" ? "wss:" : "ws:";
    var ws = new WebSocket(proto + "//" + location.host + "/api/runs/" + encodeURIComponent(run.id) + "/attach");
    ws.binaryType = "arraybuffer";
    s.ws = ws;
    setState("running", "attaching");
    buttons({ stop: run.can_run, kill: run.can_run });
    els.hint.textContent = run.can_run
      ? "Closing this window leaves the run running. Stop ends it, Kill ends it now."
      : "You are watching: your role cannot type into it or end it. Closing this window leaves it running.";

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
      notice("The connection closed. The run goes on; attach again to follow it.", "bad");
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
    if (msg.state === "killed") notice("Killed: ended from another window, by Run force, or by the server.", "bad");
    else if (msg.state === "failed") notice(msg.message || ("The job exited with status " + msg.exit_code + "."), "bad");
    else if (msg.state === "stopped") notice("Stopped.", "");
    else notice("");
    els.hint.textContent = "Finished. The run and its recording are in the history.";
    buttons({ again: s.run && s.run.can_run });
  }

  // playback writes the recording of a finished run into the terminal.
  function playback(s, run) {
    els.kind.textContent = "History";
    els.hint.textContent = "Recorded " + (run.started_at ? new Date(run.started_at).toLocaleString() : "") + ".";
    buttons({ again: run.can_run });
    fetch("/api/runs/" + encodeURIComponent(run.id) + "/cast", { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) throw new Error(res.status === 404 ? "no recording" : "HTTP " + res.status);
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
      })
      .catch(function (err) {
        if (session !== s) return;
        setState(run.state);
        notice("Could not load the recording: " + err.message, "bad");
      });
  }

  function endRun(id, how) {
    return api("POST", "/api/runs/" + encodeURIComponent(id) + "/" + how).then(function (res) {
      if (res.status >= 300) window.alert(res.data.error || ("HTTP " + res.status));
      else if (!session) window.location.reload();
    });
  }

  function runAll(folder, force) {
    api("POST", "/api/runs/all", { folder: folder, force: !!force }).then(function (res) {
      if (res.status >= 300) {
        window.alert(res.data.error || ("HTTP " + res.status));
        return;
      }
      window.location.reload();
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
      if (spec) start(Object.assign({}, spec, { force: true }));
    });
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

  // serialize reads a form into the JSON the API takes. data-type says
  // what a field is: a list splits on commas, a bool parses, an int too.
  // A field named target holds "folder/job" and is sent as both.
  function serialize(form) {
    var body = {};
    Array.prototype.forEach.call(form.elements, function (el) {
      if (!el.name || el.disabled) return;
      var value = el.value;
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
      }
      if (el.name === "target" && typeof value === "string") {
        var at = value.indexOf("/");
        body.folder = value.slice(0, at);
        body.job = value.slice(at + 1);
        return;
      }
      body[el.name] = value;
    });
    return body;
  }

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

  function submit(form) {
    var method = (form.dataset.method || "POST").toUpperCase();
    var button = form.querySelector("[type=submit]");
    if (button) button.disabled = true;
    formError(form, "");
    api(method, form.dataset.api, serialize(form)).then(function (res) {
      if (res.status >= 300) {
        if (button) button.disabled = false;
        formError(form, (res.data.error || ("HTTP " + res.status)) + (res.data.output ? "\n\n" + res.data.output : ""));
        return;
      }
      next(form.dataset.next);
    }).catch(function (err) {
      if (button) button.disabled = false;
      formError(form, err.message);
    });
  }

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
    var btn = event.target.closest("[data-run], [data-run-all], [data-attach], [data-stop], [data-kill], [data-delete], [data-copy]");
    if (!btn) return;

    if (btn.hasAttribute("data-copy")) {
      var value = btn.getAttribute("data-copy");
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(value).then(function () {
          btn.classList.add("copied");
          setTimeout(function () { btn.classList.remove("copied"); }, 1200);
        });
      } else {
        window.prompt("Copy the public key:", value);
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
      confirm("Kill the run? It ends now, as killed.", "Kill").then(function (ok) { if (ok) endRun(btn.dataset.kill, "kill"); });
      return;
    }

    if (btn.hasAttribute("data-run-all")) {
      event.preventDefault();
      runAll(btn.dataset.folder, btn.hasAttribute("data-force"));
      return;
    }

    if (!dialog || !window.Terminal) return; // the plain links still work
    event.preventDefault();

    if (btn.hasAttribute("data-attach")) {
      attach(btn.dataset.attach);
      return;
    }
    start({
      folder: btn.dataset.folder,
      job: btn.dataset.job,
      title: btn.dataset.title || (btn.dataset.folder + " / " + btn.dataset.job),
      force: btn.hasAttribute("data-force"),
    });
  });

  // #run=<id> opens a run on load, so a window can be shared.
  (function openFromHash() {
    var match = /^#run=([0-9a-z]+)$/.exec(location.hash);
    if (match && dialog && window.Terminal) attach(match[1]);
  })();

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

  function renderRecents(runs) {
    if (!recents) return;
    if (!runs.length) {
      recents.innerHTML = '<div class="empty">Runs started by anyone show up here.</div>';
      return;
    }
    recents.innerHTML = runs.map(function (r) {
      return '<a href="/runs/' + encodeURIComponent(r.id) + '" data-attach="' + escapeHTML(r.id) + '" class="recent" title="' +
        escapeHTML(r.folder + " / " + r.job + ", " + r.user_name + ", " + r.state) + '">' +
        '<i class="dot ' + escapeHTML(r.state) + '"></i>' +
        '<span class="what"><b>' + escapeHTML(r.job) + '</b><small>' + escapeHTML(r.folder) + '</small></span>' +
        '<span class="avatar sm" title="' + escapeHTML(r.user_name) + '">' + escapeHTML(initials(r.user_name)) + '</span></a>';
    }).join("");
  }

  function poll() {
    if (document.hidden) return;
    api("GET", "/api/runs?limit=10").then(function (res) {
      if (res.status !== 200 || !res.data.runs) return;
      var runs = res.data.runs;
      renderRecents(runs);
      var running = {};
      runs.forEach(function (r) { if (r.state === "running") running[r.folder] = true; });
      document.querySelectorAll(".side [data-folder] .dot").forEach(function (dot) {
        dot.classList.toggle("running", !!running[dot.parentElement.dataset.folder]);
      });
      var any = document.querySelector("[data-any-running]");
      if (any) any.classList.toggle("running", Object.keys(running).length > 0);
      var signature = runs.map(function (r) { return r.id + ":" + r.state + ":" + r.attached; }).join(",");
      // Pages that list runs refresh when the set changes, unless a window
      // is open: that window refreshes the page when it closes.
      if (known !== null && signature !== known && !session && document.querySelector("[data-live]")) {
        window.location.reload();
      }
      known = signature;
    }).catch(function () {});
  }

  function refreshFrames() {
    if (document.hidden) return;
    document.querySelectorAll("img[data-frame]").forEach(function (img) {
      img.src = "/api/runs/" + encodeURIComponent(img.dataset.frame) + "/frame.svg?t=" + Date.now();
    });
  }

  if (document.querySelector(".side")) {
    poll();
    setInterval(poll, 4000);
    if (document.querySelector("img[data-frame]")) setInterval(refreshFrames, 2000);
  }
})();
