package view

import (
	"strconv"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/hostinfo"
)

// HostDetail is a host's expanded row on the Inventory page: its
// containers, Beszel's measures and docker's sizes put together, its
// durable filesystems, and its month of charts.
type HostDetail struct {
	System beszel.System
	// Login is who task-ui signs in to the host as, "" for none.
	Login string
	// Containers are the host's containers, the busiest first; Measured
	// says Beszel reported the host's containers, Sized that docker
	// answered over ssh.
	Containers []hostinfo.Container
	Measured   bool
	Sized      bool
	// Filesystems are the durable ones; FromDF says they come from df
	// over ssh, else from Beszel.
	Filesystems []hostinfo.Filesystem
	FromDF      bool
	// Read is when the host last answered over ssh; Unsized says why
	// docker's sizes are missing: no login, no answer yet, the error of
	// the last try, or no docker for the login.
	Read    time.Time
	Unsized string
}

// Uptime says how long the host has run: in days past two, else in
// hours, else in minutes.
func (d *HostDetail) Uptime() string {
	up := d.System.UptimeDuration()
	switch {
	case up >= 48*time.Hour:
		return strconv.Itoa(int(up/(24*time.Hour))) + " days"
	case up >= 2*time.Hour:
		return strconv.Itoa(int(up/time.Hour)) + " hours"
	}
	return strconv.Itoa(int(up/time.Minute)) + " minutes"
}

// SizeNote says why a container has no size.
func (d *HostDetail) SizeNote() string {
	if !d.Sized {
		return d.Unsized
	}
	return "docker ps does not list the container"
}

// MeasureNote says why a container has no measures.
func (d *HostDetail) MeasureNote() string {
	if !d.Measured {
		return "Beszel reports no containers of the host"
	}
	return "Beszel reports no measures of the container"
}
