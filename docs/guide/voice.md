# Dictation

An orb sits in the bottom right of every page. Press it and a window opens over the page with the orb at full size, the words under it and three buttons. Press the orb in the middle to record and press it again to end the recording and send it; **Send** during a recording does the same thing. The left button steps back: it throws a recording away, then dismisses an answer that came back, then closes the window, and its icon says which. [Wake](#wake) starts the recording on a name you say.

The orb shows wherever the browser can open a microphone: a secure page, so `https://` or `localhost`. It does not show over `http://` on a LAN address.

## What send does

Send writes the recording to `voice/<stamp>.wav` in the [artifact tree](artifacts.md), which needs `artifact/edit`; without it you still dictate and a toast says the recording went nowhere. Where it goes next depends on the browser and the page:

- **A run is attached and takes input.** The words go to the terminal as if typed, with no newline: punctuation is spelled out as `[dot]`, `[slash]` and the rest, so a dictated sentence cannot run a command.
- **Anything else.** The words go to Claude, in the session below. The window shows what the request is doing, then what came back, played as audio.

## Wake

**Wake**, the switch above the orb, starts the recording when you say a name. The window keeps the microphone open while it is on and nothing else is going on, reads what it hears a piece at a time and throws every piece away. A piece that holds the name of a user the window may call starts the recording there: that user is picked, their avatar comes up, and whatever followed the name is the first thing said. `miguel, what did the deploy do` is one sentence, and the question reaches `@miguel`.

The names are the ones the picker lists, your own among them, and `hey` or `ok` in front of one is a word like any other. The switch shows where a name can be heard at all, which takes a transcriber: this instance's, or the one in the browser. It holds between windows and pages, so hands free is one press ever.

A recording a name started ends after two and a half seconds of quiet and sends itself, which is the press that is not made hands free. One a press started still ends on a press, however long the quiet runs. Pressing the orb while the window is listening starts the recording there and then, and drops everything heard before the press.

The listening is a microphone and nothing else: a piece with no name in it is dropped where it is read, no piece is kept, written or sent, and the whole of it stops when the window closes, when the page goes to the background, or while a reply is being read out. A quiet room costs the transcriber nothing, since a piece nobody spoke into is dropped before it is posted.

## Who transcribes

With the **Speech to text** integration set up, this instance transcribes and the words arrive while you are still talking: every four seconds the window posts the piece of the recording since the last one to `POST /api/voice/transcribe`, which runs whisper.cpp on the server and answers with the words. The transcript the prompt is built from is what those pieces said, and the last piece is waited for before the prompt goes out. Transcribing takes `prompt/edit`, the same as asking Claude.

The integration names two paths: the `whisper-cli` binary and the ggml model, both built and downloaded by the `setup` job of the `stt` folder. A piece is transcribed on its own, so a cut lands mid-word often enough.

**The model has to run faster than the talking.** A piece covers four seconds, so a model that takes longer than four seconds over four seconds of speech falls behind, and the words land after the sentence is finished. One measurement, over `samples/jfk.wav` on the four cores of task.cdn.si with `threads` at 4:

| Model      | Clip | Transcribed in | Ratio |
|------------|------|----------------|-------|
| `small.en` | 9.4s | 67.7s          | 7.2x  |

`small.en` is what the `setup` job downloads, and 7.2x is too slow for this: run the job with `MODEL=tiny.en` for one that keeps up, and time a clip of your own before trusting it. A slow model breaks nothing, as a piece is posted only once the one before it has come back and the window counts the seconds it waits, but the words stop being live.

Without the integration, Chrome and Edge transcribe on the device and the words never leave it: the orb reads `SpeechRecognition.available({processLocally: true})` and fetches the language the first time you press record. Safari and Firefox have no such recogniser, and neither do most phones. There the orb records and the session does the rest.

Only one of the two runs. Where the server transcribes, the device's recogniser stays shut, so one transcript is built and not two.

## The session

Everything the orb sends, bar a run's terminal, goes to one prompt thread. The first clip opens it; every later one replies in it. Claude reads the whole conversation each time and holds [the task tools](mcp.md), so it can run jobs, start workflows and read artifacts to answer, or to do the work itself.

The thread opens with an instruction the orb writes: reply in two or three spoken sentences of plain words; for work that takes a while, leave a sentence saying what you started, do it, then answer with what came of it; and after each reply write the same sentences to `<clip>.N.reply.txt` and run the `speak` job of `titpetric/stt`, so the window has the audio.

Nothing is fed forward while the microphone is open. The transcript goes to `<clip>.txt` as it is built, every ten seconds and once more at the end, and the prompt is written from the same words when the recording is sent. A reply that lands mid sentence waits there: its audio does not play into the recording and its question does not take the window. Sending answers it, and the thread keeps the reply the window moved past; discarding the recording plays it.

The window follows the thread over `/api/sse/notes` for as long as it is open. Nothing is polled: a reply pushes itself, and that holds for the second reply as much as the first, so work that finishes ten minutes later arrives and plays on its own. The replay button plays the last one again.

A clip nothing transcribed is sent with no words, and the thread is asked to run the `transcribe-audio` workflow over it first. A transcribed clip goes with its words and the path of the recording: no second pass over the same audio.

Closing the window ends the session. The thread stays on the board under the page it belongs to, and the next time the orb opens it starts a new one.

## Questions back

A reply that ends with a [form block](forms.md) is a question. The card opens over the window with one field on it, the options as buttons you press to select, and **Submit** and **Cancel**. What you answered goes back into the thread as the next thing said, so Claude reads it before asking anything else.

The thread's opening instruction tells Claude how: one field per question, `bool` for a yes or no, `choice` with `options` for a pick, `multi` for several, `text` for words. This is the same dialect the form block of a note's reply renders on a page, so one block works in both places.

A questionnaire is one block of many short fields. The card draws them one at a time and counts them, `4 of 23`, and sends once, so feature mapping costs one reply and not twenty. `when: <key>=<value>` on a field draws it only where the field it names holds that value, so a hundred declared questions are the dozen that follow from what you said. `submit.prompt` is the sentence the values are read back in, `{{key}}` for each; a form that names none sends `label: value` a line at a time.

```form
title: The list page
fields:
  - key: cache
    label: Should the list page cache its results?
    component: bool
  - key: ttl
    label: How long for?
    component: choice
    when: cache=yes
    options:
      - value: 1m
      - value: 5m
      - value: 1h
submit:
  prompt: "cache: {{cache}}, ttl: {{ttl}}"
```

Cancel sends one line saying the questions go unanswered, so a thread waiting on an answer is not waiting on one that never comes. The window plays the reply as it always does, so a question is heard as well as read, and the microphone is shut while the card is open.

The instance has to hold a transcriber and a synthesiser for the audio half of this: `transcribe-audio` over whisper.cpp, and a `speak` job over piper. Without them the conversation still works in text, read out by whatever voice the device holds.

## The orb itself

Four orbs are drawn on one canvas, and two more join them on the ring an answer brings. Each is a radial greyscale, white in the middle and black at the rim, and the four are summed, so two that overlap sum past white. The sum is read through a palette of black to colour to white, and the pair then draws as one body with one outline. Every orb keeps the path it came along: the trail is drawn out of it, and so is the chain.

One formation per state of the window, and one becomes the next over a second and a half. The window names its state on every step of the run, so the three change flight path as the work moves on:

| State        | Colour  | What the orbs do                                                                                                                                             |
|--------------|---------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Idle         | violet  | The middle orb holds at the centre and the other three walk their own tilted rosettes around it, each swinging in and out on its own period.                 |
| Listening    | violet  | The idle formation: a window waiting on a name is a window doing nothing, and an avatar of a user's own reads `listening` off the canvas and may say so.     |
| Recording    | red     | The middle orb leads, sweeping twice across the canvas for once down, and the other three follow along the path it left, each a gap behind the one in front. |
| Paused       | orange  | The chain curls onto one upright circle around the middle orb and turns slowly on it.                                                                        |
| Transcribing | teal    | The three stack into a column and sweep across the canvas, through the middle orb.                                                                           |
| Thinking     | blue    | The three turn fast and close around the middle orb, each on its own tilted plane.                                                                           |
| Speaking     | magenta | Each of the three runs out from the middle orb along its own spoke, swelling from nothing and back to nothing, one after the other.                          |
| Answered     | green   | The middle orb returns to the centre and five turn on one flat circle that runs over it, drawn smaller because every z on that circle is negative.           |
| Failed       | grey    | The three stand off at the rim and turn there, and the middle orb is left alone.                                                                             |

Two of the three orbits hold off the middle orb. Each is pulled in and tilted until its closest approach leaves a thin neck of field between the two, so a pass draws a drop that forms and breaks; the distance is solved from the orb radii when the page loads. The third dives: its path runs from the centre out past the middle orb and back, so the pair merges at the inner end of every swing.

The middle orb is the one that answers you. Its radius swings 12% over four and a half seconds while nothing is happening, rises with how loud you are while recording, and swells a quarter and settles when you press it.

Every orb puts the same field into the middle of itself. The trail is normalised: an orb's samples land on the orb wherever it is slow, so a long trail would otherwise draw a third brighter than a short one.

Reduced motion draws one frame and runs no loop.

## What the recording is

16 kHz mono, as a transcriber reads it: whisper resamples to that rate, so capturing more buys a later pass nothing. A recording is bounded at the size an artifact write takes, which is about eight minutes of it. What the server transcribes as you talk and what is written at the end are the same samples: the pieces are slices of the one buffer, and the artifact holds every frame.
