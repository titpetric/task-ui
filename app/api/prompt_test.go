package api

import (
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/folder"
)

// placeFolders is a catalogue of one folder, infra.
type placeFolders struct{}

func (placeFolders) Find(id string) *folder.Folder {
	if id == "infra" {
		return &folder.Folder{ID: "infra"}
	}
	return nil
}

func (placeFolders) List() []*folder.Folder {
	return []*folder.Folder{{ID: "infra"}}
}

// A place is seen by a user who sees every folder, or the folder the
// place lies in; the root is seen by the user who sees every folder.
func TestSeesPlace(t *testing.T) {
	t.Parallel()
	every := &model.User{}
	listed := &model.User{Folders: model.EncodeList([]string{"infra"})}
	other := &model.User{Folders: model.EncodeList([]string{"tools"})}

	assert.True(t, seesPlace(placeFolders{}, every, model.Default), "the root is a user's with every folder")
	assert.False(t, seesPlace(placeFolders{}, listed, model.Default), "a listed user sees only the list")
	assert.False(t, seesPlace(placeFolders{}, nil, model.Default))
	assert.True(t, seesPlace(placeFolders{}, listed, "infra"))
	assert.True(t, seesPlace(placeFolders{}, listed, "infra/src"))
	assert.False(t, seesPlace(placeFolders{}, other, "infra"))
	assert.False(t, seesPlace(placeFolders{}, listed, "missing"))
}
