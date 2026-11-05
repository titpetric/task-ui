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
| `voice-discard`  | Ends the session and writes nothing              |

The avatar is one SVG: three ellipses of fog, a core, and seven logarithmic
spirals in three whorls, every layer additive and frayed by a static
`feTurbulence` so no stroke reads as a surface. Only transforms and opacity
animate, and `--voice-level` sets the microphone's level on all of them.

Empty markup the script fills. `assets/voice/voice.js` owns the behaviour and
`assets/voice/voice.css` the look; neither can reach a run's socket, only
`taskUI.paste`.
