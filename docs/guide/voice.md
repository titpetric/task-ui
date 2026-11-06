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

Four orbs circle one centre on a canvas. Each is drawn as a plain radial greyscale, white in the middle and black at the rim, and the four are added together, so where two overlap the sum runs past white. That sum is read through a palette of black to colour to white, and a pair that overlaps then reads as one body with one outline. Each orb is drawn again at the points it came through, dimmer each time, so it drags its own path behind it.

The orb at the centre is the one that answers you. It breathes while nothing is happening, swells with how loud you are while recording, and swells and settles when you press it. Its colour is what the window is doing: purple waiting, red recording, green when an answer has landed.

Reduced motion draws one frame of it and runs no loop.

## What the recording is

16 kHz mono, as a transcriber reads it: whisper resamples to that rate, so capturing more buys a later pass nothing. A recording is bounded at the size an artifact write takes, which is about eight minutes of it. Pause and resume write one file, so a pause costs nothing.
