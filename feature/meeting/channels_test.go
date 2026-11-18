package meeting

import (
	"encoding/binary"
	"io"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A channel appends PCM, answers its size, and reads back as a WAV with the same samples.
func TestChannels(t *testing.T) {
	t.Parallel()
	channels := &Channels{Root: t.TempDir()}
	id, user := model.NewID(), model.NewID()

	offset, err := channels.Append(id, user, []byte{1, 2})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), offset)
	offset, err = channels.Append(id, user, []byte{3, 4})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), offset)

	size, err := channels.Size(id, user)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), size)

	reader, total, err := channels.WAV(id, user)
	assert.NoError(t, err)
	assert.Equal(t, int64(48), total)
	data, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Len(t, data, 48)
	assert.Equal(t, "RIFF", string(data[0:4]))
	assert.Equal(t, uint32(4), binary.LittleEndian.Uint32(data[40:44]))
	assert.Equal(t, []byte{1, 2, 3, 4}, data[44:])

	// A range reads from the middle of the samples, past the header.
	_, err = reader.Seek(46, io.SeekStart)
	assert.NoError(t, err)
	rest := make([]byte, 2)
	_, err = io.ReadFull(reader, rest)
	assert.NoError(t, err)
	assert.Equal(t, []byte{3, 4}, rest)

	assert.NoError(t, channels.Remove(id))
	_, err = channels.Size(id, user)
	assert.True(t, err != nil)
}

// A channel refuses an ID that is not a ULID.
func TestChannels_Refuses(t *testing.T) {
	t.Parallel()
	channels := &Channels{Root: t.TempDir()}
	_, err := channels.Append("../escape", model.NewID(), []byte{1})
	assert.Error(t, err)
	_, err = channels.Size(model.NewID(), "x")
	assert.Error(t, err)
	_, _, err = channels.WAV("../../etc", "passwd")
	assert.Error(t, err)
	assert.Error(t, channels.Remove("."))
}
