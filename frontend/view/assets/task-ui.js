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
    reason: document.getElementById("run-reason"),
    reasonText: document.getElementById("run-reason-text"),
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
    els.reason.hidden = true;
    els.reasonText.value = "";
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
    var body = { folder: spec.folder, job: spec.job, force: !!spec.force, reason: spec.reason || "", cols: s.term.cols, rows: s.term.rows };
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
        notice("Out of hours: " + spec.title + " runs " + (res.data.hours || "never").replace(/\n/g, ", ") + ". Run force needs a reason.", "signal");
        els.hint.textContent = "The reason is recorded with the run, for everyone to see.";
        els.reason.hidden = false;
        els.reasonText.focus();
        buttons({ force: true });
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
      // A finished run stays live in memory for a while, for the windows
      // on it; opened afterwards it is history, not something to attach to.
      if (run.live && run.state === "running") {
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
    s.run = run;
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
        notice("Say why this has to run out of hours.", "bad");
        els.reasonText.focus();
        return;
      }
      start(Object.assign({}, spec, { force: true, reason: reason }));
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
    var body = serialize(form);
    if (form.hasAttribute("data-workflow")) body.steps = workflowSteps(form);
    api(method, form.dataset.api, body).then(function (res) {
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

  // A select that narrows another keeps only the options of its folder:
  // the job list follows the folder picked for a schedule.
  function narrow(select) {
    var target = document.getElementById(select.dataset.narrows);
    if (!target) return;
    if (!target.allOptions) target.allOptions = Array.prototype.slice.call(target.options);
    target.innerHTML = "";
    target.allOptions.forEach(function (option) {
      if (option.dataset.folder === select.value) target.appendChild(option);
    });
  }
  document.querySelectorAll("select[data-narrows]").forEach(function (select) {
    narrow(select);
    select.addEventListener("change", function () { narrow(select); });
  });

  // ------------------------------------------------------------- workflows

  // A step's folder select also lists the workflows, as workflow:<id>.
  var WORKFLOW = "workflow:";

  // workflowSteps reads a workflow form's steps, in their order: a job of
  // a folder, or a workflow.
  function workflowSteps(form) {
    return Array.prototype.map.call(form.querySelectorAll("[data-step]"), function (row) {
      var folder = row.querySelector("[data-step-folder]").value;
      if (folder.indexOf(WORKFLOW) === 0) return { workflow: folder.slice(WORKFLOW.length) };
      return { folder: folder, job: row.querySelector("[data-step-job]").value };
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
    if (flow) return;
    if (!job.allOptions) job.allOptions = Array.prototype.slice.call(job.options);
    var picked = job.value;
    job.innerHTML = "";
    job.allOptions.forEach(function (option) {
      if (option.dataset.folder === folder.value) job.appendChild(option);
    });
    if (picked && Array.prototype.some.call(job.options, function (o) { return o.value === picked; })) job.value = picked;
  }

  // stepTools enables the move buttons that have somewhere to go, and
  // keeps one step at least.
  function stepTools(list) {
    var rows = list.querySelectorAll("[data-step]");
    Array.prototype.forEach.call(rows, function (row, i) {
      row.querySelector("[data-step-up]").disabled = i === 0;
      row.querySelector("[data-step-down]").disabled = i === rows.length - 1;
      row.querySelector("[data-step-remove]").disabled = rows.length === 1;
    });
  }

  document.querySelectorAll("form[data-workflow]").forEach(function (form) {
    var list = form.querySelector("[data-steps]");
    list.querySelectorAll("[data-step]").forEach(narrowStep);
    stepTools(list);
    form.addEventListener("change", function (event) {
      if (event.target.matches("[data-step-folder]")) narrowStep(event.target.closest("[data-step]"));
    });
    form.addEventListener("click", function (event) {
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
        narrowStep(fresh);
        fresh.querySelector("[data-step-job]").focus();
      } else if (btn.hasAttribute("data-step-up") && row.previousElementSibling) {
        list.insertBefore(row, row.previousElementSibling);
      } else if (btn.hasAttribute("data-step-down") && row.nextElementSibling) {
        list.insertBefore(row.nextElementSibling, row);
      } else if (btn.hasAttribute("data-step-remove") && list.children.length > 1) {
        row.remove();
      }
      stepTools(list);
    });
  });

  document.addEventListener("click", function (event) {
    var btn = event.target.closest("[data-workflow-run], [data-workflow-stop]");
    if (!btn) return;
    event.preventDefault();
    btn.disabled = true;
    var run = btn.hasAttribute("data-workflow-run");
    var path = run
      ? "/api/workflows/" + encodeURIComponent(btn.dataset.workflowRun) + "/runs"
      : "/api/workflow-runs/" + encodeURIComponent(btn.dataset.workflowStop) + "/stop";
    api("POST", path).then(function (res) {
      // A workflow already running answers 409; its page shows that run.
      if (res.status >= 300 && res.status !== 409) {
        btn.disabled = false;
        window.alert(res.data.error || ("HTTP " + res.status));
        return;
      }
      if (run) window.location.href = "/workflows/" + encodeURIComponent(btn.dataset.workflowRun);
      else window.location.reload();
    });
  });

  // A section with data-refresh reloads the page while something in it
  // runs, so a workflow's steps show their progress.
  var refresh = document.querySelector("[data-refresh]");
  if (refresh) {
    setTimeout(function () { window.location.reload(); }, parseInt(refresh.dataset.refresh, 10) || 3000);
  }

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
    var btn = event.target.closest("[data-run], [data-attach], [data-stop], [data-kill], [data-delete], [data-copy]");
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

  Array.prototype.forEach.call(document.querySelectorAll("[data-output]"), function (pre) {
    if (!window.Terminal) return;
    var id = pre.dataset.output;
    fetch("/api/runs/" + encodeURIComponent(id) + "/cast", { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) throw new Error(res.status === 404 ? "no recording" : "HTTP " + res.status);
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
  });

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

  // The sidebar is a drawer under the breakpoint. #menu opens it on load.
  var menuToggle = document.getElementById("menu-toggle");
  if (menuToggle) {
    var setMenu = function (open) {
      document.body.classList.toggle("menu-open", open);
      menuToggle.setAttribute("aria-expanded", open ? "true" : "false");
    };
    menuToggle.addEventListener("click", function () { setMenu(!document.body.classList.contains("menu-open")); });
    document.getElementById("menu-backdrop").addEventListener("click", function () { setMenu(false); });
    if (location.hash === "#menu") setMenu(true);
  }

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
      var running = r.state === "running";
      var verb = !running ? "View" : r.can_run ? "Attach to" : "Watch";
      return '<a href="/runs/' + encodeURIComponent(r.id) + '" data-attach="' + escapeHTML(r.id) + '" class="recent" title="' +
        escapeHTML(verb + " " + r.folder + " / " + r.job + ", " + r.user_name + ", " + r.state) + '">' +
        '<i class="dot ' + escapeHTML(r.state) + '"></i>' +
        '<span class="what"><b>' + escapeHTML(r.job) + '</b><small>' + escapeHTML(r.folder) + '</small></span>' +
        (running ? '<svg class="ic act"><use href="#i-' + (r.can_run ? "play" : "eye") + '"/></svg>' : "") +
        '<span class="avatar sm" title="' + escapeHTML(r.user_name) + '">' + escapeHTML(initials(r.user_name)) + '</span></a>';
    }).join("");
  }

  function poll() {
    if (document.hidden) return;
    api("GET", "/api/runs?limit=10").then(function (res) {
      if (res.status !== 200 || !res.data.runs) return;
      var runs = res.data.runs;
      renderRecents(runs);
      var running = {}, runningJobs = {};
      runs.forEach(function (r) {
        if (r.state !== "running") return;
        running[r.folder] = true;
        runningJobs[r.folder + "/" + r.job] = true;
      });
      document.querySelectorAll(".side [data-folder] > summary .dot").forEach(function (dot) {
        dot.classList.toggle("running", !!running[dot.closest("[data-folder]").dataset.folder]);
      });
      document.querySelectorAll(".side [data-job-key] .dot").forEach(function (dot) {
        dot.classList.toggle("running", !!runningJobs[dot.parentElement.dataset.jobKey]);
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

  // The frames are inline SVG, so they use the page's own faces; each
  // one is fetched again and swapped in place.
  function refreshFrames() {
    if (document.hidden) return;
    document.querySelectorAll(".frame[data-frame]").forEach(function (frame) {
      fetch("/api/runs/" + encodeURIComponent(frame.dataset.frame) + "/frame.svg", { credentials: "same-origin" })
        .then(function (res) { return res.ok ? res.text() : null; })
        .then(function (svg) { if (svg) frame.innerHTML = svg; })
        .catch(function () {});
    });
  }

  if (document.querySelector(".side")) {
    poll();
    setInterval(poll, 4000);
    if (document.querySelector(".frame[data-frame]")) setInterval(refreshFrames, 2000);
  }
})();
