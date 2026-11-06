# voice

Input scope: none.

The orb and the window behind it, written once in the shell because every page
dictates.

| Element          | What it is                                              |
|------------------|---------------------------------------------------------|
| `voice-orb`      | The button in the corner, hidden until the browser proves it transcribes on the device |
| `voice-overlay`  | A modal dialog: the avatar, the words heard, and the two acts |
| `voice-said`     | The transcript as it arrives, the final text with the interim run after it |
| `voice-where`    | What send will do with the words, since the window opens over a run and over a page without one |
| `voice-send`     | Puts the words in the run or the header prompt, and keeps the recording |
| `voice-discard`  | Steps back: the recording, then the reply, then the window |
| `voice-avatar`   | The canvas `assets/voice/orbs.js` draws the orbs into |

The avatar is a canvas: four orbs added together as greyscale and read through
a palette, so a pair that overlaps reads as one body. The middle one is the
microphone's meter. See `docs/guide/voice.md`.

Empty markup the script fills. `assets/voice/voice.js` owns the behaviour and
`assets/voice/voice.css` the look; neither can reach a run's socket, only
`taskUI.paste`.
