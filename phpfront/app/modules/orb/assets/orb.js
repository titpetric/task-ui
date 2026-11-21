// orb.js drives the meetings page: create a room, join one, send the
// microphone over the pipe, play the other speakers, and write the live
// transcript under the name of whoever said it.
(function () {
  "use strict";

  // RATE is the wire and channel rate, mono 16-bit PCM, and CHUNK is how much
  // of it one binary frame holds.
  var RATE = 16000;
  var CHUNK = 250;
  // VOICE is the level over which a chunk is speech: quiet chunks are not
  // sent, so a channel holds only the speech recorded for it.
  var VOICE = 0.02;

  var own = document.currentScript ? document.currentScript.src : "";

  if (!own) {
    var tags = document.querySelectorAll('script[src*="/modules/orb/orb.js"]');

    if (tags.length) {
      own = tags[tags.length - 1].src;
    }
  }

  var WORKLET = own ? new URL("../../theme/voice/tap.js" + (own.indexOf("?") >= 0 ? own.slice(own.indexOf("?")) : ""), own).href : "";

  if (!WORKLET) {
    console.warn("orb: no worklet URL, the microphone is not recorded");
  }

  var room = document.querySelector("[data-orb-room]");

  if (!room) {
    initNew();
    return;
  }

  var cfg = {
    id: room.getAttribute("data-meeting"),
    state: room.getAttribute("data-state"),
    self: room.getAttribute("data-self"),
    selfName: room.getAttribute("data-self-name"),
    host: room.getAttribute("data-host") === "1",
    canEdit: room.getAttribute("data-can-edit") === "1",
    api: room.getAttribute("data-api"),
    base: room.getAttribute("data-base"),
    started: parseInt(room.getAttribute("data-started"), 10) || 0,
    limit: parseInt(room.getAttribute("data-limit"), 10) || 3600,
  };

  var el = {
    people: room.querySelector("[data-people]"),
    join: room.querySelector("[data-orb-join]"),
    leave: document.querySelector("[data-orb-leave]"),
    speak: room.querySelector("[data-orb-speak]"),
    mic: room.querySelector("[data-orb-mic]"),
    micLabel: room.querySelector("[data-orb-miclabel]"),
    status: room.querySelector("[data-orb-status]"),
    log: room.querySelector("[data-orb-log]"),
    empty: room.querySelector("[data-orb-empty]"),
    timer: room.querySelector("[data-orb-timer]"),
    summary: room.querySelector("[data-orb-summary]"),
    summaryBody: room.querySelector("[data-summary-body]"),
    summaryMessage: room.querySelector("[data-summary-message]"),
    summaryRetry: room.querySelector("[data-summary-retry]"),
  };

  var ws = null;
  var joined = false;
  var joining = false;
  var leaving = false;
  var rejoining = 0;
  var ping = 0;
  var micOn = false;
  var capture = null;
  var sound = null;
  var pending = [];
  var pendingLen = 0;
  var spoken = 0;
  var recognizer = null;
  var audioMeta = null;
  var nextAt = {};
  var speaking = {};
  var ownSpeaking = 0;
  var polling = 0;

  function api(method, path, body) {
    var options = { method: method, credentials: "same-origin", headers: {} };

    if (body !== undefined) {
      options.headers["Content-Type"] = "application/json";
      options.body = JSON.stringify(body);
    }

    return fetch(cfg.api + path, options).then(function (res) {
      return res.text().then(function (text) {
        var data = {};

        if (text) {
          try {
            data = JSON.parse(text);
          } catch (err) {
            data = {};
          }
        }

        if (!res.ok) {
          throw new Error(data.error || "the request failed (" + res.status + ")");
        }

        return data;
      });
    });
  }

  function say(text, bad) {
    el.status.textContent = text;
    el.status.classList.toggle("bad", !!bad);
  }

  function socketURL(path) {
    var scheme = location.protocol === "https:" ? "wss:" : "ws:";

    return scheme + "//" + location.host + cfg.api + path;
  }

  function send(message) {
    if (ws && ws.readyState === 1) {
      ws.send(JSON.stringify(message));
    }
  }

  // ------------------------------------------------------------- transcript

  function lineID(line) {
    return line.id || (line.kind + "-" + line.user_id + "-" + line.body);
  }

  function clock(iso) {
    // The column holds Go's time shape, which Date reads only as RFC 3339.
    var go = /^(\d{4}-\d{2}-\d{2}) (\d{2}:\d{2}:\d{2}(?:\.\d+)?) ([+-]\d{2})(\d{2}) \S+$/.exec(iso || "");

    if (go) {
      iso = go[1] + "T" + go[2] + go[3] + ":" + go[4];
    } else if (iso && !/(?:Z|[+-]\d{2}:\d{2})$/.test(iso)) {
      iso += "Z";
    }

    var at = new Date(iso || "");

    if (isNaN(at.getTime())) {
      return "";
    }

    return at.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: false });
  }

  // renderTimes redraws the server's transcript times in the reader's zone.
  function renderTimes() {
    var cells = room.querySelectorAll("[data-at]");

    for (var i = 0; i < cells.length; i++) {
      var at = clock(cells[i].getAttribute("data-at"));

      if (at !== "") {
        cells[i].textContent = at;
      }
    }
  }

  function saidLine(line) {
    var row = document.createElement("div");
    row.className = "orb-line";
    row.setAttribute("data-line-id", lineID(line));

    var who = document.createElement("span");
    who.className = "orb-who";
    who.textContent = line.user_name;
    row.appendChild(who);

    var body = document.createElement("span");
    body.className = "orb-said";
    body.textContent = line.body;
    row.appendChild(body);

    var at = document.createElement("span");
    at.className = "orb-at";
    at.textContent = clock(line.created_at);
    row.appendChild(at);

    var play = document.createElement("button");
    play.type = "button";
    play.className = "btn-icon";
    play.title = "Play this line";
    play.setAttribute("data-play", "");
    play.setAttribute("data-user", line.user_id);
    play.setAttribute("data-start", line.start_ms);
    play.setAttribute("data-end", line.end_ms);
    play.innerHTML = '<i class="ic ic-play" aria-hidden="true"></i>';
    row.appendChild(play);

    return row;
  }

  function systemLine(line) {
    var row = document.createElement("div");
    row.className = "orb-line system";
    row.setAttribute("data-line-id", lineID(line));

    var body = document.createElement("span");
    body.textContent = line.body;
    row.appendChild(body);

    var at = document.createElement("span");
    at.className = "orb-at";
    at.textContent = clock(line.created_at);
    row.appendChild(at);

    return row;
  }

  function addLine(line) {
    if (!line || !el.log) {
      return;
    }

    if (line.kind !== "said") {
      line.id = line.id || ("system-" + line.kind + "-" + line.user_id + "-" + line.body);
    }

    if (el.log.querySelector('[data-line-id="' + cssEscape(lineID(line)) + '"]')) {
      return;
    }

    var row = line.kind === "said" ? saidLine(line) : systemLine(line);
    el.log.appendChild(row);
    hideEmpty();
    el.log.scrollTop = el.log.scrollHeight;
  }

  function partialID(user) {
    return "partial-" + user;
  }

  function showPartial(user, name, text) {
    if (!el.log) {
      return;
    }

    var row = el.log.querySelector('[data-line-id="' + cssEscape(partialID(user)) + '"]');

    if (!row) {
      row = document.createElement("div");
      row.className = "orb-line partial";
      row.setAttribute("data-line-id", partialID(user));

      var who = document.createElement("span");
      who.className = "orb-who";
      who.textContent = name;
      row.appendChild(who);

      var body = document.createElement("span");
      body.className = "orb-said";
      row.appendChild(body);

      var play = document.createElement("button");
      play.type = "button";
      play.className = "btn-icon";
      play.innerHTML = '<i class="ic ic-play" aria-hidden="true"></i>';
      row.appendChild(play);

      el.log.appendChild(row);
    }

    row.querySelector(".orb-said").textContent = text;
    hideEmpty();
    el.log.scrollTop = el.log.scrollHeight;
  }

  function clearPartial(user) {
    var row = el.log && el.log.querySelector('[data-line-id="' + cssEscape(partialID(user)) + '"]');

    if (row) {
      row.parentNode.removeChild(row);
    }
  }

  function hideEmpty() {
    if (el.empty) {
      el.empty.hidden = true;
    }
  }

  function cssEscape(value) {
    return String(value).replace(/["\\]/g, "\\$&");
  }

  // ------------------------------------------------------------ participants

  function person(id) {
    var found = null;

    if (!el.people) {
      return found;
    }

    found = el.people.querySelector('[data-person="' + cssEscape(id) + '"]');

    if (found || id === "") {
      return found;
    }

    found = document.createElement("span");
    found.className = "orb-person";
    found.setAttribute("data-person", id);
    found.setAttribute("data-here", "0");

    var dot = document.createElement("span");
    dot.className = "orb-dot";
    found.appendChild(dot);

    var name = document.createElement("span");
    name.className = "orb-name";
    found.appendChild(name);

    el.people.appendChild(found);

    return found;
  }

  function markPerson(p) {
    var row = person(p.user_id);

    if (!row) {
      return;
    }

    row.querySelector(".orb-name").textContent = p.user_name;
    row.setAttribute("data-here", p.left_at ? "0" : "1");
    row.classList.toggle("here", !p.left_at);
    row.classList.toggle("gone", !!p.left_at);
    row.classList.toggle("self", p.user_id === cfg.self);
  }

  function markHere(id, name, here) {
    var row = person(id);

    if (!row) {
      return;
    }

    if (name) {
      row.querySelector(".orb-name").textContent = name;
    }

    row.setAttribute("data-here", here ? "1" : "0");
    row.classList.toggle("here", here);
    row.classList.toggle("gone", !here);
  }

  function markMuted(id, muted) {
    var row = person(id);

    if (row) {
      row.classList.toggle("muted", muted);
    }
  }

  function markSpeaking(id) {
    var row = person(id);

    if (!row) {
      return;
    }

    row.classList.add("speaking");
    clearTimeout(speaking[id]);
    speaking[id] = setTimeout(function () {
      row.classList.remove("speaking");
    }, 400);
  }

  // ------------------------------------------------------------------- pipe

  function join() {
    if (joined || joining || !cfg.canEdit) {
      return;
    }

    joining = true;
    ensureSound();
    say("Joining...");
    api("POST", "/meetings/" + cfg.id + "/join").then(function () {
      joining = false;
      joined = true;

      if (el.join) {
        el.join.hidden = true;
      }

      if (el.leave) {
        el.leave.setAttribute("data-joined", "1");
      }

      if (el.speak) {
        el.speak.hidden = false;
      }

      say("Connecting...");
      connect();
    }, function (err) {
      joining = false;
      say(err.message, true);
    });
  }

  // stopLeaving ends the socket and the microphone without logging a part:
  // the leave button does that, and a closed socket waits for the sweeper.
  function stopLeaving() {
    leaving = false;
    joined = false;
    clearInterval(ping);
    stopMic();

    if (ws) {
      var dead = ws;
      ws = null;
      dead.onclose = null;
      dead.close();
    }

    if (el.speak) {
      el.speak.hidden = true;
    }

    if (el.join) {
      el.join.hidden = false;
    }

    if (el.leave) {
      el.leave.setAttribute("data-joined", "0");
    }

    markHere(cfg.self, cfg.selfName, false);
    say("You left the room. Join to listen again.");
  }

  // checkState reloads a room that ended while this page could not hear it.
  function checkState() {
    api("GET", "/meetings/" + cfg.id).then(function (data) {
      if (data.meeting && data.meeting.state !== "open") {
        location.reload();
      }
    }, function () {
      // A failed read leaves the status unchanged.
    });
  }

  function connect() {
    ws = new WebSocket(socketURL("/meetings/" + cfg.id + "/pipe"));
    ws.binaryType = "arraybuffer";

    ws.onopen = function () {
      say("Connected. Press the microphone to speak.");
      clearInterval(ping);
      ping = setInterval(function () {
        send({ type: "ping" });
      }, 10000);
    };

    ws.onmessage = function (event) {
      if (typeof event.data === "string") {
        handle(JSON.parse(event.data));
        return;
      }

      if (audioMeta) {
        play(event.data, audioMeta);
        audioMeta = null;
      }
    };

    ws.onclose = function () {
      clearInterval(ping);

      if (joined && !leaving) {
        var wait = Math.min(10000, 1000 * Math.pow(2, rejoining++));
        say("Reconnecting...", true);
        setTimeout(function () {
          if (joined && !leaving) {
            connect();
          }
        }, wait);
      }
    };
  }

  function handle(message) {
    rejoining = 0;

    switch (message.type) {
      case "hello":
        (message.participants || []).forEach(markPerson);
        (message.lines || []).forEach(addLine);

        if (message.meeting && message.meeting.state !== "open") {
          location.reload();
        }

        break;
      case "audio":
        audioMeta = message;
        break;
      case "line":
        clearPartial(message.line.user_id);
        addLine(message.line);
        markPerson({ user_id: message.line.user_id, user_name: message.line.user_name, left_at: "" });
        break;
      case "partial":
        showPartial(message.user_id, message.user_name, message.text);
        break;
      case "join":
      case "rejoin":
        markHere(message.line.user_id, message.line.user_name, true);
        markMuted(message.line.user_id, false);
        addLine(message.line);
        break;
      case "leave":
        markHere(message.line.user_id, message.line.user_name, false);
        addLine(message.line);
        break;
      case "mute":
        markMuted(message.user_id, !!message.muted);
        break;
      case "end":
        joined = false;
        leaving = true;
        clearInterval(ping);
        location.reload();
        break;
      case "summary":
        if (message.state) {
          finishSummary({ state: message.state, message: message.message, html: message.html || "" });
        }
        break;
      case "error":
        say(message.message, true);
        checkState();
        break;
    }
  }

  // ------------------------------------------------------------------ audio

  function ensureSound() {
    if (!sound) {
      sound = new (window.AudioContext || window.webkitAudioContext)();
    }

    if (sound.state === "suspended") {
      sound.resume();
    }

    return sound;
  }

  function play(bytes, meta) {
    var ctx = ensureSound();
    var samples = new Int16Array(bytes, 0, Math.floor(bytes.byteLength / 2));

    if (!samples.length) {
      return;
    }

    var buffer = ctx.createBuffer(1, samples.length, RATE);
    var data = buffer.getChannelData(0);

    for (var i = 0; i < samples.length; i++) {
      data[i] = samples[i] / 32768;
    }

    var source = ctx.createBufferSource();
    source.buffer = buffer;
    source.connect(ctx.destination);

    var at = Math.max(ctx.currentTime + 0.08, nextAt[meta.user_id] || 0);
    source.start(at);
    nextAt[meta.user_id] = at + buffer.duration;
    markSpeaking(meta.user_id);
  }

  function captureStart() {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      say("This page needs a secure origin and a microphone.", true);
      return Promise.reject(new Error("no microphone"));
    }

    return navigator.mediaDevices.getUserMedia({
      audio: { channelCount: 1, echoCancellation: true, noiseSuppression: true },
    }).then(function (stream) {
      var ctx = ensureSound();
      var source = ctx.createMediaStreamSource(stream);
      capture = { stream: stream, ctx: ctx, node: null };

      if (!WORKLET) {
        return;
      }

      return ctx.audioWorklet.addModule(WORKLET).then(function () {
        var node = new AudioWorkletNode(ctx, "tap");
        node.port.onmessage = function (event) {
          if (micOn) {
            onSamples(event.data);
          }
        };
        source.connect(node);
        // A silent gain keeps the graph pulled without playing the microphone
        // back into the room.
        var out = ctx.createGain();
        out.gain.value = 0;
        node.connect(out);
        out.connect(ctx.destination);
        capture.node = node;
      });
    });
  }

  function captureStop() {
    if (!capture) {
      return;
    }

    if (capture.node) {
      capture.node.port.onmessage = null;
    }

    capture.stream.getTracks().forEach(function (track) {
      track.stop();
    });
    capture = null;
    pending = [];
    pendingLen = 0;
  }

  function onSamples(frame) {
    pending.push(frame);
    pendingLen += frame.length;

    if (pendingLen < Math.floor(capture.ctx.sampleRate * CHUNK / 1000)) {
      return;
    }

    var all = new Float32Array(pendingLen);
    var at = 0;

    pending.forEach(function (part) {
      all.set(part, at);
      at += part.length;
    });

    pending = [];
    pendingLen = 0;

    var ratio = capture.ctx.sampleRate / RATE;
    var count = Math.floor(all.length / ratio);

    if (count < 1) {
      return;
    }

    var pcm = new Int16Array(count);
    var energy = 0;

    for (var i = 0; i < count; i++) {
      var from = Math.floor(i * ratio);
      var to = Math.min(all.length, Math.floor((i + 1) * ratio));
      var sum = 0;

      for (var j = from; j < to; j++) {
        sum += all[j];
      }

      var value = Math.max(-1, Math.min(1, sum / Math.max(1, to - from)));
      pcm[i] = value * 0x7fff;
      energy += value * value;
    }

    if (Math.sqrt(energy / count) < VOICE) {
      return;
    }

    if (ws && ws.readyState === 1) {
      ws.send(pcm.buffer);
      ownVoice();
    }

    spoken += Math.round(1000 * count / RATE);
  }

  function listen() {
    var SR = window.SpeechRecognition || window.webkitSpeechRecognition;

    if (!SR) {
      say("Live. This browser writes no live transcript; the room still hears you.");
      return;
    }

    recognizer = new SR();
    recognizer.continuous = true;
    recognizer.interimResults = true;
    recognizer.lang = document.documentElement.lang || navigator.language || "en-US";

    recognizer.onresult = function (event) {
      if (!micOn) {
        return;
      }

      var interim = "";
      var final = "";

      for (var i = event.resultIndex; i < event.results.length; i++) {
        var result = event.results[i];

        if (result.isFinal) {
          final += result[0].transcript;
        } else {
          interim += result[0].transcript;
        }
      }

      if (interim.trim() !== "") {
        showPartial(cfg.self, cfg.selfName, interim.trim());
        send({ type: "partial", text: interim.trim() });
      }

      if (final.trim() !== "") {
        var duration = spoken;

        if (duration < 500) {
          duration = Math.min(60000, Math.max(1000, final.split(/\s+/).length * 350));
        }

        clearPartial(cfg.self);
        send({ type: "line", text: final.trim(), duration_ms: spoken || duration });
        spoken = 0;
      }
    };

    recognizer.onend = function () {
      if (micOn && joined && !leaving) {
        try {
          recognizer.start();
        } catch (err) {
          // A restart while the engine is still closing is the engine's to
          // refuse; the next press opens it again.
        }
      }
    };

    recognizer.onerror = function (event) {
      if (event.error === "not-allowed" || event.error === "service-not-allowed") {
        say("The browser refused speech recognition; the room still hears you.", true);
      }
    };

    try {
      recognizer.start();
    } catch (err) {
      recognizer = null;
    }
  }

  function stopListening() {
    if (!recognizer) {
      return;
    }

    var dead = recognizer;
    recognizer = null;

    try {
      dead.stop();
    } catch (err) {
      // Stopping an engine that already ended changes nothing.
    }
  }

  // ownVoice lights the microphone while a chunk crosses the pipe.
  function ownVoice() {
    if (!el.mic) {
      return;
    }

    el.mic.classList.add("speaking");

    if (el.micLabel) {
      el.micLabel.textContent = "Speaking";
    }

    clearTimeout(ownSpeaking);
    ownSpeaking = setTimeout(function () {
      el.mic.classList.remove("speaking");

      if (micOn && el.micLabel) {
        el.micLabel.textContent = "Microphone on";
      }
    }, 500);
  }

  function unmute() {
    ensureSound();
    micOn = true;
    el.mic.setAttribute("aria-pressed", "true");
    el.mic.setAttribute("title", "Mute");
    say("Live. Listening and speaking.");
    send({ type: "mute", on: false });
    markMuted(cfg.self, false);

    if (el.micLabel) {
      el.micLabel.textContent = "Microphone on";
    }

    captureStart().then(function () {
      if (micOn) {
        listen();
      }
    }, function (err) {
      micOn = false;
      el.mic.setAttribute("aria-pressed", "false");
      el.mic.setAttribute("title", "Start speaking");
      say("No microphone: " + err.message, true);
    });
  }

  // stopMic closes the microphone and the recogniser without telling the room.
  function stopMic() {
    micOn = false;
    clearTimeout(ownSpeaking);

    if (el.mic) {
      el.mic.setAttribute("aria-pressed", "false");
      el.mic.setAttribute("title", "Start speaking");
      el.mic.classList.remove("speaking");
    }

    if (el.micLabel) {
      el.micLabel.textContent = "Microphone off";
    }

    captureStop();
    stopListening();
    clearPartial(cfg.self);
    spoken = 0;
  }

  function mute() {
    send({ type: "mute", on: true });
    markMuted(cfg.self, true);
    stopMic();
    say("Muted. Press the microphone to speak again.");
  }

  // ------------------------------------------------------------------ timer

  function tick() {
    if (!el.timer || !cfg.started) {
      return;
    }

    var left = cfg.limit - (Math.floor(Date.now() / 1000) - cfg.started);

    if (left <= 0) {
      el.timer.textContent = "ending...";
      return;
    }

    var minutes = Math.floor(left / 60);
    var seconds = left % 60;
    el.timer.textContent = minutes + ":" + (seconds < 10 ? "0" : "") + seconds + " left";
  }

  // ----------------------------------------------------------------- summary

  function finishSummary(data) {
    if (!el.summary) {
      return;
    }

    el.summary.setAttribute("data-state", data.state);

    if (data.html) {
      el.summaryBody.innerHTML = data.html;
    }

    if (el.summaryMessage) {
      el.summaryMessage.textContent = data.message || "";
    }

    if (el.summaryRetry) {
      el.summaryRetry.hidden = data.state !== "failed";
    }

    if (data.state === "working") {
      pollSummary();
    }
  }

  function pollSummary() {
    if (!el.summary || polling) {
      return;
    }

    polling = setInterval(function () {
      api("GET", "/meetings/" + cfg.id + "/summary").then(function (data) {
        if (data.state !== "working") {
          clearInterval(polling);
          polling = 0;
          finishSummary(data);
        }
      }, function () {
        // A failed read is tried again on the next tick.
      });
    }, 3000);
  }

  function loadSummary() {
    if (!el.summary) {
      return;
    }

    api("GET", "/meetings/" + cfg.id + "/summary").then(function (data) {
      finishSummary(data);

      if (data.state === "working" && !polling) {
        pollSummary();
      }
    }, function (err) {
      say(err.message, true);
    });
  }

  // ------------------------------------------------------------- page events

  document.addEventListener("click", function (event) {
    ensureSound();

    var play = event.target.closest("[data-play]");

    if (play) {
      var url = cfg.api + "/meetings/" + cfg.id + "/channels/" + play.getAttribute("data-user");
      var start = parseInt(play.getAttribute("data-start"), 10) / 1000;
      var end = parseInt(play.getAttribute("data-end"), 10) / 1000;
      var audio = new Audio(url + "#t=" + start.toFixed(3) + "," + end.toFixed(3));

      audio.play();
      audio.addEventListener("timeupdate", function () {
        if (audio.currentTime >= end) {
          audio.pause();
        }
      });
      return;
    }

    var copy = event.target.closest("[data-orb-copy]");

    if (copy) {
      var link = location.origin + cfg.base + "/orb/" + cfg.id;

      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(link).then(function () {
          say("Link copied.");
        }, function () {
          window.prompt("Copy this link", link);
        });
      } else {
        window.prompt("Copy this link", link);
      }

      return;
    }

    var leave = event.target.closest("[data-orb-leave]");

    if (leave) {
      if (leaving) {
        return;
      }

      leaving = true;
      api("POST", "/meetings/" + cfg.id + "/leave").then(stopLeaving, function (err) {
        leaving = false;
        say(err.message, true);
      });
      return;
    }

    var finish = event.target.closest("[data-orb-finish]");

    if (finish) {
      window.taskUI.confirm(finish.getAttribute("data-confirm") || "End the meeting?", "End").then(function (ok) {
        if (!ok) return;

        api("POST", "/meetings/" + cfg.id + "/finish").then(function () {
          location.reload();
        }, function (err) {
          say(err.message, true);
        });
      });
      return;
    }

    var remove = event.target.closest("[data-orb-delete]");

    if (remove) {
      window.taskUI.confirm(remove.getAttribute("data-confirm") || "Delete the meeting?").then(function (ok) {
        if (!ok) return;

        api("DELETE", "/meetings/" + cfg.id).then(function () {
          location.href = cfg.base + "/orb/";
        }, function (err) {
          say(err.message, true);
        });
      });
      return;
    }

    var retry = event.target.closest("[data-summary-retry]");

    if (retry) {
      api("POST", "/meetings/" + cfg.id + "/summary").then(function () {
        location.reload();
      }, function (err) {
        say(err.message, true);
      });
    }
  });

  var agentPicker = room.querySelector("[data-orb-agent]");

  if (agentPicker) {
    agentPicker.addEventListener("submit", function (event) {
      event.preventDefault();
      var pick = agentPicker.querySelector("select[name=agent]");
      api("POST", "/meetings/" + cfg.id + "/agent", { agent: pick ? pick.value : "" }).then(function () {
        location.reload();
      }, function (err) {
        say(err.message, true);
      });
    });
  }

  if (el.join) {
    el.join.addEventListener("click", join);
  }

  if (el.mic) {
    el.mic.addEventListener("click", function () {
      if (!joined) {
        join();
        return;
      }

      if (micOn) {
        mute();
      } else {
        unmute();
      }
    });
  }

  renderTimes();
  setInterval(tick, 1000);
  tick();

  if (cfg.state !== "open" && el.summary) {
    loadSummary();
  }

  // Opening a room is joining it: a meeting nobody is in ends when everyone left.
  if (cfg.state === "open" && cfg.canEdit) {
    join();
  }

  // The index page opens a room and goes to it, and deletes an ended one in place.
  function initNew() {
    document.addEventListener("click", function (event) {
      var remove = event.target.closest("[data-orb-delete]");

      if (!remove) {
        return;
      }

      window.taskUI.confirm(remove.getAttribute("data-confirm") || "Delete the meeting?").then(function (ok) {
        if (!ok) return;

        var host = remove.closest("[data-api]");
        var api = host ? host.getAttribute("data-api") : "";

        fetch(api + "/meetings/" + remove.getAttribute("data-orb-delete"), {
          method: "DELETE",
          credentials: "same-origin",
        }).then(function (res) {
          if (!res.ok) {
            throw new Error("the request failed (" + res.status + ")");
          }

          window.location.reload();
        }, function (err) {
          window.alert(err.message);
        });
      });
    });

    var form = document.querySelector("[data-orb-create]");

    if (!form) {
      return;
    }

    form.addEventListener("submit", function (event) {
      event.preventDefault();

      var input = form.querySelector('input[name="title"]');
      var title = input ? input.value.trim() : "";
      var api = form.getAttribute("data-api");
      var button = form.querySelector('button[type="submit"]');

      if (button) {
        button.disabled = true;
      }

      fetch(api + "/meetings", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title: title }),
      }).then(function (res) {
        return res.text().then(function (text) {
          var data = {};

          try {
            data = JSON.parse(text);
          } catch (err) {
            data = {};
          }

          if (!res.ok) {
            throw new Error(data.error || "the request failed (" + res.status + ")");
          }

          return data;
        });
      }).then(function (data) {
        location.href = (form.getAttribute("data-base") || "") + "/orb/" + data.meeting.id;
      }, function (err) {
        if (button) {
          button.disabled = false;
        }

        window.alert(err.message);
      });
    });
  }
})();
