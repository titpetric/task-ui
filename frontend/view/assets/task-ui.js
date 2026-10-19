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
    // A page that follows its topic drew the change already.
    if (changed && !follows) window.location.reload();
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
    // A run that ended well is run again; one that did not is forced,
    // which also gets past what refused it.
    var can = s.run && s.run.can_run;
    buttons({ again: can && msg.state === "ok", force: can && msg.state !== "ok" });
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
      // A set is the checked values of the boxes that share a name: a
      // list, empty when none is checked.
      if (el.dataset.type === "set") {
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

  // The workflow page follows its topic; live says so, below. Any page
  // that follows a topic draws its changes and is never reloaded.
  var live = document.querySelector('[data-sse^="workflow:"]');
  var follows = document.querySelector("[data-sse]");

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
      // The workflow's own page hears of the run from its events and
      // draws it; elsewhere, go there.
      if (live) return;
      if (run) window.location.href = "/workflows/" + encodeURIComponent(btn.dataset.workflowRun);
      else window.location.reload();
    });
  });

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
      if (runs && !row) {
        var empty = runs.querySelector(".empty");
        if (empty) empty.remove();
        row = el("tr");
        row.dataset.runId = data.run_id;
        var started = el("td", "nowrap", "just now");
        started.dataset.label = "Started";
        var by = el("td", "", data.user_name || "");
        by.dataset.label = "By";
        var result = el("td");
        result.dataset.label = "Result";
        result.dataset.result = "";
        result.appendChild(stateSpan("running"));
        var steps = el("td");
        steps.dataset.label = "Steps";
        var chain = el("span", "chain");
        nodes.forEach(function (node, i) {
          if (i) chain.appendChild(el("span", "faint", " → "));
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
        row.appendChild(started);
        row.appendChild(by);
        row.appendChild(result);
        row.appendChild(steps);
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
          a.href = "/f/" + data.folder + "/" + data.job;
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
  // the page hold for the rows it adds.
  var runsTable = document.querySelector('[data-sse="runs"] [data-runs-table]');
  if (runsTable && window.EventSource) {
    var params = new URLSearchParams(location.search);
    var matches = function (run) {
      return (!params.get("folder") || params.get("folder") === run.folder) &&
        (!params.get("job") || run.job.indexOf(params.get("job")) >= 0) &&
        (!params.get("state") || params.get("state") === run.state);
    };
    var took = function (run) {
      if (!run.started_at) return "";
      var end = run.finished_at ? Date.parse(run.finished_at) : Date.now();
      var s = Math.max(0, (end - Date.parse(run.started_at)) / 1000);
      return s < 60 ? s.toFixed(1) + "s" : Math.floor(s / 60) + "m" + ("0" + Math.floor(s % 60)).slice(-2) + "s";
    };
    var stateHTML = function (run) {
      return '<span class="state"><i class="dot ' + escapeHTML(run.state) + '"></i>' + escapeHTML(run.state) +
        (run.state !== "ok" && run.state !== "running" ? ' <span class="faint">(' + (run.exit_code || 0) + ")</span>" : "") + "</span>";
    };
    var actionsHTML = function (run, canRun) {
      var id = escapeHTML(run.id);
      var html = '<a class="btn btn-sm" href="/runs/' + id + '" data-attach="' + id + '">' + (run.state === "running" ? "Attach" : "View") + "</a>";
      if (run.state === "running" && canRun) {
        html += '<button type="button" class="btn-icon" data-stop="' + id + '" title="Stop the run" aria-label="Stop the run"><svg class="ic"><use href="#i-stop"/></svg></button>' +
          '<button type="button" class="btn-icon danger" data-kill="' + id + '" title="Kill the run" aria-label="Kill the run"><svg class="ic"><use href="#i-trash"/></svg></button>';
      }
      return html;
    };
    var whoHTML = function (run) {
      return '<span class="who-ran"><span class="avatar sm" title="' + escapeHTML(run.user_name) + '">' + escapeHTML(initials(run.user_name)) + "</span> " +
        escapeHTML(run.user_name) + (run.scheduled ? ' <span class="badge">scheduled</span>' : "") +
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
      if (!matches(run)) return;
      var empty = runsTable.querySelector(".empty");
      if (empty) empty.remove();
      row = document.createElement("tr");
      row.dataset.runId = run.id;
      var folder = escapeHTML(run.folder), job = escapeHTML(run.job);
      row.innerHTML =
        '<td class="nowrap" data-label="Started">just now</td>' +
        '<td data-label="Folder"><a href="/f/' + folder + '">' + folder + "</a></td>" +
        '<td class="key"><a href="/f/' + folder + "/" + job + '">' + job + "</a></td>" +
        '<td data-label="By">' + whoHTML(run) + "</td>" +
        '<td class="nowrap" data-label="Took" data-took>' + took(run) + "</td>" +
        '<td data-label="Result" data-result>' + stateHTML(run) + "</td>" +
        '<td class="actions" data-actions>' + actionsHTML(run, false) + "</td>";
      runsTable.insertBefore(row, runsTable.firstChild);
      // Whether this user may end it comes with the run's own view.
      api("GET", "/api/runs/" + encodeURIComponent(run.id)).then(function (res) {
        if (res.status === 200 && res.data.state === "running") row.querySelector("[data-actions]").innerHTML = actionsHTML(res.data, res.data.can_run);
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

  // The dashboard follows the runs topic too: a run that starts adds its
  // tile, polled for its frame; one that ends marks its tile and the last
  // frame is fetched. The running count and the control centre's buttons
  // follow: an eye on the job running, Run again when it ended.
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
    var slotHTML = function (slot, run) {
      var label = slot.dataset.label ? "<span>" + escapeHTML(slot.dataset.label) + "</span>" : "";
      var named = slot.dataset.label ? " named" : "";
      var what = escapeHTML(slot.dataset.folder + " / " + slot.dataset.name);
      if (run && run.state === "running") {
        return '<a class="cc' + named + ' running" href="/runs/' + escapeHTML(run.id) + '" data-attach="' + escapeHTML(run.id) +
          '" title="Attach to ' + what + ', running"><svg class="ic"><use href="#i-eye"/></svg>' + label + "</a>";
      }
      if (slot.hasAttribute("data-can")) {
        return '<button type="button" class="cc' + named + '" data-run data-folder="' + escapeHTML(slot.dataset.folder) + '" data-job="' +
          escapeHTML(slot.dataset.job) + '" data-title="' + what + '" title="Run ' + what + '"><svg class="ic"><use href="#i-play"/></svg>' + label + "</button>";
      }
      return '<a class="cc' + named + ' idle" href="/f/' + escapeHTML(slot.dataset.folder) + "/" + escapeHTML(slot.dataset.job) + '" title="' + what +
        ': your role may not run it"><i class="shebang" aria-hidden="true">#!</i>' + label + "</a>";
    };
    onRuns(function (run) {
      var tile = tiles.querySelector('[data-tile="' + run.id + '"]');
      if (!tile && run.state === "running") {
        tile = document.createElement("a");
        tile.className = "tile";
        tile.href = "/runs/" + run.id;
        tile.dataset.attach = run.id;
        tile.dataset.tile = run.id;
        tile.innerHTML = '<div class="tile-head"><i class="dot running"></i><b>' + escapeHTML(run.folder + " / " + run.job) + "</b>" +
          '<span class="spacer"></span><span class="faint"></span>' +
          '<span class="avatar sm" title="' + escapeHTML(run.user_name) + '">' + escapeHTML(initials(run.user_name)) + "</span></div>" +
          '<div class="frame" data-frame="' + escapeHTML(run.id) + '" role="img" aria-label="Terminal of ' + escapeHTML(run.folder + " / " + run.job) + '"></div>';
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
      dashboard.querySelectorAll('.cc-slot[data-cc="' + run.folder + "/" + run.job + '"]').forEach(function (slot) {
        slot.innerHTML = slotHTML(slot, run);
      });
      countRunning();
    });
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
    // A link inside the element matched is a link: it navigates, it does
    // not run or attach for the element around it.
    var link = event.target.closest("a[href]");
    if (link && link !== btn && btn.contains(link)) return;

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
    // Opened by the link it is simply open, without the slide: the page
    // is new, there is nothing to slide in from.
    if (location.hash === "#menu") {
      var side = document.getElementById("side");
      side.style.transition = "none";
      setMenu(true);
      requestAnimationFrame(function () { requestAnimationFrame(function () { side.style.transition = ""; }); });
    }
  }

  // The projects in the sidebar fold on their chevron, and the folded
  // ones are remembered in this browser. The open page's project stays
  // open.
  var FOLDED = "task-ui.folded";
  var folded = [];
  try { folded = JSON.parse(localStorage.getItem(FOLDED) || "[]"); } catch (e) { folded = []; }
  document.querySelectorAll(".folders details.project").forEach(function (project) {
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

  // renderRecents draws the runs of jobs and of workflows, newest first.
  function renderRecents(runs, flows) {
    if (!recents) return;
    var entries = runs.map(function (r) { return { run: r, at: r.started_at || "" }; })
      .concat((flows || []).map(function (f) { return { flow: f, at: f.started_at || "" }; }));
    entries.sort(function (a, b) { return a.at < b.at ? 1 : a.at > b.at ? -1 : 0; });
    entries = entries.slice(0, 10);
    if (!entries.length) {
      recents.innerHTML = '<div class="empty">Runs started by anyone show up here.</div>';
      return;
    }
    recents.innerHTML = entries.map(function (e) {
      if (e.flow) {
        var f = e.flow;
        return '<a href="/workflows/' + encodeURIComponent(f.workflow_id) + '" class="recent flow" title="' +
          escapeHTML("Workflow " + f.name + ", " + f.user_name + ", " + f.state) + '">' +
          '<i class="dot ' + escapeHTML(f.state) + '"></i>' +
          '<span class="what"><b>' + escapeHTML(f.name) + '</b><small><svg class="ic"><use href="#i-flow"/></svg>workflow</small></span>' +
          (f.state === "running" ? '<svg class="ic act"><use href="#i-eye"/></svg>' : "") +
          '<span class="avatar sm" title="' + escapeHTML(f.user_name) + '">' + escapeHTML(initials(f.user_name)) + '</span></a>';
      }
      var r = e.run;
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
    Promise.all([api("GET", "/api/runs?limit=10"), api("GET", "/api/workflow-runs?limit=10")]).then(function (both) {
      var res = both[0], flows = both[1];
      if (res.status !== 200 || !res.data.runs) return;
      var runs = res.data.runs;
      renderRecents(runs, flows.status === 200 ? flows.data.runs : []);
      var running = {}, runningJobs = {};
      runs.forEach(function (r) {
        if (r.state !== "running") return;
        running[r.folder] = true;
        runningJobs[r.folder + "/" + r.job] = true;
      });
      document.querySelectorAll(".side [data-folder] > summary .dot").forEach(function (dot) {
        dot.classList.toggle("running", !!running[dot.closest("[data-folder]").dataset.folder]);
      });
      document.querySelectorAll(".side [data-job-key] .shebang").forEach(function (dot) {
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

  // frameURL is where a run's frame is drawn: its last screen, or with
  // all its whole output so far.
  function frameURL(runId, all) {
    return "/api/runs/" + encodeURIComponent(runId) + "/frame.svg" + (all ? "?all=1" : "");
  }

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
