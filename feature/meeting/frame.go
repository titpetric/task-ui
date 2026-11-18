package meeting

import (
	"encoding/json"
)

// TextFrame marshals a control message as a pipe frame.
func TextFrame(msg Message) Frame {
	data, _ := json.Marshal(msg)
	return Frame{Text: data}
}

// AudioFrame names the chunk that follows, with the PCM as the frame's binary part.
func AudioFrame(userID, userName string, seq, startMS int64, pcm []byte) Frame {
	msg := Message{Type: "audio", UserID: userID, UserName: userName, Seq: seq, Rate: Rate, StartMS: startMS}
	data, _ := json.Marshal(msg)
	return Frame{Text: data, Data: pcm}
}
