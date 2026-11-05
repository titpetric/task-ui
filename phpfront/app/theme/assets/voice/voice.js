// voice: dictation into an attached run, transcribed on the device.
//
// The microphone opens once. The stream feeds two things: an analyser, whose
// level drives the avatar, and a worklet that collects the samples so the
// recording can be kept beside the words it came out as.
//
// Nothing reaches the run while you talk. Interim results are drawn in the
// overlay; the only write is taskUI.paste on send, which is the whole of the
// interface this file has. There is no path from the recogniser to the socket.
(function (TAP) {
  "use strict";

  var dialog = document.getElementById("run-dialog");
  var overlay = document.getElementById("voice-overlay");
  var mic = document.getElementById("run-mic");
  if (!dialog || !overlay || !mic || !window.taskUI) return;

  var said = document.getElementById("voice-said");
  var sendBtn = document.getElementById("voice-send");
  var discardBtn = document.getElementById("voice-discard");

  // RATE is the sample rate the recording is written at. 16 kHz mono is what a
  // transcriber reads: Whisper and wav2vec both resample to it, so capturing
  // more buys a later pass nothing. Raise it to 48000 to make the file worth
  // listening to, at three times the bytes.
  var RATE = 16000;

  // MAX bounds a recording at the artifact endpoint's own limit, 16 MiB, so a
  // dictation left running cannot fail its upload after the fact.
  var MAX = 16 << 20;

  var lang = document.documentElement.lang || navigator.language || "en-US";
  var ready = null; // null until asked, then true or false
  var rec = null;   // the recogniser while it runs
  var media = null; // { stream, ctx, node, frames, length }
  var raf = 0;
  var text = "";    // every final result of this session

  // ------------------------------------------------------------------ gate

  // paint is the only writer of the button's visibility, and it needs two
  // facts: whether the run takes input, which task-ui.js publishes as
  // data-allowed, and whether this browser transcribes on the device.
  function paint() {
    mic.hidden = !(ready && mic.dataset.allowed);
  }

  // probe settles once whether this browser transcribes without sending the
  // audio anywhere. A language that is downloadable counts: the first press
  // installs it. Anything else leaves the button hidden for good, since the
  // alternative is sending the audio to a transcriber elsewhere.
  function probe() {
    if (ready !== null) {
      paint();
      return;
    }
    ready = false;
    var SR = window.SpeechRecognition;
    if (!SR || !SR.available || !window.isSecureContext) {
      mic.title = !SR || !SR.available
        ? "On-device speech is not available in this browser"
        : "Dictation needs a secure origin";
      paint();
      return;
    }
    SR.available({ langs: [lang], processLocally: true }).then(function (state) {
      ready = state === "available" || state === "downloadable";
      if (!ready) mic.title = "The on-device language for " + lang + " is not available";
      mic.dataset.install = state === "downloadable" ? "1" : "";
      paint();
    }, function () {
      paint();
    });
  }

  // ----------------------------------------------------------------- level

  // level reads the microphone and writes it to the overlay as --voice-level,
  // which every layer of the avatar is sized and lit by.
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
      overlay.style.setProperty("--voice-level", smoothed.toFixed(3));
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

  function listen() {
    return navigator.mediaDevices.getUserMedia({
      audio: { channelCount: 1, echoCancellation: true, noiseSuppression: true },
    }).then(function (stream) {
      var ctx = new (window.AudioContext || window.webkitAudioContext)();
      var src = ctx.createMediaStreamSource(stream);
      var analyser = ctx.createAnalyser();
      analyser.fftSize = 1024;
      src.connect(analyser);
      // cap is the frame count whose downsampled bytes reach MAX, so the
      // bound is on the file that gets written and not on the samples the
      // device happens to deliver it in.
      media = {
        stream: stream, ctx: ctx, frames: [], length: 0, rate: ctx.sampleRate,
        cap: Math.floor((MAX / 2) * (ctx.sampleRate / RATE)),
      };
      level(analyser);
      if (!WORKLET) return;
      return ctx.audioWorklet.addModule(WORKLET).then(function () {
        var node = new AudioWorkletNode(ctx, "tap");
        node.port.onmessage = function (event) {
          if (media && media.length < media.cap) {
            media.frames.push(event.data);
            media.length += event.data.length;
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

  function release() {
    if (raf) cancelAnimationFrame(raf);
    raf = 0;
    overlay.style.setProperty("--voice-level", "0");
    if (!media) return;
    if (media.node) media.node.port.onmessage = null;
    media.stream.getTracks().forEach(function (t) { t.stop(); });
    if (media.ctx.state !== "closed") media.ctx.close();
  }

  // wav writes the frames as a RIFF file, mono, downsampled to RATE. Each
  // output sample is the mean of the source window behind it; one sample
  // picked out of that window would alias a voice badly.
  function wav() {
    if (!media || !media.length) return null;
    var all = new Float32Array(media.length);
    var at = 0;
    media.frames.forEach(function (f) { all.set(f, at); at += f.length; });
    var ratio = media.rate / RATE;
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

  // keep writes the recording and the words beside each other in the artifact
  // tree, which versions them and ages them out on its own schedule. A reader
  // without artifact/edit dictates all the same, and the console says the
  // recording went nowhere.
  function keep(words) {
    var audio = wav();
    if (!audio) return Promise.resolve(false);
    var stamp = new Date().toISOString().replace(/[:.]/g, "-").replace("Z", "");
    // The API answers under one mount for both front ends, so the path is
    // absolute and takes no prefix.
    var base = "/api/artifacts/files/voice/" + stamp;
    var put = function (path, body, type) {
      return fetch(path, {
        method: "PUT",
        credentials: "same-origin",
        headers: { "Content-Type": type },
        body: body,
      }).then(function (res) { return res.ok; });
    };
    return Promise.all([
      put(base + ".wav", audio, "audio/wav"),
      put(base + ".txt", words, "text/plain"),
    ]).then(function (ok) { return ok[0] && ok[1]; }, function () { return false; });
  }

  // ------------------------------------------------------------------ flow

  function show(on) {
    overlay.hidden = !on;
    mic.setAttribute("aria-pressed", on ? "true" : "false");
    if (!on) said.textContent = "";
  }

  function heard(final, partial) {
    said.textContent = final;
    if (!partial) return;
    var span = document.createElement("span");
    span.className = "voice-partial";
    span.textContent = (final ? " " : "") + partial;
    said.appendChild(span);
  }

  function start() {
    var SR = window.SpeechRecognition;
    text = "";
    show(true);
    said.textContent = "";
    var run = function () {
      return listen().then(function () {
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
          said.textContent = event.error === "not-allowed"
            ? "The microphone was refused."
            : "Dictation stopped: " + event.error;
          stop(false);
        };
        rec.start();
      });
    };
    var first = mic.dataset.install && SR.install
      ? SR.install({ langs: [lang], processLocally: true }).then(function () { mic.dataset.install = ""; })
      : Promise.resolve();
    said.textContent = mic.dataset.install ? "Fetching the language, once" : "";
    first.then(run).catch(function (err) {
      said.textContent = "Dictation could not start: " + (err && err.message ? err.message : err);
      window.setTimeout(function () { stop(false); }, 2500);
    });
  }

  // stop ends the session. send writes the words into the run and keeps the
  // recording; discard writes nothing anywhere.
  function stop(send) {
    var words = sanitize(text);
    if (rec) {
      rec.onresult = null;
      rec.onerror = null;
      try { rec.stop(); } catch (e) { /* already stopped */ }
      rec = null;
    }
    if (!send) {
      release();
      media = null;
      show(false);
      return;
    }
    if (words) window.taskUI.paste(words);
    var done = keep(words);
    release();
    media = null;
    show(false);
    done.then(function (ok) {
      if (!ok) console.info("voice: the recording was not kept");
    });
  }

  // ------------------------------------------------------------------ bind

  mic.addEventListener("click", function () {
    if (rec || !overlay.hidden) {
      stop(false);
      return;
    }
    if (!window.taskUI.attached()) return;
    start();
  });
  sendBtn.addEventListener("click", function () { stop(true); });
  discardBtn.addEventListener("click", function () { stop(false); });

  dialog.addEventListener("run-open", function () {
    show(false);
    probe();
  });
  dialog.addEventListener("run-buttons", paint);
  dialog.addEventListener("close", function () { stop(false); });
  document.addEventListener("visibilitychange", function () {
    if (document.hidden && rec) stop(false);
  });
// document.currentScript names this tag only while the file runs, so it is
// read at the call and passed in; a callback would find it null.
})(document.currentScript && document.currentScript.src);
