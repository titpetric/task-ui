package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestEncodeList(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "[]", model.EncodeList(nil))
	assert.Equal(t, `["a","b"]`, model.EncodeList([]string{"a", "b"}))
}

func TestDecodeList(t *testing.T) {
	t.Parallel()
	assert.Len(t, model.DecodeList(""), 0)
	assert.Len(t, model.DecodeList("not json"), 0)
	assert.Equal(t, []string{"a", "b"}, model.DecodeList(`["a","b"]`))
}

func TestEncodeMap(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "{}", model.EncodeMap(nil))
	assert.Equal(t, `{"a":["x"],"b":[]}`, model.EncodeMap(map[string][]string{"b": {}, "a": {"x"}}))
}

func TestDecodeMap(t *testing.T) {
	t.Parallel()
	assert.Len(t, model.DecodeMap(""), 0)
	assert.Equal(t, map[string][]string{"deploy": {"admin"}}, model.DecodeMap(`{"deploy":["admin"]}`))
}

func TestEncodeStrings(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "{}", model.EncodeStrings(nil))
	assert.Equal(t, `{"make":"gmake"}`, model.EncodeStrings(map[string]string{"make": "gmake"}))
}

func TestDecodeStrings(t *testing.T) {
	t.Parallel()
	assert.Len(t, model.DecodeStrings(""), 0)
	assert.Equal(t, map[string]string{"make": "gmake"}, model.DecodeStrings(`{"make":"gmake"}`))
}

func TestSortedKeys(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"a", "b"}, model.SortedKeys(map[string]int{"b": 1, "a": 2}))
	assert.Len(t, model.SortedKeys(map[string]bool{}), 0)
}
