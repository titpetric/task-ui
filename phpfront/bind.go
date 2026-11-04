package phpfront

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/titpetric/phpscript/model"
	"github.com/titpetric/phpscript/runner"

	"github.com/titpetric/task-ui/internal/assets"
	"github.com/titpetric/task-ui/markdown"
)

// The calls are functions and not constants: the handler freezes the standard
// library when it builds a runtime and the reset between requests drops a
// constant registered after that, while the host function table survives.
//
// The name is namespaced. A root class needs a leading backslash from inside
// the tree's own Task namespace, and `phpscript fmt` strips one off a
// single-segment name, which repoints the call at Task\Task. See
// docs/todo-correction.md.
//
// Task\Host::publish and Task\Host::dispatch are the only calls that change
// anything: the one tells every open page a row moved, the other starts a
// run. Every write to a row goes through code/Task over the instance's
// database instead, and the spelling is the contract: Task\Host::x() is the
// host, every other Task\ name is not.

// HostClass is the class the host calls hang off.
const HostClass = `Task\Host`

// clockLayout is the shape the clock answers in: RFC 3339, to the nanosecond,
// which is the one layout a Go time round trips through text without losing
// the zone. The zone is the point — view.NewActivity counts a chart's days in
// the zone of the clock it was handed, so a stamp that arrived as UTC would
// draw another window than the Go page over the same rows.
const clockLayout = time.RFC3339Nano

// Bind registers the host calls on a runtime serving one request.
func Bind(rt *runner.Runtime, opts Options, static *assets.Tree) {
	// A call, not a constant, so dropping the prefix is one Go line.
	rt.RegisterFunc(HostClass+"::base", func() (string, error) {
		return opts.Base, nil
	})

	rt.RegisterFunc(HostClass+"::version", func() (string, error) {
		return opts.Version, nil
	})

	// The instance's clock and not the runtime's: TASK_UI_NOW pins what the
	// hours gate and every page reads, and a chart of the last thirty days
	// ends where this says it does.
	rt.RegisterFunc(HostClass+"::now", func() (string, error) {
		now := time.Now
		if opts.Now != nil {
			now = opts.Now
		}

		return now().Format(clockLayout), nil
	})

	// Whether there is a session to end: the Guard's answer, not a page's.
	rt.RegisterFunc(HostClass+"::login", func() (bool, error) {
		return opts.Login, nil
	})

	// The static files sit at the root, outside the base, so an asset URL
	// survives the cutover unchanged. The version is their content hash,
	// which the URLs carry as ?v= so an hour of caching never outlives a
	// deploy.
	rt.RegisterFunc(HostClass+"::assets", func() (string, error) {
		return strings.TrimSuffix(assets.Prefix, "/"), nil
	})
	rt.RegisterFunc(HostClass+"::assetVersion", func() (string, error) {
		return static.Version(), nil
	})

	// The name the Guard resolved; the row behind it is Task\User.
	rt.RegisterFunc(HostClass+"::whoami", func(ctx context.Context) (string, error) {
		if opts.Host == nil {
			return "", nil
		}
		return opts.Host.Whoami(ctx), nil
	})

	// The catalogue, which nothing records: it is read off the checkouts.
	rt.RegisterFunc(HostClass+"::folders", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return folders(opts.Host.Folders(ctx)), nil
	})

	// One job of the catalogue, which a listing cannot answer: the jobs of a
	// directory inside a folder are found by reading that directory, and the
	// four every folder has it declares nowhere. A run names all three parts,
	// so the gate of a run's page asks here.
	rt.RegisterFunc(HostClass+"::job", func(ctx context.Context, folder, dir, id string) (*model.Array, error) {
		if opts.Host == nil {
			return job(Job{}), nil
		}
		return job(opts.Host.Job(ctx, folder, dir, id)), nil
	})

	// The directories right inside a named workspace's src, as folder IDs.
	// A checkout with no pipeline file is in no catalogue and still has a
	// page, and the sidebar lists it beside the catalogue's folders. The
	// workspaces are rows, so PHP names the one it is asking about.
	rt.RegisterFunc(HostClass+"::projects", func(ctx context.Context, workspace string) ([]string, error) {
		if opts.Host == nil {
			return nil, nil
		}
		return opts.Host.Projects(ctx, workspace), nil
	})

	// The directory a folder ID names, which is the catalogue's for a folder
	// it holds and the layout's for a project it does not: what dirty and
	// commit are asked about.
	rt.RegisterFunc(HostClass+"::folderDir", func(ctx context.Context, id string, workspaces any) (string, error) {
		if opts.Host == nil {
			return "", nil
		}
		return opts.Host.FolderDir(ctx, id, names(workspaces)), nil
	})

	// The same, for every row of a tree at once: the chrome draws the pins
	// and every workspace's projects, and asking per row handed the whole
	// workspace list across the boundary again each time.
	//
	// A list parameter is any and not *model.Array throughout, because the
	// caller here asks with array_keys, which answers a Go slice and not a
	// script array: see names.
	rt.RegisterFunc(HostClass+"::folderDirs", func(ctx context.Context, ids, workspaces any) (*model.Array, error) {
		if opts.Host == nil {
			return model.NewArray(), nil
		}
		return dirs(opts.Host.FolderDirs(ctx, names(ids), names(workspaces))), nil
	})

	// Whether a directory inside a folder is there, which is what a pin of a
	// directory is resolved against. The folders sit outside the directory
	// the PHP may open.
	rt.RegisterFunc(HostClass+"::isDir", func(ctx context.Context, folder, rel string) (bool, error) {
		if opts.Host == nil {
			return false, nil
		}
		return opts.Host.IsDir(ctx, folder, rel), nil
	})

	// git's answer again, through the cache both front ends share: a project
	// costs one git process a minute between them.
	rt.RegisterFunc(HostClass+"::commit", func(ctx context.Context, dir string) (string, error) {
		if opts.Host == nil {
			return "", nil
		}
		return opts.Host.Commit(ctx, dir), nil
	})

	// The runner's memory, where a run lives until it ends.
	rt.RegisterFunc(HostClass+"::running", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return running(opts.Host.Running(ctx)), nil
	})

	// The workflow runner's memory, which decides what counts as running: a
	// row left behind by a process that died is a row and not a run.
	rt.RegisterFunc(HostClass+"::flows", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return flows(opts.Host.Flows(ctx)), nil
	})

	// The run the runner holds, which is ahead of the row: a page reads the
	// row and lays this over the fields a session decides.
	rt.RegisterFunc(HostClass+"::session", func(ctx context.Context, id string) (*model.Array, error) {
		if opts.Host == nil {
			return session(Session{}), nil
		}
		return session(opts.Host.Session(ctx, id)), nil
	})

	// The terminal of a live run, which is in no table: the last screen as
	// the SVG the page embeds.
	rt.RegisterFunc(HostClass+"::frame", func(ctx context.Context, id string) (string, error) {
		if opts.Host == nil {
			return "", nil
		}
		return opts.Host.Frame(ctx, id), nil
	})

	// The terminal of a run that has ended, replayed off its recording: the
	// screen a pipeline draws under a step, with the lines the whole output
	// has. The columns come over because the row is the caller's.
	rt.RegisterFunc(HostClass+"::screen", func(ctx context.Context, id, castPath string, cols int) (*model.Array, error) {
		if opts.Host == nil {
			return screen(Screen{}), nil
		}
		return screen(opts.Host.Screen(ctx, id, castPath, cols)), nil
	})

	// Whether a recording is still on disk, which the retention sweep
	// decides and no column records.
	rt.RegisterFunc(HostClass+"::expired", func(ctx context.Context, castPath, state string) (bool, error) {
		if opts.Host == nil {
			return false, nil
		}
		return opts.Host.Expired(ctx, castPath, state), nil
	})

	// Whether the reader has a token for Claude, which is a secret of the
	// integration store and not a row the front end reads.
	rt.RegisterFunc(HostClass+"::claude", func(ctx context.Context) (bool, error) {
		if opts.Host == nil {
			return false, nil
		}
		return opts.Host.Claude(ctx), nil
	})

	// git's answer, not a row.
	rt.RegisterFunc(HostClass+"::dirty", func(ctx context.Context, dir string) (int, error) {
		if opts.Host == nil {
			return 0, nil
		}
		return opts.Host.Dirty(ctx, dir), nil
	})

	// config/permissions.yml. Task\Permission evaluates it over the policy rows.
	rt.RegisterFunc(HostClass+"::rules", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return rules(Rules{}), nil
		}
		return rules(opts.Host.Rules(ctx)), nil
	})

	// A write that does not publish leaves every open page stale. The payload
	// is read as any and not as a Go map, because coerceArg maps no PHP array
	// onto one, and not as *model.Array either: see names.
	rt.RegisterFunc(HostClass+"::publish", func(ctx context.Context, topic string, payload any) (bool, error) {
		if opts.Host == nil {
			return false, nil
		}
		if err := opts.Host.Publish(ctx, topic, fields(payload)); err != nil {
			return false, err
		}
		return true, nil
	})

	// A run over a comment and the markdown saying what to do with it, and
	// the note it opened. The action is a path in this tree, so a page
	// ships what it dispatches beside itself: a thread's correction and the
	// instructions that apply it are one directory.
	//
	// An empty comment is refused rather than dispatched: a run with
	// nothing asked of it is a run that costs money and answers the action
	// file back. The plugin host refuses the same call for the same reason.
	rt.RegisterFunc(HostClass+"::dispatch", func(ctx context.Context, comment, action, folder string) (string, error) {
		if opts.Host == nil {
			return "", errors.New("this instance starts no runs")
		}
		if strings.TrimSpace(comment) == "" {
			return "", errors.New(HostClass + "::dispatch takes the comment to act on")
		}
		return opts.Host.Dispatch(ctx, comment, action, folder)
	})

	// Which run a dispatch turned into, so a page draws the status label
	// the interface shares: <span data-run-id="..."></span>, which the
	// site's own script fills and keeps current. "" is a run that has not
	// started, which is every dispatch until the board gets to it.
	rt.RegisterFunc(HostClass+"::run", func(ctx context.Context, note string) (string, error) {
		if opts.Host == nil || strings.TrimSpace(note) == "" {
			return "", nil
		}
		return opts.Host.RunOf(ctx, note), nil
	})

	// The renderer that owns the content security policy, with what a
	// relative link in the text resolves against: the prefix the files sit
	// under and the directory the text was written in. Both empty is text
	// whose relative links resolve nowhere and stay as text, which is what
	// a report of a run is - the same two arguments frontend/view.Markdown
	// passes.
	rt.RegisterFunc(HostClass+"::markdown", func(src, prefix, dir string) (string, error) {
		return markdown.HTML([]byte(src), prefix, dir), nil
	})

	// A note's text, which is the same renderer with a line break kept and
	// @name marked. A name in link becomes a link under prefix, and every
	// other mention is marked and not linked.
	rt.RegisterFunc(HostClass+"::message", func(src, prefix string, link any) (string, error) {
		return Message(src, prefix, names(link)), nil
	})

	// The artifact tree, which sits outside the directory the PHP may open.
	rt.RegisterFunc(HostClass+"::artifacts", func(folder string) ([]string, error) {
		return opts.Artifacts.Files(folder), nil
	})
	rt.RegisterFunc(HostClass+"::artifact", func(rel string) (string, error) {
		return opts.Artifacts.Read(rel), nil
	})

	// One directory of the tree, which the page lists a level at a time, and
	// the space the whole of it takes.
	rt.RegisterFunc(HostClass+"::tree", func(ctx context.Context, dir string) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return treeItems(opts.Host.Tree(ctx, dir)), nil
	})
	rt.RegisterFunc(HostClass+"::space", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return space(Space{}), nil
		}
		return space(opts.Host.Space(ctx)), nil
	})
	rt.RegisterFunc(HostClass+"::spaces", func(ctx context.Context, dirs any) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return spaces(opts.Host.Spaces(ctx, names(dirs))), nil
	})

	// The hosts: beszel's hub, the ssh prober and hostinfo, none of them a
	// row. The hosts permission is the reader's, which PHP decides, and it
	// decides whether task-ui reads a host now or answers from what it kept.
	rt.RegisterFunc(HostClass+"::inventory", func(ctx context.Context, hosts bool) (*model.Array, error) {
		if opts.Host == nil {
			return inventory(Inventory{Off: true}), nil
		}
		return inventory(opts.Host.Inventory(ctx, hosts)), nil
	})
	rt.RegisterFunc(HostClass+"::host", func(ctx context.Context, name string, hosts bool) (*model.Array, error) {
		if opts.Host == nil {
			return facts(HostFacts{}), nil
		}
		return facts(opts.Host.HostFacts(ctx, name, hosts)), nil
	})
	rt.RegisterFunc(HostClass+"::reachable", func(ctx context.Context) ([]string, error) {
		if opts.Host == nil {
			return nil, nil
		}
		return opts.Host.Reachable(ctx), nil
	})

	// What atkins recorded for the tree, which it answers by running atkins.
	rt.RegisterFunc(HostClass+"::tooling", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return toolingScopes(opts.Host.Tooling(ctx)), nil
	})

	// A folder of the catalogue, or a directory inside one, in the depth the
	// page about it draws, and the Jobs table of it: the detection of a
	// directory and the lines of a file, neither of them a row.
	rt.RegisterFunc(HostClass+"::place", func(ctx context.Context, folder, rel string) (*model.Array, error) {
		if opts.Host == nil {
			return place(Place{}), nil
		}
		return place(opts.Host.Place(ctx, folder, rel)), nil
	})
	rt.RegisterFunc(HostClass+"::groups", func(ctx context.Context, folder, rel string, seen any, shell, claude bool) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return groups(opts.Host.Groups(ctx, folder, rel, names(seen), shell, claude)), nil
	})

	// A directory of a folder as the Files panel lists it, and when a
	// directory last changed, for a folder the listing above it does not
	// hold. The folders sit outside the directory the PHP may open.
	rt.RegisterFunc(HostClass+"::listing", func(ctx context.Context, folder, rel string, seen any) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return entries(opts.Host.Listing(ctx, folder, rel, names(seen))), nil
	})
	rt.RegisterFunc(HostClass+"::rootTree", func(ctx context.Context, named bool, seen any) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return entries(opts.Host.RootTree(ctx, named, names(seen))), nil
	})
	rt.RegisterFunc(HostClass+"::modTime", func(ctx context.Context, dir string) (string, error) {
		if opts.Host == nil {
			return "", nil
		}
		return opts.Host.ModTime(ctx, dir), nil
	})

	// What git reports of a checkout: its state now, and the commit it is at
	// with the space it takes.
	rt.RegisterFunc(HostClass+"::status", func(ctx context.Context, dir string) (*model.Array, error) {
		if opts.Host == nil {
			return status(Status{}), nil
		}
		return status(opts.Host.Status(ctx, dir)), nil
	})
	rt.RegisterFunc(HostClass+"::checkout", func(ctx context.Context, dir string) (*model.Array, error) {
		if opts.Host == nil {
			return checkout(Checkout{}), nil
		}
		return checkout(opts.Host.Checkout(ctx, dir)), nil
	})

	// An hours rule read, whose grammar is a Go parser.
	rt.RegisterFunc(HostClass+"::hours", func(ctx context.Context, text string) (*model.Array, error) {
		if opts.Host == nil {
			return hours(Hours{}), nil
		}
		return hours(opts.Host.Hours(ctx, text)), nil
	})

	// What the settings read of the instance itself: the key folder, the
	// configuration the server runs with, what the retention rules apply to,
	// the tracer's options, the plugins and the project directories.
	rt.RegisterFunc(HostClass+"::keys", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return keystore(Keystore{}), nil
		}
		return keystore(opts.Host.Keys(ctx)), nil
	})
	rt.RegisterFunc(HostClass+"::keyFolders", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return keyFolders(nil), nil
		}
		return keyFolders(opts.Host.KeyFolders(ctx)), nil
	})
	rt.RegisterFunc(HostClass+"::instance", func(ctx context.Context, admin bool) (*model.Array, error) {
		if opts.Host == nil {
			return instance(Instance{}), nil
		}
		return instance(opts.Host.Instance(ctx, admin)), nil
	})
	rt.RegisterFunc(HostClass+"::usage", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return usage(Retention{}), nil
		}
		return usage(opts.Host.Usage(ctx)), nil
	})
	rt.RegisterFunc(HostClass+"::telemetry", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return telemetry(Telemetry{}), nil
		}
		return telemetry(opts.Host.Telemetry(ctx)), nil
	})
	rt.RegisterFunc(HostClass+"::integrations", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return integrations(opts.Host.Services(ctx)), nil
	})
	rt.RegisterFunc(HostClass+"::variables", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return variables(Variables{}), nil
		}
		return variables(opts.Host.Variables(ctx)), nil
	})
	rt.RegisterFunc(HostClass+"::variable", func(ctx context.Context, id string) (*model.Array, error) {
		if opts.Host == nil {
			return variable(Variable{}), nil
		}
		return variable(opts.Host.Variable(ctx, id)), nil
	})
	rt.RegisterFunc(HostClass+"::integrationEnv", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return envVars(opts.Host.IntegrationEnv(ctx)), nil
	})
	rt.RegisterFunc(HostClass+"::modules", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return plugins(opts.Host.Modules(ctx)), nil
	})
	rt.RegisterFunc(HostClass+"::entries", func(ctx context.Context, workspaces any) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return projects(opts.Host.Entries(ctx, names(workspaces))), nil
	})

	rt.RegisterClass(&model.Class{Name: HostClass})
}

// database is what `new Database("task-ui")` resolves through.
func database(opts Options) model.DatabaseProvider {
	if opts.Database == nil {
		return nil
	}
	return opts.Database
}
