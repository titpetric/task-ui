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
// Send goes to the run when one is attached and takes input, and to the header
// prompt otherwise, filled and focused but never submitted, so the words are
// read before they go anywhere.
(function (TAP) {
  "use strict";

  var overlay = document.getElementById("voice-overlay");
  var orb = document.getElementById("voice-orb");
  if (!overlay || !orb || !window.taskUI) return;

  var dialog = document.getElementById("run-dialog");
  var mic = document.getElementById("run-mic");
  var said = document.getElementById("voice-said");
  var where = document.getElementById("voice-where");
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
  // canRecord is what the orb needs: a microphone and a secure page. ready is
  // the stronger claim, that this browser also transcribes on the device.
  var canRecord = !!(window.isSecureContext && navigator.mediaDevices &&
    navigator.mediaDevices.getUserMedia);
  var ready = null;     // null until probed, then true or false
  var install = false;  // the language needs fetching before the first press
  var rec = null;       // the recogniser while it runs
  var media = null; // { stream, ctx, node, frames, length }
  var raf = 0;
  var text = "";    // every final result of this session

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
  //
  // With no words there is no .txt, and that absence is the queue: the
  // transcribe workflow reads every recording in voice/ that has no .txt
  // beside it. An empty file here would take a recording out of that queue
  // and leave it transcribed by nobody.
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
    var writes = [put(base + ".wav", audio, "audio/wav")];
    if (words) writes.push(put(base + ".txt", words, "text/plain"));
    return Promise.all(writes).then(function (ok) {
      return ok.indexOf(false) < 0;
    }, function () { return false; });
  }

  // ------------------------------------------------------------------ flow

  // target picks where send puts the words. A run that is attached and takes
  // input is the terminal; anything else is the header prompt, which is on
  // every page a user may ask Claude from.
  function target() {
    if (!ready) return { kind: "keep", say: "Keeps the recording, transcribed in the artifacts" };
    if (window.taskUI.attached()) return { kind: "run", say: "Sends to the run" };
    var ask = document.getElementById("header-prompt");
    if (ask) return { kind: "prompt", el: ask, say: "Fills the prompt, yours to send" };
    return { kind: "keep", say: "Keeps the recording, transcribed in the artifacts" };
  }

  // show opens and closes the window. showModal puts it in the top layer, so
  // it draws over a run window, which is a modal dialog of its own.
  function show(on) {
    if (on) {
      if (!overlay.open) overlay.showModal();
      return;
    }
    said.textContent = "";
    where.textContent = "";
    if (overlay.open) overlay.close();
  }

  function heard(final, partial) {
    said.textContent = final;
    if (!partial) return;
    var span = document.createElement("span");
    span.className = "voice-partial";
    span.textContent = (final ? " " : "") + partial;
    said.appendChild(span);
  }

  // start runs on the press and nowhere else. install() is the first call it
  // makes, because fetching a downloadable language is refused unless it is
  // handling the user's gesture, and anything before it risks spending that.
  function start() {
    var SR = window.SpeechRecognition;
    var first = ready && install && SR.install
      ? SR.install({ langs: [lang], processLocally: true }).then(function () { install = false; })
      : Promise.resolve();
    text = "";
    show(true);
    // An empty line draws "Listening" from the stylesheet. With no recogniser
    // there is nothing to listen with, so it says what it is doing instead.
    said.textContent = install ? "Fetching the language, once" : (ready ? "" : "Recording");
    where.textContent = target().say;
    // The recogniser opens its own microphone and is what dictation is; the
    // analyser and the worklet are a second consumer of the same device, so
    // they come after and are allowed to fail. Where there is no recogniser
    // the worklet does the work alone: the press records, and the transcript
    // is the workflow's to write once the artifact lands.
    var run = function () {
      if (ready) {
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
      return listen().then(function () {
        if (!media || !media.node) throw new Error("this browser cannot record");
      });
    };
    first.then(run).catch(function (err) {
      fault(err && err.name === "NotAllowedError"
        ? "The microphone was refused. Allow it for this site and press again."
        : "Dictation could not start: " + (err && err.message ? err.message : err));
    });
  }

  // fault ends the recording and leaves the window open with the reason in
  // it. Closing on its own would take the message with it.
  function fault(why) {
    if (rec) {
      rec.onresult = null;
      rec.onerror = null;
      try { rec.stop(); } catch (e) { /* already stopped */ }
      rec = null;
    }
    release();
    media = null;
    said.textContent = why;
    where.textContent = "";
  }

  // stop ends the session. send puts the words where target named and keeps
  // the recording; discard writes nothing anywhere.
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
    var to = target();
    var done = keep(words);
    release();
    media = null;
    show(false);
    if (words && to.kind === "run") window.taskUI.paste(words);
    if (words && to.kind === "prompt") {
      to.el.value = to.el.value ? to.el.value + " " + words : words;
      to.el.focus();
      to.el.dispatchEvent(new Event("input", { bubbles: true }));
    }
    done.then(function (ok) {
      if (!ok) console.info("voice: the recording was not kept");
    });
  }

  // ------------------------------------------------------------------ bind

  function toggle() {
    if (rec || overlay.open) {
      stop(false);
      return;
    }
    start();
  }

  orb.addEventListener("click", toggle);
  if (mic) mic.addEventListener("click", toggle);
  sendBtn.addEventListener("click", function () { stop(true); });
  discardBtn.addEventListener("click", function () { stop(false); });

  // Escape closes a modal dialog on its own, and a close with a recording
  // running is a discard.
  overlay.addEventListener("cancel", function (event) {
    event.preventDefault();
    stop(false);
  });
  overlay.addEventListener("close", function () {
    if (rec || media) stop(false);
  });

  if (dialog) {
    dialog.addEventListener("run-buttons", paint);
    dialog.addEventListener("close", function () { if (rec) stop(false); });
  }
  document.addEventListener("visibilitychange", function () {
    if (document.hidden && rec) stop(false);
  });
  probe();
// document.currentScript names this tag only while the file runs, so it is
// read at the call and passed in; a callback would find it null.
})(document.currentScript && document.currentScript.src);
