// voice: dictation, transcribed on the device.
//
// The orb in the corner opens the window: the avatar at full size, the words
// under it, and two buttons. The microphone opens once and the stream feeds
// two things, an analyser whose level drives the avatar and a worklet that
// collects the samples, so the recording can be kept beside the words it came
// out as.
//
// Nothing reaches a run while you talk. Interim results are drawn in the
// window; the only write is taskUI.paste on send, which is the whole of the
// interface this file has. There is no path from the recogniser to the socket.
//
// Send goes to the run when one is attached and takes input, and to Claude
// otherwise: the window then shows what the request is doing and reads the
// reply out through a voice this device holds.
//
// A device with no recogniser sends the recording alone and the instance does
// the rest: the Orb input workflow transcribes it, answers the question in it
// and synthesises a reply the window plays. See
// docs/guide/voice.md.
(function (TAP) {
  "use strict";

  var overlay = document.getElementById("voice-overlay");
  var orb = document.getElementById("voice-orb");
  if (!overlay || !orb || !window.taskUI) return;

  var dialog = document.getElementById("run-dialog");
  var mic = document.getElementById("run-mic");
  var stage = document.getElementById("voice-stage");
  var said = document.getElementById("voice-said");
  var where = document.getElementById("voice-where");
  var sendBtn = document.getElementById("voice-send");
  var discardBtn = document.getElementById("voice-discard");
  var replayBtn = document.getElementById("voice-replay");
  var avatar = document.getElementById("voice-avatar");
  if (!stage || !replayBtn || !discardBtn) return;
  var discardIcon = discardBtn.querySelector("use");

  // RATE is the sample rate the recording is written at. 16 kHz mono is what a
  // transcriber reads: Whisper and wav2vec both resample to it, so capturing
  // more buys a later pass nothing. Raise it to 48000 to make the file worth
  // listening to, at three times the bytes.
  var RATE = 16000;

  // MAX bounds a recording at the artifact endpoint's own limit, 16 MiB, so a
  // dictation left running cannot fail its upload after the fact.
  var MAX = 16 << 20;

  // BRIEF rides with a dictated question: the reply is read aloud, and a page
  // of markdown takes minutes to speak.
  var BRIEF = "Answer in one or two sentences.";

  // FLOW is the workflow that answers a recording on a device that cannot
  // transcribe: it reads the recordings with no answer beside them, so it
  // takes no argument and the orb starts it and watches the artifacts.
  var FLOW = "Orb input";

  // STAGES name what FLOW is doing, by the step it is on.
  var STAGES = ["Transcribing", "Thinking", "Speaking"];

  var lang = document.documentElement.lang || navigator.language || "en-US";
  // canRecord is what the orb needs: a microphone and a secure page. ready is
  // the stronger claim, that this browser also transcribes on the device.
  var canRecord = !!(window.isSecureContext && navigator.mediaDevices &&
    navigator.mediaDevices.getUserMedia);
  var ready = null;     // null until probed, then true or false
  var install = false;  // the language needs fetching before the first press
  var rec = null;       // the recogniser while it runs
  var media = null;     // { stream, ctx, node } while the microphone is open
  var starting = false; // a press is waiting on the microphone to open
  var buffer = null;    // { frames, length, rate, cap } across every pause
  var raf = 0;
  var text = "";        // every final result since the last send
  var answer = "";      // the last reply, as spoken, kept for replay
  var pending = false;  // a question is out and the window is following it
  var reply = "";       // the reply as audio the instance made, for replay
  var player = null;    // the element playing that audio
  var flowID = "";      // FLOW's ID, read once
  var gen = 0;          // a send or a discard retires the watch before it
  var since = 0;        // when the request went out, for the counter
  var clock = 0;        // the interval writing the counter
  var doing = "";       // the step the counter sits beside

  // ------------------------------------------------------------------ gate

  // paint is the only writer of what is visible to press. The orb shows
  // wherever a microphone can be opened, whether or not this browser can
  // transcribe: without a recogniser the press still records, and the
  // transcript comes from the workflow that reads the artifact later. The run
  // window's own button shows on top of that when the run takes input,
  // because a modal run window puts the orb behind its backdrop.
  function paint() {
    orb.hidden = !canRecord;
    orb.title = ready ? "Dictate" : "Record";
    if (mic) mic.hidden = !(canRecord && mic.dataset.allowed);
  }

  // probe settles once whether this browser transcribes without sending the
  // audio anywhere. A language that is downloadable counts: the first press
  // installs it. Anything else records and nothing more, since the
  // alternative is sending the audio to a transcriber elsewhere.
  function probe() {
    if (ready !== null) {
      paint();
      return;
    }
    ready = false;
    paint();
    var SR = window.SpeechRecognition;
    if (!SR || !SR.available || !window.isSecureContext) {
      console.info(!SR || !SR.available
        ? "voice: no on-device speech recognition here, recording only"
        : "voice: a secure origin is needed, recording only");
      return;
    }
    SR.available({ langs: [lang], processLocally: true }).then(function (state) {
      ready = state === "available" || state === "downloadable";
      if (!ready) console.info("voice: no on-device language for " + lang + ", state " + state + ", recording only");
      install = state === "downloadable";
      paint();
    }, function () {
      paint();
    });
  }

  // ----------------------------------------------------------------- level

  // level reads the microphone each frame and sets the avatar's level, which
  // swells the orb at its centre and lets it down again in silence.
  function level(analyser) {
    var buf = new Float32Array(analyser.fftSize);
    var smoothed = 0;
    var tick = function () {
      analyser.getFloatTimeDomainData(buf);
      var sum = 0;
      for (var i = 0; i < buf.length; i++) sum += buf[i] * buf[i];
      var rms = Math.sqrt(sum / buf.length);
      // A voice sits well below full scale, so lift it and clamp; smoothing
      // keeps the glow from flickering on each frame's noise.
      var now = Math.min(1, rms * 7);
      smoothed += (now - smoothed) * (now > smoothed ? 0.5 : 0.12);
      if (window.voiceOrbs) window.voiceOrbs.level(smoothed);
      raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
  }

  // ----------------------------------------------------------------- audio

  // listen opens the microphone and keeps every sample. The worklet is the
  // file beside this one, found off this script's own src: the two front ends
  // serve the theme under different prefixes, and the page's CSP is
  // script-src 'self', which a blob: URL is not.
  // The ?v= of this script goes with it, so a changed worklet is a changed
  // URL and no browser answers it out of a cache.
  var WORKLET = TAP ? new URL("tap.js" + TAP.replace(/^[^?]*/, ""), TAP).href : "";

  // The gravity filter reads its displacement map off this script's own URL,
  // for the same reason the worklet does.
  if (TAP) {
    var field = document.getElementById("voice-well-map");
    if (field) field.setAttribute("href", new URL("well.png" + TAP.replace(/^[^?]*/, ""), TAP).href);
  }

  function listen() {
    return navigator.mediaDevices.getUserMedia({
      audio: { channelCount: 1, echoCancellation: true, noiseSuppression: true },
    }).then(function (stream) {
      var ctx = new (window.AudioContext || window.webkitAudioContext)();
      var src = ctx.createMediaStreamSource(stream);
      var analyser = ctx.createAnalyser();
      analyser.fftSize = 1024;
      src.connect(analyser);
      media = { stream: stream, ctx: ctx };
      // The buffer outlives the stream, so a pause and a resume write one
      // file. cap bounds the frames whose downsampled bytes reach MAX.
      if (!buffer) {
        buffer = {
          frames: [], length: 0, rate: ctx.sampleRate,
          cap: Math.floor((MAX / 2) * (ctx.sampleRate / RATE)),
        };
      }
      level(analyser);
      if (!WORKLET) return;
      return ctx.audioWorklet.addModule(WORKLET).then(function () {
        var node = new AudioWorkletNode(ctx, "tap");
        node.port.onmessage = function (event) {
          if (buffer && buffer.length < buffer.cap) {
            buffer.frames.push(event.data);
            buffer.length += event.data.length;
          }
        };
        src.connect(node);
        media.node = node;
      }, function (err) {
        // No worklet: the level still reads and the words still paste, and
        // the console names the recording that went nowhere.
        console.info("voice: no recording, the worklet did not load", err);
      });
    });
  }

  // release closes the microphone and leaves the buffer where it is.
  function release() {
    if (raf) cancelAnimationFrame(raf);
    raf = 0;
    if (window.voiceOrbs) window.voiceOrbs.level(0);
    if (!media) return;
    if (media.node) media.node.port.onmessage = null;
    media.stream.getTracks().forEach(function (t) { t.stop(); });
    if (media.ctx.state !== "closed") media.ctx.close();
    media = null;
  }

  // wav writes the frames as a RIFF file, mono, downsampled to RATE. Each
  // output sample is the mean of the source window behind it; one sample
  // picked out of that window would alias a voice badly.
  function wav() {
    if (!buffer || !buffer.length) return null;
    var all = new Float32Array(buffer.length);
    var at = 0;
    buffer.frames.forEach(function (f) { all.set(f, at); at += f.length; });
    var ratio = buffer.rate / RATE;
    var count = Math.floor(all.length / ratio);
    var out = new Int16Array(count);
    for (var i = 0; i < count; i++) {
      var from = Math.floor(i * ratio);
      var to = Math.min(all.length, Math.floor((i + 1) * ratio));
      var sum = 0;
      for (var j = from; j < to; j++) sum += all[j];
      var v = sum / Math.max(1, to - from);
      out[i] = Math.max(-1, Math.min(1, v)) * 0x7fff;
    }
    var buf = new ArrayBuffer(44 + out.length * 2);
    var view = new DataView(buf);
    var ascii = function (off, s) { for (var k = 0; k < s.length; k++) view.setUint8(off + k, s.charCodeAt(k)); };
    ascii(0, "RIFF");
    view.setUint32(4, 36 + out.length * 2, true);
    ascii(8, "WAVEfmt ");
    view.setUint32(16, 16, true);
    view.setUint16(20, 1, true);
    view.setUint16(22, 1, true);
    view.setUint32(24, RATE, true);
    view.setUint32(28, RATE * 2, true);
    view.setUint16(32, 2, true);
    view.setUint16(34, 16, true);
    ascii(36, "data");
    view.setUint32(40, out.length * 2, true);
    new Int16Array(buf, 44).set(out);
    return new Blob([buf], { type: "audio/wav" });
  }

  // -------------------------------------------------------------- sanitise

  // NAMED is every character that is not a letter, a digit, a space, an
  // apostrophe or a hyphen, written as the word for it. The transcript is
  // English, not shell: a recogniser puts in punctuation nobody spoke, and a
  // newline would press Return on a line the reader never saw.
  var NAMED = {
    ",": "comma", ".": "dot", "!": "exclamation", "?": "question",
    "#": "hash", "^": "caret", "`": "backtick", "$": "dollar",
    "|": "pipe", ";": "semicolon", "&": "ampersand", ":": "colon",
    "<": "less", ">": "greater", "(": "paren", ")": "paren",
    "{": "brace", "}": "brace", "[": "bracket", "]": "bracket",
    "*": "star", "~": "tilde", "/": "slash", "\\": "backslash",
    '"': "quote", "=": "equals", "%": "percent", "@": "at", "+": "plus",
  };

  function sanitize(raw) {
    var out = "";
    // NFD splits an accented letter into its base and a combining mark, so an
    // accent the recogniser writes costs the mark and keeps the letter under
    // it. A letter with no ASCII base, and the mark itself, still go.
    var flat = String(raw || "")
      .normalize("NFD")
      .replace(/\p{M}/gu, "")
      .replace(/\s+/g, " ");
    for (var i = 0; i < flat.length; i++) {
      var ch = flat[i];
      var code = flat.charCodeAt(i);
      if (code < 0x20 || code === 0x7f) continue;
      if (/[A-Za-z0-9 '-]/.test(ch)) {
        out += ch;
        continue;
      }
      if (NAMED[ch]) out += "[" + NAMED[ch] + "]";
    }
    return out.replace(/\s+/g, " ").trim();
  }

  // ----------------------------------------------------------------- store

  // file returns the API path of a file of the artifact tree. The API answers
  // under one mount for both front ends, so the path is absolute.
  function file(rel) {
    return "/api/artifacts/files/" + rel;
  }

  // put writes one file into the artifact tree; false when the write failed.
  function put(rel, body, type) {
    return fetch(file(rel), {
      method: "PUT",
      credentials: "same-origin",
      headers: { "Content-Type": type },
      body: body,
    }).then(function (res) { return res.ok; }, function () { return false; });
  }

  // read returns a file of the artifact tree as the API describes it, or null
  // when it is not there yet, so the window can wait for a reply.
  function read(rel) {
    return fetch(file(rel), { credentials: "same-origin" })
      .then(function (res) { return res.ok ? res.json() : null; }, function () { return null; });
  }

  // keep writes the recording and the words beside each other in the artifact
  // tree, which versions them and ages them out on its own schedule. It
  // returns the path the pair shares, or empty when a write failed: a reader
  // without artifact/edit dictates all the same.
  //
  // With no words there is no .txt, and that absence is the queue FLOW reads.
  // An empty file here would take a recording out of it and leave it
  // transcribed by nobody.
  function keep(words) {
    var audio = wav();
    if (!audio) return Promise.resolve("");
    var stamp = new Date().toISOString().replace(/[:.]/g, "-").replace("Z", "");
    var base = "voice/" + stamp;
    var writes = [put(base + ".wav", audio, "audio/wav")];
    if (words) writes.push(put(base + ".txt", words, "text/plain"));
    return Promise.all(writes).then(function (ok) {
      return ok.indexOf(false) < 0 ? base : "";
    }, function () { return ""; });
  }

  // ------------------------------------------------------------------ flow

  // target picks where send puts the words. A run that is attached and takes
  // input is the terminal; anything else is the header prompt, which is on
  // every page a user may ask Claude from.
  function target() {
    if (window.taskUI.attached()) return { kind: "run", say: "Sends to the run" };
    if (document.getElementById("header-prompt")) return { kind: "prompt", say: "Asks Claude, and reads the answer out" };
    return { kind: "keep", say: "Keeps the recording in the artifacts" };
  }

  // show opens and closes the window. showModal puts it in the top layer, so
  // it draws over a run window, which is a modal dialog of its own. The orbs
  // draw only while the window is open: nothing of them is seen after.
  function show(on) {
    if (on) {
      if (!overlay.open) overlay.showModal();
      if (window.voiceOrbs) window.voiceOrbs.start(avatar);
      return;
    }
    said.textContent = "";
    where.textContent = "";
    if (window.voiceOrbs) window.voiceOrbs.stop();
    if (overlay.open) overlay.close();
  }

  // state is the window's one source of truth: idle, recording or paused. A
  // press counts as recording while the microphone opens, so the window
  // answers the press and not the device.
  function state() {
    if (media || starting) return "recording";
    return buffer || text ? "paused" : "idle";
  }

  // dress writes the state onto the window and says what the orb will do.
  function dress(note) {
    var now = state();
    overlay.dataset.state = now;
    stage.setAttribute("aria-pressed", now === "recording" ? "true" : "false");
    stage.title = now === "recording" ? "Pause" : (now === "paused" ? "Resume" : "Record");
    stage.setAttribute("aria-label", stage.title);
    sendBtn.disabled = now === "idle";
    // The left button is whatever there is to undo: the recording, then the
    // answer it brought back, then the window.
    var shut = now === "idle";
    var told = shut && !!(answer || reply);
    discardIcon.setAttribute("href", !shut ? "#i-trash" : (told ? "#i-back" : "#i-x"));
    discardBtn.title = !shut ? "Discard" : (told ? "Back" : "Close");
    discardBtn.setAttribute("aria-label", discardBtn.title);
    discardBtn.classList.toggle("shut", shut);
    if (window.voiceOrbs) {
      window.voiceOrbs.mood(now === "recording" ? "recording" : (told ? "answered" : "idle"));
    }
    // A question in flight owns the line; the state does not take it back.
    if (!pending) where.textContent = now === "idle" ? "" : target().say;
    if (note !== undefined) {
      said.textContent = note;
      return;
    }
    if (now === "idle" && !pending && !answer) said.textContent = "";
    if (now === "paused" && !text) said.textContent = "Paused";
  }

  function heard(final, partial) {
    said.textContent = final;
    if (!partial) return;
    var span = document.createElement("span");
    span.className = "voice-partial";
    span.textContent = (final ? " " : "") + partial;
    said.appendChild(span);
  }

  // open shows the window with nothing recording; the orb starts it.
  function open() {
    show(true);
    dress();
  }

  // record starts or resumes capture into the same buffer. install() is the
  // first call it makes, because fetching a downloadable language is refused
  // unless it is handling the user's gesture.
  function record() {
    if (starting || media) return;
    var SR = window.SpeechRecognition;
    var first = ready && install && SR.install
      ? SR.install({ langs: [lang], processLocally: true }).then(function () { install = false; })
      : Promise.resolve();
    starting = true;
    dress(install ? "Fetching the language, once" : (text || (ready ? "" : "Recording")));
    first.then(listening).then(function () {
      starting = false;
      dress(text || undefined);
    }).catch(function (err) {
      starting = false;
      refused(err);
    });
  }

  // pause closes the microphone and keeps everything heard and captured.
  function pause() {
    starting = false;
    if (rec) {
      rec.onresult = null;
      rec.onerror = null;
      try { rec.stop(); } catch (e) { /* already stopped */ }
      rec = null;
    }
    release();
    dress();
  }

  // listening opens the recogniser and the microphone, and returns once both
  // are running or the microphone alone where there is no recogniser.
  //
  // The recogniser opens its own microphone and is what dictation is; the
  // analyser and the worklet are a second consumer of the same device, so
  // they come after and are allowed to fail.
  function listening() {
    var SR = window.SpeechRecognition;
    if (!ready) {
      return listen().then(function () {
        if (!media || !media.node) throw new Error("this browser cannot record");
      }).catch(refused);
    }
    rec = new SR();
    rec.lang = lang;
    rec.continuous = true;
    rec.interimResults = true;
    rec.processLocally = true;
    rec.onresult = function (event) {
      var partial = "";
      for (var i = event.resultIndex; i < event.results.length; i++) {
        var r = event.results[i];
        if (r.isFinal) text += (text ? " " : "") + r[0].transcript.trim();
        else partial += r[0].transcript;
      }
      heard(text, partial.trim());
    };
    rec.onerror = function (event) {
      fault(event.error === "not-allowed"
        ? "The microphone was refused. Allow it for this site and press again."
        : "Dictation stopped: " + event.error);
    };
    rec.start();
    return listen().catch(function (err) {
      console.info("voice: no level and no recording", err);
    });
  }

  // refused reports a microphone the browser or the reader would not open.
  function refused(err) {
    fault(err && err.name === "NotAllowedError"
      ? "The microphone was refused. Allow it for this site and press again."
      : "Dictation could not start: " + (err && err.message ? err.message : err));
  }

  // fault pauses and leaves the window open with the reason in it, because
  // closing on its own would take the message with it.
  function fault(why) {
    pause();
    dress(why);
  }

  // label returns a thread's state as a line, capitalised as the window reads.
  function label(word) {
    var say = String(word || "Working");
    return say.charAt(0).toUpperCase() + say.slice(1);
  }

  // ----------------------------------------------------------------- clock

  // span is a number of seconds as the window writes it.
  function span(secs) {
    if (secs < 60) return secs + "s";
    return Math.floor(secs / 60) + "m" + ("0" + (secs % 60)).slice(-2) + "s";
  }

  // tell writes the step and how long it has been going. A transcription on a
  // busy machine takes a minute, and a line that does not move reads as stuck.
  function tell(now) {
    if (now !== undefined) doing = now;
    where.textContent = doing + " · " + span(Math.round((Date.now() - since) / 1000));
  }

  // ticking starts the counter beside a step; ended stops it and leaves how
  // long the whole thing took on the line.
  function ticking(now) {
    since = Date.now();
    if (clock) window.clearInterval(clock);
    clock = window.setInterval(function () { tell(); }, 1000);
    tell(now);
  }

  function ended(word) {
    var secs = Math.round((Date.now() - since) / 1000);
    stopped();
    where.textContent = since ? word + " · " + span(secs) : word;
  }

  function stopped() {
    if (clock) window.clearInterval(clock);
    clock = 0;
  }

  // asked puts the question to Claude and follows it in the window, so the
  // orb says what the request is doing and reads the reply out when it ends.
  // The reply goes beside the recording at, once it is written, so the pair
  // reads the same whichever path transcribed it.
  function asked(words, at) {
    var mine = ++gen;
    answer = "";
    reply = "";
    pending = true;
    replayBtn.hidden = true;
    if (!window.taskUI.ask) {
      pending = false;
      var box = document.getElementById("header-prompt");
      if (box) {
        box.value = box.value ? box.value + " " + words : words;
        box.dispatchEvent(new Event("input", { bubbles: true }));
      }
      return;
    }
    ticking("Asking");
    window.taskUI.ask(words + " " + BRIEF, function (update) {
      if (mine !== gen) return;
      if (update.state === "answered") {
        pending = false;
        answer = plain(update.body);
        said.textContent = answer.slice(0, 400);
        ended("Answered");
        speak(answer);
        at.then(function (base) { if (base) put(base + ".reply.txt", answer, "text/plain"); });
        return;
      }
      if (update.state === "failed") {
        pending = false;
        var why = update.text && update.text !== "failed" ? update.text : "";
        stopped();
        where.textContent = why || "The request failed";
        return;
      }
      if (update.state === "done") return;
      tell(label(update.text || update.state));
    });
  }

  // ------------------------------------------------------------------ flow

  // flow returns FLOW's ID, or empty where the instance has no such workflow.
  function flow() {
    if (flowID) return Promise.resolve(flowID);
    return fetch("/api/workflows", { credentials: "same-origin" })
      .then(function (res) { return res.ok ? res.json() : null; })
      .then(function (data) {
        ((data && data.workflows) || []).forEach(function (wf) {
          if (wf.name === FLOW) flowID = wf.id;
        });
        return flowID;
      });
  }

  // handed starts FLOW over the recording at and watches for the reply it
  // writes beside it, which is the whole of a device that cannot transcribe.
  function handed(at) {
    var mine = ++gen;
    answer = "";
    reply = "";
    pending = true;
    replayBtn.hidden = true;
    said.textContent = "";
    ticking(STAGES[0]);
    flow().then(function (id) {
      if (!id) throw new Error("this instance has no " + FLOW + " workflow");
      return fetch("/api/workflows/" + encodeURIComponent(id) + "/runs", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: "{}",
      }).then(function (res) {
        if (res.status === 403) throw new Error("you may not run " + FLOW);
        return res.json().then(function (body) { return { id: id, run: body.run_id || body.id }; });
      });
    }).then(function (started) {
      if (mine === gen) watch(mine, started.id, started.run, at);
    }).catch(function (err) { if (mine === gen) fell(err); });
  }

  // watch polls for the reply FLOW writes beside the recording, then for the
  // audio of it, and says which step the run is on while neither is there.
  // It stops after TRIES polls, a transcription and a reply over a cold
  // whisper build with room to spare.
  function watch(mine, id, runID, at) {
    var TRIES = 200;
    var tries = 0;
    var tick = function () {
      if (mine !== gen) return;
      if (++tries > TRIES) {
        done(mine);
        return;
      }
      var again = function () { window.setTimeout(tick, 3000); };
      if (!answer) {
        read(at + ".reply.txt").then(function (data) {
          if (mine !== gen) return;
          if (data && data.text) {
            answer = data.text.trim();
            said.textContent = answer;
            tell(STAGES[2]);
            again();
            return;
          }
          return onStep(id, runID).then(again);
        }).catch(function (err) { if (mine === gen) fell(err); });
        return;
      }
      read(at + ".reply.wav").then(function (data) {
        if (mine !== gen) return;
        if (!data) {
          again();
          return;
        }
        reply = file(at + ".reply.wav") + "?raw";
        done(mine);
      });
    };
    tick();
  }

  // onStep writes the step the run is on onto the window, and fails the watch
  // when the run ended with nothing beside the recording.
  function onStep(id, runID) {
    return fetch("/api/workflows/" + encodeURIComponent(id) + "/runs", { credentials: "same-origin" })
      .then(function (res) { return res.ok ? res.json() : null; })
      .then(function (data) {
        var run = ((data && data.runs) || []).filter(function (r) { return r.id === runID; })[0];
        if (!run) return;
        var steps = run.steps || [];
        for (var i = 0; i < steps.length; i++) {
          if (steps[i].state === "running") {
            tell(STAGES[i] || "Working");
            return;
          }
        }
        if (run.state === "running") return;
        if (run.state !== "ok") throw new Error(FLOW + " " + run.state + (run.message ? ": " + run.message : ""));
      });
  }

  // done ends the watch with whatever arrived and plays it.
  function done(mine) {
    if (mine !== gen) return;
    pending = false;
    replayBtn.hidden = !answer && !reply;
    if (!answer && !reply) {
      ended("No answer came back");
      return;
    }
    ended("Answered");
    play();
  }

  // fell leaves the reason in the window and in a toast, since a phone shows
  // the window for as long as it is open and the toast after that.
  function fell(err) {
    var why = (err && err.message) || String(err);
    pending = false;
    stopped();
    where.textContent = why;
    say("The recording was kept but not answered: " + why, "bad");
  }

  // play plays the reply the instance synthesised, and reads the words out
  // on this device where it synthesised none.
  function play() {
    if (!reply) {
      speak(answer);
      return;
    }
    if (!player) player = new Audio();
    player.src = reply;
    var started = player.play();
    if (started && started.catch) started.catch(function () { speak(answer); });
  }

  // drop lets go of the reply and the words, and retires the watch that
  // would have written them, so nothing draws into a window that moved on.
  function drop() {
    gen++;
    stopped();
    since = 0;
    answer = "";
    reply = "";
    pending = false;
    replayBtn.hidden = true;
    if (player) player.pause();
    if (window.speechSynthesis) window.speechSynthesis.cancel();
  }

  // trash drops the recording, the words and the reply, and closes the
  // window.
  function trash() {
    drop();
    pause();
    buffer = null;
    text = "";
    show(false);
  }

  // back steps back through what the window holds: a recording, then a reply,
  // then the window itself.
  function back() {
    if (state() !== "idle") {
      trash();
      return;
    }
    if (!answer && !reply) {
      trash();
      return;
    }
    drop();
    said.textContent = "";
    where.textContent = "";
    dress();
  }

  // send delivers the words, writes the file and leaves the window open on
  // an empty buffer, so the next sentence is one press away. A recording with
  // no words of its own goes to FLOW, which answers it on the instance.
  function send() {
    if (state() === "idle") return;
    var words = sanitize(text);
    var to = target();
    var at = keep(words);
    pause();
    buffer = null;
    text = "";
    dress();
    if (words && to.kind === "run") window.taskUI.paste(words);
    if (words && to.kind === "prompt") asked(words, at);
    at.then(function (base) {
      say(base ? "Recording kept in artifacts, under voice/" : "The recording was not kept", base ? "ok" : "bad");
      if (!words && base) handed(base);
    });
  }

  // say reports through the page's toasts, which is the only report a phone
  // ever sees: the window has closed and nobody opens a console there.
  function say(text, tone) {
    if (window.taskUI.toast) window.taskUI.toast(text, tone);
    else console.info("voice: " + text);
  }

  // ----------------------------------------------------------------- speak

  // localVoice returns a voice this device holds, or null: a network voice
  // would send the reply to a reader elsewhere.
  function localVoice() {
    var all = window.speechSynthesis ? window.speechSynthesis.getVoices() : [];
    var here = all.filter(function (v) { return v.localService; });
    var base = lang.slice(0, 2).toLowerCase();
    return here.filter(function (v) { return (v.lang || "").toLowerCase().indexOf(base) === 0; })[0] || here[0] || null;
  }

  // plain strips the markdown out of a reply, since a reader speaks every
  // backtick and hash out loud.
  function plain(body) {
    return String(body || "")
      .replace(/```[\s\S]*?```/g, " code block ")
      .replace(/`([^`]*)`/g, "$1")
      .replace(/!?\[([^\]]*)\]\([^)]*\)/g, "$1")
      .replace(/^[#>\s-]*/gm, "")
      .replace(/[*_~]/g, "")
      .replace(/\s+/g, " ")
      .trim();
  }

  // speak reads a reply aloud and keeps it for replay.
  function speak(body) {
    answer = plain(body);
    replayBtn.hidden = !answer;
    if (!answer || !window.speechSynthesis || !window.SpeechSynthesisUtterance) return;
    var pick = localVoice();
    if (!pick) {
      console.info("voice: no on-device voice, the answer is not read aloud");
      return;
    }
    window.speechSynthesis.cancel();
    var line = new SpeechSynthesisUtterance(answer.slice(0, 1200));
    line.voice = pick;
    line.lang = pick.lang || lang;
    line.rate = 1.02;
    window.speechSynthesis.speak(line);
  }

  // ------------------------------------------------------------------ bind

  // The orb in the corner opens and closes the window; the orb in the middle
  // of it records, pauses and resumes.
  function visit() {
    if (overlay.open) trash();
    else open();
  }

  orb.addEventListener("click", visit);
  if (mic) mic.addEventListener("click", visit);
  stage.addEventListener("click", function () {
    if (window.voiceOrbs) window.voiceOrbs.tap();
    if (media || starting) pause();
    else record();
  });
  sendBtn.addEventListener("click", send);
  discardBtn.addEventListener("click", back);
  replayBtn.addEventListener("click", function () { if (answer || reply) play(); });

  overlay.addEventListener("cancel", function (event) {
    event.preventDefault();
    trash();
  });
  overlay.addEventListener("close", trash);

  if (dialog) {
    dialog.addEventListener("run-buttons", paint);
    dialog.addEventListener("close", function () { if (overlay.open) trash(); });
  }
  // A page sent to the background keeps the words and stops the microphone.
  document.addEventListener("visibilitychange", function () {
    if (document.hidden && media) pause();
  });
  probe();
// document.currentScript names this tag only while the file runs, so it is
// read at the call and passed in; a callback would find it null.
})(document.currentScript && document.currentScript.src);
