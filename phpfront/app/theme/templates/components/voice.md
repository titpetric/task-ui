# voice

Input scope: none.

The orb and the window behind it, written once in the shell because every page
dictates.

| Element          | What it is                                              |
|------------------|---------------------------------------------------------|
| `voice-orb`      | The button in the corner, hidden until the browser proves it records |
| `voice-overlay`  | A modal dialog: the avatar, the words heard, the status and the two actions |
| `voice-stage`    | The avatar as a button: record, begin, stop or send, as `orb.json` says for the state |
| `voice-said`     | The transcript, the reply or the saved message |
| `voice-where`    | The state's status, such as `Transcript saved · thread created` |
| `voice-discard`  | Back: closes the window, or trashes the recording while one is held |
| `voice-avatar`   | The canvas `assets/voice/orbs.js` draws the orbs into |

The avatar is a canvas: four orbs added together as greyscale and read through
a palette, so a pair that overlaps reads as one body. The middle one is the
microphone's meter. See `docs/guide/voice.md`.

`assets/voice/orb.json` names the state, its two actions, and the title, icon
and class each action wears. `assets/voice/voice.js` applies it and owns the
behaviour; `assets/voice/voice.css` owns the look. Neither can reach a run's
socket, only `taskUI.paste`.
