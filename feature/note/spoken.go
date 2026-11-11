package note

// Spoken is what the orb heard: Audio is the recording the message was
// spoken into, a path of the artifact tree, and Text the transcription of
// it.
//
// A prompt's body carries more than was said - where the clip is, and the
// brief the session opens on - so Text is what a thread shows and Body
// what the agent reads. A typed message leaves both empty.
type Spoken struct {
	Audio string
	Text  string
}

// Said reports whether anything was spoken.
func (s Spoken) Said() bool {
	return s.Audio != "" || s.Text != ""
}
