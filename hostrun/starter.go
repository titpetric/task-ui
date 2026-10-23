// Package hostrun starts the runs of @host:<host>: a command on hosts of
// the inventory, over the ssh login task-ui has to each, as ordinary
// runs of the root's host jobs, one per host, so each has a row, a
// recording, a place in the recents and the run history, and can run
// again.
package hostrun

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/sshprobe"
)

// Timeout is how long a command runs on one host when the starter sets
// no other.
const Timeout = 2 * time.Minute

// The errors a start ends in.
var (
	// ErrUnknownHost refuses a target the inventory doesn't list.
	ErrUnknownHost = errors.New("no such host in the inventory")
	// ErrNoHosts refuses a start that reaches no host: * with no ssh
	// login to any.
	ErrNoHosts = errors.New("task-ui has no ssh login to any host")
	// ErrNoCommand refuses a start without a command.
	ErrNoCommand = errors.New("enter a command")
)

// Starter starts host runs through the runner.
type Starter struct {
	// Catalogue finds the root, _, whose host jobs the runs are.
	Catalogue interface {
		Find(id string) *folder.Folder
	}
	// Hosts is the inventory, and SSH who task-ui signs in to each host
	// as and the ssh command a run starts.
	Hosts  *beszel.Inventory
	SSH    *sshprobe.Prober
	Runner *runner.Manager
	// Timeout bounds the command on one host, Timeout unless set; Wait
	// bounds finding the logins, ten seconds unless set.
	Timeout time.Duration
	Wait    time.Duration
}

// Resolve returns the hosts the targets name, in the inventory's order,
// and the login to each: the hosts named, or for * or none every host
// with a login.
func (s *Starter) Resolve(ctx context.Context, targets []string) ([]string, map[string]string, error) {
	if s.Hosts == nil || s.SSH == nil {
		return nil, nil, beszel.ErrOff
	}
	systems, err := s.Hosts.Systems(ctx)
	if err != nil {
		return nil, nil, err
	}
	known := map[string]bool{}
	var all []string
	for _, system := range beszel.ByCapacity(systems) {
		known[system.Host] = true
		all = append(all, system.Host)
	}
	every := len(targets) == 0
	var hosts []string
	for _, target := range targets {
		switch {
		case target == "*":
			every = true
		case !known[target]:
			return nil, nil, fmt.Errorf("%w: %s", ErrUnknownHost, target)
		default:
			hosts = append(hosts, target)
		}
	}
	if every {
		hosts = all
	}
	wait := s.Wait
	if wait <= 0 {
		wait = 10 * time.Second
	}
	logins := s.SSH.Logins(ctx, hosts, wait)
	if every {
		var reachable []string
		for _, host := range hosts {
			if logins[host] != "" {
				reachable = append(reachable, host)
			}
		}
		hosts = reachable
	}
	if len(hosts) == 0 {
		return nil, nil, ErrNoHosts
	}
	return hosts, logins, nil
}

// Start starts the command on each host the request names as the user,
// one run each, at once, and returns the batch's ID, "" without one,
// and the sessions, a run of @host:<host> a host. It skips a host
// already running a command; a host without a login gets a run that
// fails saying so. The caller checks the user may run on hosts.
func (s *Starter) Start(ctx context.Context, user *model.User, req Request) (string, []*runner.Session, error) {
	command := strings.TrimSpace(req.Command)
	if command == "" {
		return "", nil, ErrNoCommand
	}
	var root *folder.Folder
	if s.Catalogue != nil {
		root = s.Catalogue.Find(model.Default)
	}
	if root.HostJob("*") == nil {
		return "", nil, errors.New("the root folder has no host jobs")
	}
	hosts, logins, err := s.Resolve(ctx, req.Targets)
	if err != nil {
		return "", nil, err
	}
	batch := ""
	if req.Batch || len(hosts) > 1 {
		batch = model.NewID()
	}
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = Timeout
	}
	var sessions []*runner.Session
	var errs []error
	for _, host := range hosts {
		job := root.HostJob(host)
		job.Command = s.command(host, logins[host], command)
		session, err := s.Runner.Start(ctx, runner.Request{
			Folder: root, Job: job, User: user, Target: host, BatchID: batch, Args: []string{command}, Timeout: timeout,
			Scheduled: req.Scheduled, Force: req.Force, Reason: req.Reason, Cols: req.Cols, Rows: req.Rows,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", host, err))
			continue
		}
		sessions = append(sessions, session)
	}
	if len(sessions) == 0 {
		return "", nil, errors.Join(errs...)
	}
	return batch, sessions, nil
}

// command is the argv of a host's run: ssh on its login, which says
// so when ssh itself fails, quiet as it runs; or, without a login, a
// command that fails saying so.
func (s *Starter) command(host, login, command string) []string {
	if login == "" {
		return []string{"sh", "-c", `echo "task-ui has no ssh login to $0" >&2; exit 255`, host}
	}
	wrap := `"$@"; code=$?; [ "$code" -ne 255 ] || echo "ssh to ` + login + ` ended with 255: it could not sign in or run the command, or the command exited 255" >&2; exit "$code"`
	return append([]string{"sh", "-c", wrap, "sh"}, s.SSH.Argv(login, command)...)
}
