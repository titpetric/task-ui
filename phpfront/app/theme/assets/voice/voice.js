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
// Everything but a run goes to one prompt thread, which is the session: the
// first clip opens it and every later one replies in it, so Claude reads the
// conversation so far and holds the task tools to do what is asked with. The
// window follows the thread over the notes event stream for as long as it is
// open, so a reply Claude writes after the work it started finishes arrives
// and plays on its own. A device with no recogniser sends the clip alone and
// the session transcribes it. See docs/guide/voice.md.
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

  // BRIEF opens the session, once, with the first thing said into it. The
  // thread that follows is one conversation: Claude reads everything said in
  // it so far, holds the task tools, and may reply again long after the
  // first answer. A later reply is how finished work reaches the window.
  var BRIEF = [
    "You are answering by voice, in a window that plays what you write here.",
    "",
    "Every reply: two or three spoken sentences, plain words, no markdown, no lists, no code.",
    "You hold the task tools. Run jobs, start workflows and read artifacts to answer or to do what is asked.",
    "Work that takes a while: leave one sentence with note_comment saying what you started, do the work,",
    "then answer with what came of it. Both reach the window, the second whenever it is ready.",
    "After each reply, write the same sentences to <base>.N.reply.txt with write_artifact, N counting",
    "your replies in this thread from 1, then run the speak job of titpetric/stt so the window has the",
    "audio. The window plays <base>.N.reply.wav and nothing else, so a reply without one is never heard.",
  ].join("\n");

  // HEARD rides with a clip this device could not transcribe: the words it
  // holds are the question, so the session reads them first.
  var HEARD = "The clip is at <clip>.wav. Transcribe it with the transcribe-audio workflow and answer what it asks.";

  // STAGES name what the thread is doing while nothing has come back.
  var STAGES = ["Transcribing", "Thinking", "Speaking"];

  // TRIES is how many times a reply's audio is looked for, half a second
  // apart: piper takes seconds over a sentence, and a cold voice download
  // takes longer the first time.
  var TRIES = 120;

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
  var thread = "";      // the session, one thread for as long as the window is open
  var stamp = "";       // the clip the session opened on, which names its audio
  var heard = 0;        // how many replies of the thread have been played
  var source = null;    // the event stream the thread is followed over
  var gen = 0;          // a send or a discard retires the watch before it
  var since = 0;        // when the request went out, for the counter
  var clock = 0;        // the interval writing the counter
  var doing = "";       // the step the counter sits beside
  var failed = false;   // the run ended and left nothing beside the recording

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
  // without artifact/edit dictates all the same. With no words there is no
  // .txt, and the clip alone is what the session is asked to transcribe.
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

  // FLIGHT is the formation each step of a run puts the orbs into, keyed by
  // the word the window writes beside the counter.
  var FLIGHT = {
    Asking: "thinking",
    Transcribing: "transcribing",
    Thinking: "thinking",
    Speaking: "speaking",
    Answered: "answered",
    Working: "thinking",
  };

  // shape tells the orbs which formation the window is in. Every path the
  // window can be on ends here: a press, a step of the run, an answer, a
  // failure.
  function shape() {
    if (!window.voiceOrbs) return;
    var now = state();
    if (now === "recording") return window.voiceOrbs.mood("recording");
    if (answer || reply) return window.voiceOrbs.mood("answered");
    if (failed) return window.voiceOrbs.mood("failed");
    if (pending) return window.voiceOrbs.mood(FLIGHT[doing] || "thinking");
    if (now === "paused" || buffer) return window.voiceOrbs.mood("paused");
    window.voiceOrbs.mood("idle");
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
    shape();
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
    if (now !== undefined) {
      doing = now;
      shape();
    }
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

  // ------------------------------------------------------------------ flow

  // asked puts something to the session: the first thing opens the thread
  // with BRIEF, every later one is a reply in it, so Claude reads the whole
  // conversation and keeps the tools it ran the last thing with.
  function asked(words, clip) {
    if (!window.taskUI.ask) {
      var box = document.getElementById("header-prompt");
      if (box) {
        box.value = box.value ? box.value + " " + words : words;
        box.dispatchEvent(new Event("input", { bubbles: true }));
      }
      return;
    }
    var mine = ++gen;
    answer = "";
    reply = "";
    pending = true;
    failed = false;
    replayBtn.hidden = true;
    said.textContent = "";
    ticking(clip && !words ? STAGES[0] : "Asking");
    var body = words;
    if (clip && !words) body = HEARD.replace("<clip>", clip);
    else if (clip) body = words + "\n\n" + HEARD.replace("<clip>", clip);
    if (!thread) {
      stamp = clip || "voice/" + new Date().toISOString().replace(/[:.]/g, "-").replace("Z", "");
      body = body + "\n\n" + BRIEF.replace("<base>", stamp);
    }
    window.taskUI.ask(body, function (update) {
      if (mine !== gen) return;
      if (update.thread) {
        thread = update.thread;
        listen();
        return;
      }
      if (update.state === "answered") {
        landed(update.body);
        return;
      }
      if (update.state === "failed") {
        fell(new Error(update.text && update.text !== "failed" ? update.text : "The request failed"));
        return;
      }
      if (update.state === "done") return;
      tell(label(update.text || update.state));
    }, thread);
  }

  // listen follows the thread for as long as the window is open, so a reply
  // Claude writes after the work it started finishes reaches the window on
  // its own. It is one stream for the session.
  function listen() {
    if (source || !window.EventSource || !thread) return;
    source = new EventSource("/api/sse/notes");
    source.addEventListener("note", function (event) {
      var note = {};
      try { note = JSON.parse(event.data); } catch (e) { return; }
      if (note.id !== thread && note.parent_id !== thread) return;
      if (note.state === "working" || note.state === "open") {
        if (!answer) tell(STAGES[1]);
        return;
      }
      pull();
    });
  }

  // hush drops the stream, for a window that closed or a session that ended.
  function hush() {
    if (source) source.close();
    source = null;
  }

  // pull reads the thread and takes the newest reply it has not played yet.
  // A reply that arrives with no state change still lands through it.
  function pull() {
    var mine = gen;
    var place = window.taskUI.place ? window.taskUI.place() : null;
    if (!place) return Promise.resolve();
    var url = place.run
      ? "/api/prompts?limit=20&run=" + encodeURIComponent(place.run)
      : "/api/prompts?limit=20&folder=" + encodeURIComponent(place.folder) +
        (place.dir ? "&dir=" + encodeURIComponent(place.dir) : "");
    return fetch(url, { credentials: "same-origin" })
      .then(function (res) { return res.ok ? res.json() : null; })
      .then(function (data) {
        if (mine !== gen) return;
        var mineThread = ((data && data.notes) || []).filter(function (n) { return n.id === thread; })[0];
        if (!mineThread) return;
        var replies = mineThread.replies || [];
        if (replies.length <= heard) return;
        heard = replies.length;
        landed(replies[replies.length - 1].body);
      })
      .catch(function () {});
  }

  // landed shows a reply, looks for the audio of it beside the clip, and
  // plays whichever arrives. The count in the name is the reply's place in
  // the thread, the name BRIEF names.
  function landed(body) {
    var mine = gen;
    var words = plain(body);
    if (!words || words === answer) return;
    pending = false;
    answer = words;
    said.textContent = answer.slice(0, 400);
    if (heard < 1) heard = 1;
    var base = stamp + "." + heard;
    put(base + ".reply.txt", answer, "text/plain");
    tell(STAGES[2]);
    audio(mine, base, 0);
  }

  // audio waits for the wav of a reply and plays it, and reads the words out
  // on this device when none arrives.
  function audio(mine, base, tries) {
    if (mine !== gen) return;
    if (tries > TRIES) {
      ended("Answered");
      speak(answer);
      replayBtn.hidden = !answer;
      return;
    }
    read(base + ".reply.wav").then(function (data) {
      if (mine !== gen) return;
      if (!data) {
        window.setTimeout(function () { audio(mine, base, tries + 1); }, 500);
        return;
      }
      reply = file(base + ".reply.wav") + "?raw";
      ended("Answered");
      replayBtn.hidden = false;
      shape();
      play();
    });
  }

  // fell leaves the reason in the window and in a toast, since a phone shows
  // the window for as long as it is open and the toast after that.
  function fell(err) {
    var why = (err && err.message) || String(err);
    pending = false;
    failed = true;
    stopped();
    shape();
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
    failed = false;
    doing = "";
    replayBtn.hidden = true;
    if (player) player.pause();
    if (window.speechSynthesis) window.speechSynthesis.cancel();
  }

  // clear drops the recording, the words and the reply, and leaves the window
  // open on an empty buffer.
  function clear() {
    drop();
    pause();
    buffer = null;
    text = "";
  }

  // trash clears the window, ends the session and closes it. The thread
  // stays on the board: closing the window is leaving the conversation, and
  // the next one starts a new one.
  function trash() {
    clear();
    hush();
    thread = "";
    stamp = "";
    heard = 0;
    show(false);
  }

  // back steps back through what the window holds: a recording, then a reply,
  // then the window itself. Only the last step closes, so discarding a
  // recording leaves the window open for the next one.
  function back() {
    if (state() !== "idle") {
      clear();
      dress();
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
  // an empty buffer, so the next sentence is one press away. Everything but
  // a run goes to the session, with the clip for Claude to transcribe when
  // this device could not.
  function send() {
    if (state() === "idle") return;
    var words = sanitize(text);
    var to = target();
    var at = keep(words);
    pause();
    buffer = null;
    text = "";
    dress();
    if (words && to.kind === "run") {
      window.taskUI.paste(words);
      at.then(function (base) {
        say(base ? "Recording kept in artifacts, under voice/" : "The recording was not kept", base ? "ok" : "bad");
      });
      return;
    }
    at.then(function (base) {
      if (!base && !words) {
        say("The recording was not kept, so there was nothing to send", "bad");
        return;
      }
      if (base) say("Recording kept in artifacts, under voice/", "ok");
      asked(words, base);
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
