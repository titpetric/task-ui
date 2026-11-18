# Meetings

**Meetings**, `/orb`, is a room that keeps every speaker apart. Open one with a title and **Start meeting**, press **Copy link** in the room, and send the link to whoever should join. Any signed-in user with `meeting/edit` may join. The page opens on **Join meeting** and joins nothing until it is pressed. A read of the room adds no participant and starts no microphone.

The room shows the participants, the microphone, the time left and the transcript. **Join meeting** connects the browser and shows the microphone; press the microphone to speak and press it again to mute. **Leave** stops the microphone, plays nothing more, and keeps the page and its transcript. Each browser captures 16 kHz mono PCM and sends chunks a quarter second long over a websocket to task-ui, which appends them to that speaker's channel and relays them to every other participant. No peer-to-peer connection is made and no signalling server is involved. Headphones keep the room's own voices out of your microphone.

Every speaker's speech is one file, `data/meetings/<meeting>/<user>.pcm`, written in the order it arrived and read as a WAV with Range support. A transcript line plays the speech it covers from that speaker's file alone: the play button beside the line takes its start and end from the line's channel offset. Silence under a small level is not sent, so a channel holds speech and not the room's clock. A channel grows with the time its speaker talks; deleting the room deletes the audio.

| One speaker's channel | Size         |
|-----------------------|--------------|
| One hour of speech    | about 115 MB |

The transcript is written by the browser's speech recogniser, Chrome's and Edge's. Interim words appear at once under the speaker's name, and a final result becomes a line with a speaker, a local time, a start and an end, saved in `meeting_line` and sent to every participant. A browser without a recogniser still sends and plays audio and says so; its speech reaches no transcript, and the summary has only what the other browsers transcribed.

## Ending a room

**Finish** ends the room at once, for its creator or an admin. A room also ends at its hour, and a minute after the last participant leaves. A reload does not log a part: a closed socket leaves presence to the sweeper, which marks a silent browser gone once its last sighting is stale and starts the empty minute only then. Everyone still connected is told, and the page reloads into the ended room.

The transcript then goes to Claude, if the Claude Code or OpenRouter integration is set up for the room's creator, in a background run that does not block the page. That reply is markdown under `## Topics discussed`, `## Action items`, `## Decisions` and `## Open questions`, shown above the transcript, with the whole log under **Transcript**. Without an integration the panel says Claude is not set up; the **Ask Claude again** button runs the summary once it is.

## Permissions and data

`meeting/view` lists the rooms, opens one, reads its log and plays a channel. `meeting/edit` joins, speaks, leaves and starts a summary. `meeting/delete` deletes an ended room from its page or its row in Meetings, and only its creator or an admin may. Runners and viewers start with all three, so a room is open to every signed-in user; an admin narrows that under [Settings, Permissions](settings.md#permissions). A room that has ended keeps its log and its audio until someone deletes it.

The pages are `/orb` and `/orb/<id>`, and `/api/meetings` answers them with the routes in [API](../api.md). The module is a plugin as well, at `/m/taskui/orb`.
