package main

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

// The fake hub answers the client as a hub would: three hosts, one
// down, their containers and a month of records with gaps.
func TestHub(t *testing.T) {
	t.Parallel()
	now := time.Now()
	server := httptest.NewServer(newHub(func() time.Time { return now }))
	t.Cleanup(server.Close)
	c := &beszel.Client{URL: server.URL, Token: "t", TTL: time.Minute}
	ctx := context.Background()

	systems, err := c.Systems(ctx)
	assert.NoError(t, err)
	assert.Len(t, systems, 3)
	assert.NotNil(t, systems[0].Stats)
	assert.Len(t, systems[0].Stats.ExtraFs, 1)
	assert.Nil(t, systems[2].Stats, "down")

	containers, err := c.Containers(ctx)
	assert.NoError(t, err)
	assert.Len(t, containers["pve0000000000a"], 6)

	history, err := c.History(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 90, len(history["pve0000000000a"]))
	assert.True(t, len(history["dev10000000000"]) < 90, "a gap")
	assert.True(t, len(history["docker1000000a"]) < 80, "down for days")

	month, err := c.ContainerHistory(ctx, "dev10000000000")
	assert.NoError(t, err)
	assert.Len(t, month[0].Containers, 4)
}
