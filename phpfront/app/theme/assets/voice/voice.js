// voice: dictation, transcribed on this instance or on the device.
//
// The orb in the corner opens the window: the avatar at full size, the words
// under it, and two buttons. The microphone opens once and the stream feeds
// two things, an analyser whose level drives the avatar and a worklet that
// collects the samples, so the recording can be kept beside the words it came
// out as.
//
// Where the stt integration is set up the server is the transcriber: a slice
// of the recording goes to /api/voice/transcribe every SLICE milliseconds and
// the words come back while the talking goes on. The device's own recogniser
// then stays shut, so one transcript is built and not two. Where it is not,
// the device transcribes as before.
//
// A press ends the recording and sends it: the orb and the send button are the
// same act, and pause is for a fault and a page sent to the background.
//
// A reply that holds a form block is a question. The card opens over the
// window with one field on it, its options as buttons, and Submit and Cancel;
// the answers go back as the next thing said. The server parses the yaml, so
// the dialect lives in form/ and this file holds no parser.
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
  var agent = document.getElementById("voice-agent");
  if (!stage || !replayBtn || !discardBtn) return;
  var discardIcon = discardBtn.querySelector("use");
  var acts = discardBtn.parentNode;
  var quiz = document.getElementById("voice-quiz");
  var quizCount = document.getElementById("voice-quiz-count");
  var quizAsk = document.getElementById("voice-quiz-ask");
  var quizPicks = document.getElementById("voice-quiz-picks");
  var quizHelp = document.getElementById("voice-quiz-help");
  var quizSubmit = document.getElementById("voice-quiz-submit");
  var quizCancel = document.getElementById("voice-quiz-cancel");
  if (!quiz || !quizPicks || !quizSubmit || !quizCancel) return;

  // RATE is the sample rate the recording is written at. 16 kHz mono is what a
  // transcriber reads: Whisper and wav2vec both resample to it, so capturing
  // more buys a later pass nothing. Raise it to 48000 to make the file worth
  // listening to, at three times the bytes.
  var RATE = 16000;

  // MAX bounds a recording at the artifact endpoint's own limit, 16 MiB, so a
  // dictation left running cannot fail its upload after the fact.
  var MAX = 16 << 20;

  // SLICE is how long a piece of the recording runs before it is posted to be
  // transcribed. Four seconds is short enough that the words keep up with the
  // talking and long enough that whisper has a phrase to work on: a slice is
  // transcribed on its own, so a cut lands mid-word often enough as it is.
  var SLICE = 4000;

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
    "",
    "To ask me something, end the reply with a ```form block: one field per question, component bool for",
    "a yes or no, choice with options for a pick, multi for several, text for words. The window draws it",
    "a question at a time as buttons and sends the answers back as my next message. Ask a questionnaire",
    "as one block of many short fields, with when: <key>=<value> on a field that follows from an answer,",
    "and submit.prompt as the sentence you want the answers read back in, {{key}} for each.",
  ].join("\n");

  // HEARD rides with a clip nothing could transcribe: the words it holds are
  // the question, so the session reads them first. KEPT rides with one the
  // words already cover: no second pass over the same audio.
  var HEARD = "The clip is at <clip>.wav. Transcribe it with the transcribe-audio workflow and answer what it asks.";

  var KEPT = "The recording is at <clip>.wav, already transcribed as the words above.";

  // STAGES name what the thread is doing while nothing has come back.
  var STAGES = ["Transcribing", "Thinking", "Speaking"];

  // TRIES is how many times a reply's audio is looked for, the first ten
  // half a second apart and the rest every three seconds: piper takes
  // seconds over a sentence, and minutes the first time, downloading a voice.
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
  var played = 0;       // how many answers of the thread have been played
  var source = null;    // the event stream the thread is followed over
  var gen = 0;          // a send or a discard retires the watch before it
  var since = 0;        // when the step the counter sits beside started
  var began = 0;        // when the act started, for the total beside the step
  var marks = [];       // every step of the act so far, as "Thinking 48s"
  var clock = 0;        // the interval writing the counter
  var doing = "";       // the step the counter sits beside
  var failed = false;   // the run ended and left nothing beside the recording
  var server = null;    // null until asked, then whether the server transcribes
  var sent = 0;         // how many frames of the buffer have been posted
  var posting = 0;      // the interval posting the slices
  var flight = false;   // a slice is out and the next tick waits for it
  var quizzing = null;  // { form, values, at } while a question is open

  // ------------------------------------------------------------------ gate

  // paint is the only writer of what is visible to press. The orb shows
  // wherever a microphone can be opened, whether or not this browser can
  // transcribe: without a recogniser the press still records, and the
  // transcript comes from the workflow that reads the artifact later. The run
  // window's own button shows on top of that when the run takes input,
  // because a modal run window puts the orb behind its backdrop.
  function paint() {
    orb.hidden = !canRecord;
    orb.title = ready || server ? "Dictate" : "Record";
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
      loud(smoothed);
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
    loud(0);
    if (!media) return;
    if (media.node) media.node.port.onmessage = null;
    media.stream.getTracks().forEach(function (t) { t.stop(); });
    if (media.ctx.state !== "closed") media.ctx.close();
    media = null;
  }

  // wav writes the frames from until to as a RIFF file, mono, downsampled to
  // RATE, and the whole buffer with neither. Each output sample is the mean of
  // the source window behind it; one sample picked out of that window would
  // alias a voice badly.
  function wav(from, to) {
    if (!buffer || !buffer.length) return null;
    var frames = buffer.frames.slice(from || 0, to === undefined ? buffer.frames.length : to);
    var length = 0;
    frames.forEach(function (f) { length += f.length; });
    if (!length) return null;
    var all = new Float32Array(length);
    var at = 0;
    frames.forEach(function (f) { all.set(f, at); at += f.length; });
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

  // ------------------------------------------------------------- transcribe

  // asking settles once whether the server transcribes. It does when the stt
  // integration is set up, and then it is the transcriber: the words arrive
  // while the talking goes on and the device's own recogniser stays shut, so
  // one transcript is built and not two.
  function asking() {
    if (server !== null) return Promise.resolve(server);
    return fetch("/api/voice/ready", { credentials: "same-origin", headers: { Accept: "application/json" } })
      .then(function (res) { return res.ok ? res.json() : null; })
      .then(function (data) {
        server = !!(data && data.ready);
        paint();
        return server;
      }, function () {
        server = false;
        return false;
      });
  }

  // slice posts the frames recorded since the last one and adds the words to
  // the transcript. A slice that fails stops the streaming for this recording:
  // the whole clip is still written at the end, and the session is asked to
  // transcribe it as it is for a device that cannot.
  function slice() {
    if (flight || !server || !buffer) return;
    var from = sent;
    var to = buffer.frames.length;
    if (to <= from) return;
    var clip = wav(from, to);
    if (!clip) return;
    flight = true;
    fetch("/api/voice/transcribe", {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "audio/wav", Accept: "application/json" },
      body: clip,
    }).then(function (res) {
      flight = false;
      if (!res.ok) throw new Error("the server did not transcribe the clip");
      return res.json();
    }).then(function (data) {
      sent = to;
      var words = (data && data.text ? data.text : "").trim();
      if (!words) return;
      text += (text ? " " : "") + words;
      heard(text, "");
    }).catch(function (err) {
      flight = false;
      streaming(false);
      server = false;
      console.info("voice: the server stopped transcribing, the clip goes to the session", err);
    });
  }

  // streaming starts and stops the posting of slices. Stopping posts what is
  // left, so the last words said are in the transcript the prompt is built
  // from, and answers once that slice is in.
  function streaming(on) {
    if (posting) window.clearInterval(posting);
    posting = 0;
    if (on) {
      posting = window.setInterval(slice, SLICE);
      return Promise.resolve();
    }
    if (!server || !buffer || buffer.frames.length <= sent) return Promise.resolve();
    // The last slice is waited for, up to WAIT: a prompt built before it
    // landed would be missing the end of the sentence. Past that the words
    // there are are the question.
    var WAIT = 30000;
    var until = Date.now() + WAIT;
    return new Promise(function (done) {
      var wait = function () {
        if (!flight) {
          slice();
          if (!flight) return done();
        }
        if (Date.now() > until) return done();
        window.setTimeout(wait, 100);
      };
      wait();
    });
  }

  // ------------------------------------------------------------------ ask

  // PICKS are the components the window answers with a row of options. A
  // bool is the yes or no every other question turns out to be, and it is
  // drawn as the same two options rather than a checkbox.
  var PICKS = { bool: true, choice: true, multi: true };

  var YESNO = [{ value: "yes", label: "Yes" }, { value: "no", label: "No" }];

  // forms reads the form blocks of a reply. The server parses the yaml, so
  // the dialect is declared in one place and the window holds no parser:
  // POST /api/forms answers with what each fence declares, or why it was
  // refused.
  function forms(body) {
    return fetch("/api/forms", {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "text/plain", Accept: "application/json" },
      body: body,
    }).then(function (res) { return res.ok ? res.json() : null; })
      .then(function (data) {
        return ((data && data.forms) || []).filter(function (one) { return one.form; });
      }, function () { return []; });
  }

  // visible is the fields the answers so far show, in declaration order. A
  // field names the answer that shows it as when: <key>=<value>, and one
  // whose own when-field is hidden stays hidden. This is form.Visible, which
  // the stepping repeats here because a question answered on the device
  // cannot wait for a round trip to know what to ask next.
  function visible(form, values) {
    var shown = {};
    return (form.fields || []).filter(function (field) {
      if (field.when) {
        var cut = field.when.indexOf("=");
        var key = field.when.slice(0, cut);
        if (!shown[key] || String(values[key] || "").trim() !== field.when.slice(cut + 1)) return false;
      }
      shown[field.key] = true;
      return true;
    });
  }

  // options are what a field offers: its own, or yes and no for a bool.
  function options(field) {
    return field.component === "bool" ? YESNO : (field.options || []);
  }

  // ask opens the question card over the window and draws the first field a
  // reply asked for. Nothing of the recording is touched: a question arrives
  // between recordings, and the orbs keep their formation behind it.
  function ask(form) {
    quizzing = { form: form, values: {}, at: 0 };
    (form.fields || []).forEach(function (field) {
      if (field.default) quizzing.values[field.key] = field.default;
    });
    draw();
  }

  // draw writes the field the card is on, its options and what the buttons
  // say: Submit answers the last question and sends, Next moves on.
  function draw() {
    if (!quizzing) return;
    var list = visible(quizzing.form, quizzing.values);
    if (quizzing.at >= list.length) return done();
    var field = list[quizzing.at];
    var picks = PICKS[field.component];
    quizCount.textContent = list.length > 1 ? (quizzing.at + 1) + " of " + list.length : "";
    quizAsk.textContent = field.label || field.key;
    quizHelp.textContent = field.help || "";
    quizHelp.hidden = !field.help;
    quizPicks.textContent = "";
    if (picks) options(field).forEach(function (option) { quizPicks.appendChild(pick(field, option)); });
    else quizPicks.appendChild(typed(field));
    quizSubmit.textContent = quizzing.at + 1 < list.length ? "Next" : "Submit";
    quiz.hidden = false;
    acts.hidden = true;
    stage.disabled = true;
    held(field);
  }

  // pick is one option as a button that selects it. A multi keeps every
  // option pressed, a choice keeps one.
  function pick(field, option) {
    var button = document.createElement("button");
    button.type = "button";
    button.className = "voice-pick";
    button.dataset.value = option.value;
    button.setAttribute("aria-pressed", "false");
    button.textContent = option.label || option.value;
    button.addEventListener("click", function () {
      if (field.component === "multi") {
        var set = String(quizzing.values[field.key] || "").split(",").filter(Boolean);
        var at = set.indexOf(option.value);
        if (at < 0) set.push(option.value);
        else set.splice(at, 1);
        quizzing.values[field.key] = set.join(",");
      } else {
        quizzing.values[field.key] = quizzing.values[field.key] === option.value ? "" : option.value;
      }
      held(field);
    });
    return button;
  }

  // typed is the input a field that offers no options takes, which a
  // requirement written out rather than chosen needs.
  function typed(field) {
    var input = document.createElement(field.component === "textarea" ? "textarea" : "input");
    input.className = "voice-typed";
    if (field.component !== "textarea") input.type = field.component === "number" ? "number" : "text";
    else input.rows = 3;
    input.value = quizzing.values[field.key] || "";
    input.placeholder = field.placeholder || "";
    input.setAttribute("aria-label", field.label || field.key);
    input.addEventListener("input", function () {
      quizzing.values[field.key] = input.value;
      held(field);
    });
    return input;
  }

  // held marks what is selected and refuses an empty answer to a required
  // field, so Submit is pressed only once there is something to send.
  function held(field) {
    var answer = String(quizzing.values[field.key] || "");
    var set = answer.split(",").filter(Boolean);
    quizPicks.querySelectorAll(".voice-pick").forEach(function (button) {
      var on = field.component === "multi"
        ? set.indexOf(button.dataset.value) >= 0
        : button.dataset.value === answer;
      button.setAttribute("aria-pressed", on ? "true" : "false");
      button.classList.toggle("on", on);
    });
    quizSubmit.disabled = !!field.required && !answer.trim();
  }

  // step takes the answer and moves on, or sends once the last field is
  // answered. A hidden field never asked keeps its default.
  function step() {
    if (!quizzing) return;
    quizzing.at++;
    draw();
  }

  // done sends the answers back into the thread and closes the card. The
  // form's prompt is what goes, with every placeholder filled in; a form
  // that declares none sends the answers as lines, which is what a
  // questionnaire collecting requirements wants read back to it.
  function done() {
    var form = quizzing.form;
    var values = quizzing.values;
    shut();
    var body = (form.submit && form.submit.prompt)
      ? form.submit.prompt.replace(/\{\{([^{}]*)\}\}/g, function (_, key) { return values[key.trim()] || ""; })
      : lines(form, values);
    if (!body.trim()) return;
    asked(body, "");
  }

  // lines is every answered field as "label: value", the reply to a form
  // that names no prompt of its own.
  function lines(form, values) {
    return (form.fields || []).filter(function (field) {
      return String(values[field.key] || "").trim() !== "";
    }).map(function (field) {
      return (field.label || field.key) + ": " + values[field.key];
    }).join("\n");
  }

  // drop the card and tell the thread the questions were left, so a session
  // waiting on an answer is not waiting on one that never comes.
  function quit() {
    if (!quizzing) return;
    shut();
    asked("I cancelled the questions.", "");
  }

  // shut closes the card and gives the window back its own buttons.
  function shut() {
    quizzing = null;
    quiz.hidden = true;
    quizPicks.textContent = "";
    acts.hidden = false;
    stage.disabled = false;
  }

  // ------------------------------------------------------------------ flow

  // ------------------------------------------------------------- avatar

  // The window draws one avatar at a time, the one of the user it calls.
  // An avatar is a file of the artifact tree, upload/users/<name>/avatar.js,
  // which sets window.voiceAvatar to an object with start, stop and the two
  // optional hooks level and tap. The canvas carries the window's state as
  // data-state, so an avatar reads what it is drawing without a call.
  //
  // docs/guide/voice.md holds the contract and what a user's own avatar may
  // do with the canvas.
  var AVATARS = "/api/artifacts/files/upload/users/";
  var dressed = {};
  var worn = null;

  // agentName is the user the window calls, "" for the reader themselves.
  function agentName() {
    return agent && agent.value ? agent.value : "";
  }

  // loud hands the microphone's level to the avatar, 0 to 1.
  function loud(level) {
    if (worn && worn.level) worn.level(level);
  }

  // bare stops the avatar on the canvas and clears it.
  function bare() {
    if (worn && worn.stop) worn.stop();
    worn = null;
    var out = avatar.getContext && avatar.getContext("2d");
    if (out) out.clearRect(0, 0, avatar.width, avatar.height);
  }

  // wear puts the avatar of a user on the canvas, the built-in orbs where
  // the user has none, and draws the state at once.
  function wear(name) {
    avatar.dataset.user = name;
    look(name, function (found) {
      // A second switch while this one loaded wins.
      if (avatar.dataset.user !== name || !overlay.open) return;
      bare();
      worn = found;
      if (worn && worn.start) worn.start(avatar);
      shape();
    });
  }

  // look answers the avatar of a user, loading the file once and keeping
  // whatever it registered. A file that is not there, or that registers
  // nothing, leaves the user with the built-in orbs.
  function look(name, then) {
    if (dressed[name]) return then(dressed[name]);
    if (!name) {
      dressed[name] = window.voiceOrbs;
      return then(dressed[name]);
    }
    var script = document.createElement("script");
    var done = function (found) {
      script.remove();
      window.voiceAvatar = null;
      dressed[name] = found || window.voiceOrbs;
      then(dressed[name]);
    };
    script.src = AVATARS + encodeURIComponent(name) + "/avatar.js?raw";
    script.onload = function () { done(window.voiceAvatar); };
    script.onerror = function () { done(null); };
    document.head.appendChild(script);
  }

  // target picks where send puts the words. A run that is attached and takes
  // input is the terminal; anything else is the header prompt, which is on
  // every page a user may ask Claude from.
  function target() {
    if (window.taskUI.attached()) return { kind: "run", say: "Sends to the run" };
    if (document.getElementById("header-prompt")) return { kind: "prompt", say: "Asks Claude, and reads the answer out" };
    return { kind: "keep", say: "Keeps the recording in the artifacts" };
  }

  // show opens and closes the window. showModal puts it in the top layer, so
  // it draws over a run window, which is a modal dialog of its own. The
  // avatar draws only while the window is open: nothing of it is seen after.
  function show(on) {
    if (on) {
      if (!overlay.open) overlay.showModal();
      wear(agentName());
      return;
    }
    said.textContent = "";
    where.textContent = "";
    bare();
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

  // mood is the one name for what the window is doing, which the avatar
  // draws: recording, transcribing, thinking, speaking, answered, failed,
  // paused or idle. Every path the window can be on ends here: a press, a
  // step of the run, an answer, a failure.
  function mood() {
    var now = state();
    if (now === "recording") return "recording";
    if (answer || reply) return "answered";
    if (failed) return "failed";
    if (pending) return FLIGHT[doing] || "thinking";
    if (now === "paused" || buffer) return "paused";
    return "idle";
  }

  // shape writes the mood where the avatar reads it: the canvas carries it
  // as data-state, which is what an avatar of a user's own reads, and the
  // built-in orbs take it as a call.
  function shape() {
    var name = mood();
    avatar.dataset.state = name;
    if (worn && worn.mood) worn.mood(name);
  }

  // dress writes the state onto the window and says what the orb will do.
  function dress(note) {
    var now = state();
    overlay.dataset.state = now;
    stage.setAttribute("aria-pressed", now === "recording" ? "true" : "false");
    stage.title = now === "recording" ? "Stop and send" : (now === "paused" ? "Send" : "Record");
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
    starting = true;
    dress(text || "Recording");
    asking().then(function () {
      var SR = window.SpeechRecognition;
      return !server && ready && install && SR.install
        ? SR.install({ langs: [lang], processLocally: true }).then(function () { install = false; })
        : undefined;
    }).then(listening).then(function () {
      starting = false;
      if (server) streaming(true);
      dress(text || undefined);
    }).catch(function (err) {
      starting = false;
      refused(err);
    });
  }

  // pause closes the microphone and keeps everything heard and captured. It
  // sends nothing: a fault and a page sent to the background come through
  // here, and the orb's own press ends in stop.
  function pause() {
    starting = false;
    streaming(false);
    if (rec) {
      rec.onresult = null;
      rec.onerror = null;
      try { rec.stop(); } catch (e) { /* already stopped */ }
      rec = null;
    }
    release();
    dress();
  }

  // stop ends the recording the way the reader means it to end: the
  // microphone closes, the slice still being talked into is posted and waited
  // for, and the words go wherever send puts them. Pressing the orb and
  // pressing send during a recording are the same act.
  function stop() {
    if (state() === "idle") return;
    starting = false;
    if (rec) {
      rec.onresult = null;
      rec.onerror = null;
      try { rec.stop(); } catch (e) { /* already stopped */ }
      rec = null;
    }
    release();
    dress();
    // The counter runs while the last slice is out: on a machine where the
    // model costs more than the clip, this is the wait, and a line that does
    // not move reads as stuck.
    if (server) ticking("Transcribing");
    streaming(false).then(function () {
      stopped();
      send();
    });
  }

  // listening opens the recogniser and the microphone, and returns once both
  // are running or the microphone alone where there is no recogniser.
  //
  // The recogniser opens its own microphone and is what dictation is; the
  // analyser and the worklet are a second consumer of the same device, so
  // they come after and are allowed to fail.
  function listening() {
    var SR = window.SpeechRecognition;
    if (server || !ready) {
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

  // aged is the whole seconds between a stamp and now.
  function aged(at) {
    return Math.round((Date.now() - at) / 1000);
  }

  // mark closes the step the counter sits beside, keeping how long that step
  // took, and starts the counter on the next one.
  function mark(next) {
    if (doing && since) marks.push(doing + " " + span(aged(since)));
    doing = next;
    since = Date.now();
  }

  // tell writes the step, how long it has been going, and the total since the
  // act began. A transcription on a busy machine takes a minute, and a line
  // that does not move reads as stuck; the step's own seconds say nothing
  // about the wait so far, which is the number a reader is counting.
  function tell(now) {
    if (now !== undefined) {
      if (now !== doing) mark(now);
      shape();
    }
    var line = doing + " · " + span(aged(since));
    if (marks.length) line += " · " + span(aged(began));
    where.textContent = line;
  }

  // ticking starts the counter beside a step, and the first step of an act
  // starts the total: a reader waits out every step of it, where the one on
  // screen is a fraction of the wait.
  function ticking(now) {
    if (!began) {
      began = Date.now();
      marks = [];
    }
    if (clock) window.clearInterval(clock);
    clock = window.setInterval(function () { tell(); }, 1000);
    tell(now);
  }

  // ended stops the counter and leaves the total on the line, with every step
  // of the act in the console: a reply that took two minutes is a question
  // about which step took them.
  function ended(word) {
    if (!began) {
      stopped();
      where.textContent = word;
      return;
    }
    mark("");
    var total = span(aged(began));
    stopped();
    where.textContent = word + " · " + total;
    console.info("voice: " + word.toLowerCase() + " in " + total + " · " + marks.join(", "));
    began = 0;
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
    else if (clip) body = words + "\n\n" + KEPT.replace("<clip>", clip);
    if (!thread) {
      stamp = clip || "voice/" + new Date().toISOString().replace(/[:.]/g, "-").replace("Z", "");
      body = body + "\n\n" + BRIEF.replace("<base>", stamp);
    }
    window.taskUI.ask(body, function (update) {
      if (mine !== gen) return;
      if (update.thread) {
        thread = update.thread;
        follow();
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
    }, thread, agentName());
  }

  // follow watches the thread for as long as the window is open, so a reply
  // Claude writes after the work it started finishes reaches the window on
  // its own. It is one stream for the session.
  function follow() {
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
        var answers = (mineThread.replies || []).filter(function (reply) {
          return reply.user_name !== mineThread.user_name;
        });
        if (answers.length <= played) return;
        played = answers.length;
        landed(answers[answers.length - 1].body);
      })
      .catch(function () {});
  }

  // landed shows a reply and waits for the audio of it. The count in the name
  // is how many answers the thread holds, which is what BRIEF tells Claude to
  // count; the window writes no file and only reads.
  //
  // A reply that holds a form block asks something: the card opens over the
  // window with the first question on it, and the answers go back as the
  // next thing said. The words are still read out, so the question is heard
  // as well as read.
  function landed(body) {
    var mine = gen;
    var words = plain(body);
    if (!words || words === answer) return;
    pending = false;
    answer = words;
    said.textContent = answer.slice(0, 400);
    if (played < 1) played = 1;
    tell(STAGES[2]);
    audio(mine, stamp + "." + played, 0);
    if (body.indexOf("```form") < 0) return;
    forms(body).then(function (found) {
      if (mine !== gen || !found.length || quizzing) return;
      ask(found[0].form);
    });
  }

  // audio waits for the wav of a reply and plays it, and reads the words out
  // on this device when none arrives. It asks often at first, for a reply
  // already spoken, then every few seconds while piper works.
  function audio(mine, base, tries) {
    if (mine !== gen || failed) return;
    if (tries > TRIES) {
      ended("Answered");
      speak(answer);
      replayBtn.hidden = !answer;
      return;
    }
    read(base + ".reply.wav").then(function (data) {
      if (mine !== gen || failed) return;
      if (!data) {
        window.setTimeout(function () { audio(mine, base, tries + 1); }, tries < 10 ? 500 : 3000);
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
  // A failed run posts its reason as a reply, so the window has already taken
  // it for an answer and gone looking for the audio of it. Letting go of the
  // answer here is what stops that hunt and keeps the reason off the replay.
  function fell(err) {
    var why = (err && err.message) || String(err);
    pending = false;
    failed = true;
    answer = "";
    reply = "";
    replayBtn.hidden = true;
    said.textContent = "";
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
    began = 0;
    marks = [];
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
    sent = 0;
  }

  // trash clears the window, ends the session and closes it. The thread
  // stays on the board: closing the window is leaving the conversation, and
  // the next one starts a new one.
  function trash() {
    clear();
    shut();
    hush();
    thread = "";
    stamp = "";
    played = 0;
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
    sent = 0;
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
  // The pick of user is the pick of avatar: the window wears whoever it
  // calls, and the words go to that user's Claude.
  if (agent) agent.addEventListener("change", function () { wear(agentName()); });
  stage.addEventListener("click", function () {
    if (worn && worn.tap) worn.tap();
    if (media || starting) stop();
    else record();
  });
  sendBtn.addEventListener("click", function () {
    if (media || starting) stop();
    else send();
  });
  quizSubmit.addEventListener("click", step);
  quizCancel.addEventListener("click", quit);
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
