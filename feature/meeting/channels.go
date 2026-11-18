// Package meeting keeps the rooms of the orb: who is in one, what each of
// them said, and the audio each microphone contributed.
package meeting

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/titpetric/task-ui/app/model"
)

// Rate is the sample rate every channel is written at, mono 16-bit PCM.
const Rate = 16000

// BytesPerMS counts the bytes a millisecond of a channel takes: 16 samples of two.
const BytesPerMS = Rate * 2 / 1000

// Channels keeps one PCM file per participant under <root>/<meeting>/<user>.pcm.
type Channels struct {
	Root string
}

// path resolves a participant's channel file, refusing an ID that is not one.
func (c *Channels) path(meetingID, userID string) (string, error) {
	if !model.ValidID(meetingID) || !model.ValidID(userID) {
		return "", errors.New("a channel needs a meeting and a user")
	}
	return filepath.Join(c.Root, meetingID, userID+".pcm"), nil
}

// Append writes PCM to a participant's channel at the end and answers the offset it landed on.
func (c *Channels) Append(meetingID, userID string, pcm []byte) (int64, error) {
	path, err := c.path(meetingID, userID)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	if _, err := file.Write(pcm); err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// Size returns how many bytes a participant's channel holds.
func (c *Channels) Size(meetingID, userID string) (int64, error) {
	path, err := c.path(meetingID, userID)
	if err != nil {
		return 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// WAV returns a participant's channel as a WAV, header included, for Range serving.
func (c *Channels) WAV(meetingID, userID string) (io.ReadSeeker, int64, error) {
	path, err := c.path(meetingID, userID)
	if err != nil {
		return nil, 0, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, err
	}
	header := wavHeader(info.Size())
	return &wavReader{header: header, file: file, size: int64(len(header)) + info.Size()}, int64(len(header)) + info.Size(), nil
}

// Remove deletes every channel of a meeting.
func (c *Channels) Remove(meetingID string) error {
	if !model.ValidID(meetingID) {
		return errors.New("a meeting id is required")
	}
	return os.RemoveAll(filepath.Join(c.Root, meetingID))
}

// wavHeader is a RIFF header for dataSize bytes of 16 kHz mono 16-bit PCM.
func wavHeader(dataSize int64) []byte {
	header := make([]byte, 44)
	copy(header[0:], "RIFF")
	binary.LittleEndian.PutUint32(header[4:], uint32(36+dataSize))
	copy(header[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(header[16:], 16)
	binary.LittleEndian.PutUint16(header[20:], 1)
	binary.LittleEndian.PutUint16(header[22:], 1)
	binary.LittleEndian.PutUint32(header[24:], Rate)
	binary.LittleEndian.PutUint32(header[28:], Rate*2)
	binary.LittleEndian.PutUint16(header[32:], 2)
	binary.LittleEndian.PutUint16(header[34:], 16)
	copy(header[36:], "data")
	binary.LittleEndian.PutUint32(header[40:], uint32(dataSize))
	return header
}

// wavReader reads the WAV header this instance writes, then the PCM file it names.
type wavReader struct {
	header []byte
	file   *os.File
	size   int64
	pos    int64
}

// Read reads the header before the file, so http.ServeContent sees one stream.
func (w *wavReader) Read(p []byte) (int, error) {
	if w.pos < int64(len(w.header)) {
		n := copy(p, w.header[w.pos:])
		w.pos += int64(n)
		return n, nil
	}
	n, err := w.file.ReadAt(p, w.pos-int64(len(w.header)))
	w.pos += int64(n)
	return n, err
}

// Seek moves within the header and the file as one stream.
func (w *wavReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += w.pos
	case io.SeekEnd:
		offset += w.size
	}
	if offset < 0 {
		return 0, errors.New("seek before the start of the channel")
	}
	w.pos = offset
	return offset, nil
}

// Close closes the channel file.
func (w *wavReader) Close() error {
	return w.file.Close()
}
