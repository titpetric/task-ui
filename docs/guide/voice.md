# Dictation

An orb sits in the bottom right of every page. Press it and a window opens over the page with the orb at full size, the words under it and three buttons. Press the orb in the middle to record, press it again to pause, and press it once more to add to the same recording. **Send** delivers what you said. The left button steps back: it throws a recording away, then dismisses an answer that came back, then closes the window, and its icon says which.

The orb shows wherever the browser can open a microphone: a secure page, so `https://` or `localhost`. It does not show over `http://` on a LAN address.

## What send does

Send writes the recording to `voice/<stamp>.wav` in the [artifact tree](artifacts.md), which needs `artifact/edit`; without it you still dictate and a toast says the recording went nowhere. Where it goes next depends on the browser and the page:

- **A run is attached and takes input.** The words go to the terminal as if typed, with no newline: punctuation is spelled out as `[dot]`, `[slash]` and the rest, so a dictated sentence cannot run a command.
- **Anything else.** The words go to Claude as a prompt about the page you are on, and the window shows what the request is doing, then Claude's answer, read out by a voice the device holds.

Chrome and Edge transcribe on the device, so the words never leave it: the orb reads `SpeechRecognition.available({processLocally: true})` and fetches the language the first time you press record. Safari and Firefox have no such recogniser, and neither do most phones. There the orb records and the instance does the rest.

## The Orb input workflow

A recording with no words of its own is answered by a workflow named **Orb input**, which the orb starts and then watches. It is three steps:

1. **transcribe-audio**, a workflow of its own: whisper.cpp over every recording in `voice/` with no subtitles beside it, writing `<stamp>.txt`, `.srt`, `.vtt` and an `.mp3`.
2. **A prompt in the root**: Claude reads every `<stamp>.txt` with no `<stamp>.reply.txt` beside it, answers the question in it in one or two sentences of plain words, and writes each answer to `voice/<stamp>.reply.txt` with `write_artifact`.
3. **A script**: piper over every `.reply.txt` with no `.reply.wav`, writing `voice/<stamp>.reply.wav`.

Each step finds its own work by what the step before it left in the tree, so the chain takes no argument and a recording that arrives while a run is going is picked up by the next one. The window polls for its own `<stamp>.reply.txt`, shows it, waits for the `<stamp>.reply.wav` and plays it. The replay button plays it again.

The window says which step the run is on: Transcribing, Thinking, Speaking, then Answered. A run that ends with nothing beside the recording leaves the reason in the window and in a toast, and the recording stays in the artifacts.

An instance without the workflow says so on send and keeps the recording. Build it with the three steps above, or ask Claude to on the Workflows page: the names of the jobs it runs are an instance's own, and the transcriber and the synthesiser are the instance's to install.

## The orb itself

Four orbs are drawn on one canvas, and two more join them on the ring an answer brings. Each is a radial greyscale, white in the middle and black at the rim, and the four are summed, so two that overlap sum past white. The sum is read through a palette of black to colour to white, and the pair then draws as one body with one outline. Every orb keeps the path it came along: the trail is drawn out of it, and so is the chain.

One formation per state of the window, and one becomes the next over a second and a half. The window names its state on every step of the run, so the three change flight path as the work moves on:

| State        | Colour  | What the orbs do                                                                                                                                             |
|--------------|---------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Idle         | violet  | The middle orb holds at the centre and the other three walk their own tilted rosettes around it, each swinging in and out on its own period.                 |
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

16 kHz mono, as a transcriber reads it: whisper resamples to that rate, so capturing more buys a later pass nothing. A recording is bounded at the size an artifact write takes, which is about eight minutes of it. Pause and resume write one file, so a pause costs nothing.
