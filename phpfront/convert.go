package phpfront

import (
	"github.com/titpetric/phpscript/model"
)

// A Go slice crosses into the runtime as a PHP array, and a Go struct crosses
// as a host object: its fields reach a script through published accessors and
// not through a key, so `$value["id"]` finds nothing and `get_object_vars`
// answers empty.
//
// Everything the host answers is read by a template, which indexes by key, and
// by the DAOs, which already speak arrays. So the boundary converts: what
// leaves here is an array all the way down, keyed by the names host.go
// declares as json tags.

// phpMap is a PHP array of alternating keys and values.
func phpMap(pairs ...any) *model.Array {
	out := model.NewArraySize(len(pairs) / 2)
	for i := 0; i+1 < len(pairs); i += 2 {
		out.Set(pairs[i], pairs[i+1])
	}
	return out
}

// phpList is a PHP array of values, keyed by position.
func phpList(values ...any) *model.Array {
	out := model.NewArraySize(len(values))
	for _, value := range values {
		out.Append(value)
	}
	return out
}

// strings is a list of strings as a PHP array.
func texts(values []string) *model.Array {
	out := model.NewArraySize(len(values))
	for _, value := range values {
		out.Append(value)
	}
	return out
}

// folders is the catalogue as a PHP array.
func folders(found []Folder) *model.Array {
	out := model.NewArraySize(len(found))
	for _, f := range found {
		out.Append(phpMap(
			"id", f.ID,
			"title", f.Title,
			"dir", f.Dir,
			"allow", texts(f.Allow),
			"sources", texts(f.Sources),
			"error", f.Error,
			"errors", findings(f.Errors),
			"jobs", jobs(f.Jobs),
		))
	}
	return out
}

// findings is what atkins' linter said about a folder's pipelines as a PHP
// array.
func findings(found []Finding) *model.Array {
	out := model.NewArraySize(len(found))
	for _, finding := range found {
		out.Append(phpMap("path", finding.Path, "job", finding.Job, "issue", finding.Issue, "detail", finding.Detail))
	}
	return out
}

// jobs is the jobs a folder declares as a PHP array. The folder's own allow
// list is the folder's key and not repeated under each of its jobs, so a page
// listing a catalogue reads the pair the gate takes from the two levels.
func jobs(found []Job) *model.Array {
	out := model.NewArraySize(len(found))
	for _, j := range found {
		out.Append(phpMap("id", j.ID, "name", j.Name, "dir", j.Dir, "shell", j.Shell, "host", j.Host, "allow", texts(j.Allow)))
	}
	return out
}

// job is one job as the catalogue resolved it, the whole gate in one answer:
// whether it is there at all, the two kinds a permission gates, and both allow
// lists, because the page that gates a run holds neither the folder nor the
// directory inside it the run named.
func job(found Job) *model.Array {
	return phpMap(
		"found", found.Found,
		"id", found.ID,
		"name", found.Name,
		"dir", found.Dir,
		"shell", found.Shell,
		"host", found.Host,
		"allow", texts(found.Allow),
		"folder_allow", texts(found.FolderAllow),
		"kind", found.Kind,
		"source", found.Source,
		"source_lines", found.SourceLines,
		"description", found.Description,
		"summary", found.Summary,
		"command", texts(found.Command),
		"steps", texts(found.Steps),
		"aliases", texts(found.Aliases),
		"depends", depends(found.Depends),
		"interactive", found.Interactive,
		"takes_args", found.TakesArgs,
		"args_command", found.ArgsCommand,
		"key", found.Key,
	)
}

// depends is the targets a job runs first as a PHP array.
func depends(found []Depends) *model.Array {
	out := model.NewArraySize(len(found))
	for _, dep := range found {
		out.Append(phpMap("text", dep.Text, "found", dep.Found, "id", dep.ID, "name", dep.Name, "dir", dep.Dir))
	}
	return out
}

// place is a folder, or a directory inside one, as a PHP array.
func place(found Place) *model.Array {
	return phpMap(
		"found", found.Found,
		"id", found.ID,
		"title", found.Title,
		"description", found.Description,
		"dir", found.Dir,
		"rel", found.Rel,
		"allow", texts(found.Allow),
		"hours", found.Hours,
		"error", found.Error,
		"errors", findings(found.Errors),
		"sources", sources(found.Sources),
		"includes", files(found.Includes),
	)
}

// sources is the pipeline files a place was detected in as a PHP array.
func sources(found []Source) *model.Array {
	out := model.NewArraySize(len(found))
	for _, source := range found {
		out.Append(phpMap(
			"kind", source.Kind,
			"file", source.File,
			"command", texts(source.Command),
			"lines", source.Lines,
		))
	}
	return out
}

// files is a list of files with the lines each holds as a PHP array.
func files(found []File) *model.Array {
	out := model.NewArraySize(len(found))
	for _, file := range found {
		out.Append(phpMap("name", file.Name, "lines", file.Lines))
	}
	return out
}

// groups is the Jobs table of a place as a PHP array.
func groups(found []Group) *model.Array {
	out := model.NewArraySize(len(found))
	for _, g := range found {
		out.Append(phpMap(
			"kind", g.Kind,
			"note", g.Note,
			"source", g.Source,
			"dir", g.Dir,
			"empty", g.Empty,
			"targets", g.Targets,
			"folder", where(g.Folder),
			"jobs", targets(g.Jobs),
		))
	}
	return out
}

// where is the folder a row's jobs run as as a PHP array. The keys are the
// ones Task\Folders::allows reads of a folder, so a row's gate is the folder's
// gate and not a second rule.
func where(found Where) *model.Array {
	return phpMap("id", found.ID, "rel", found.Rel, "allow", texts(found.Allow), "hours", found.Hours)
}

// targets is the jobs of one row of the Jobs table as a PHP array: what a
// button says, what it starts and what gates it, and none of the detail only
// the job's own page draws.
func targets(found []Job) *model.Array {
	out := model.NewArraySize(len(found))
	for _, j := range found {
		out.Append(phpMap(
			"id", j.ID,
			"name", j.Name,
			"dir", j.Dir,
			"kind", j.Kind,
			"source", j.Source,
			"description", j.Description,
			"key", j.Key,
			"interactive", j.Interactive,
			"takes_args", j.TakesArgs,
			"args_command", j.ArgsCommand,
			"shell", j.Shell,
			"host", j.Host,
			"allow", texts(j.Allow),
			"confirm", j.Confirm,
		))
	}
	return out
}

// entries is a directory the Files panel lists as a PHP array.
func entries(items []Entry) *model.Array {
	out := model.NewArraySize(len(items))
	for _, item := range items {
		out.Append(phpMap(
			"name", item.Name,
			"dir", item.Dir,
			"size", item.Size,
			"mod_time", item.ModTime,
			"path", item.Path,
			"folder", item.Folder,
			"sealed", item.Sealed,
			"exec", item.Exec,
		))
	}
	return out
}

// status is what git reports of a checkout as a PHP array.
func status(found Status) *model.Array {
	return phpMap(
		"found", found.Found,
		"checkout", found.Checkout,
		"branch", found.Branch,
		"detached", found.Detached,
		"upstream", found.Upstream,
		"ahead", found.Ahead,
		"behind", found.Behind,
		"dirty", found.Dirty,
		"changes", changes(found.Changes),
	)
}

// changes is the files a checkout has changed as a PHP array.
func changes(found []Change) *model.Array {
	out := model.NewArraySize(len(found))
	for _, change := range found {
		out.Append(phpMap("code", change.Code, "state", change.State, "path", change.Path, "from", change.From))
	}
	return out
}

// checkout is the commit a folder is at and the space it takes as a PHP array.
func checkout(found Checkout) *model.Array {
	return phpMap(
		"found", found.Found,
		"remote", found.Remote,
		"branch", found.Branch,
		"commit", found.Commit,
		"subject", found.Subject,
		"when", found.When,
		"space", space(found.Space),
	)
}

// hours is an hours rule read as a PHP array.
func hours(read Hours) *model.Array {
	return phpMap("line", read.Line, "open", read.Open)
}

// running is the runs in progress as a PHP array.
func running(found []Running) *model.Array {
	out := model.NewArraySize(len(found))
	for _, r := range found {
		out.Append(phpMap("key", r.Key, "id", r.ID, "state", r.State))
	}
	return out
}

// flows is the workflow runs in progress as a PHP array.
func flows(found []Flow) *model.Array {
	out := model.NewArraySize(len(found))
	for _, f := range found {
		out.Append(phpMap("workflow_id", f.WorkflowID, "run_id", f.RunID))
	}
	return out
}

// session is what the runner holds about a run as a PHP array.
func session(held Session) *model.Array {
	return phpMap(
		"live", held.Live,
		"attached", held.Attached,
		"message", held.Message,
		"state", held.State,
		"exit_code", held.ExitCode,
		"title", held.Title,
		"finished_at", held.FinishedAt,
	)
}

// screen is the replayed last frame of a run as a PHP array.
func screen(shown Screen) *model.Array {
	return phpMap(
		"svg", shown.SVG,
		"lines", shown.Lines,
		"complete", shown.Complete,
	)
}

// rules is the permission file as a PHP array.
func rules(file Rules) *model.Array {
	return phpMap(
		"rules", titles(file.Rules),
		"sections", titles(file.Sections),
		"permissions", titles(file.Permissions),
		"paths", texts(file.Paths),
		"checks", texts(file.Checks),
		"extras", titles(file.Extras),
	)
}

// titles is a titled list of the permission file as a PHP array.
func titles(entries []Rule) *model.Array {
	out := model.NewArraySize(len(entries))
	for _, entry := range entries {
		out.Append(phpMap("key", entry.Key, "title", entry.Title))
	}
	return out
}

// tree is one directory of the artifact tree as a PHP array.
func treeItems(items []TreeItem) *model.Array {
	out := model.NewArraySize(len(items))
	for _, item := range items {
		out.Append(phpMap("name", item.Name, "dir", item.Dir, "size", item.Size, "mod_time", item.ModTime))
	}
	return out
}

// space is what a tree takes on disk as a PHP array.
func space(usage Space) *model.Array {
	return phpMap("measured", usage.Measured, "files", usage.Files, "history", usage.History, "ignored", usage.Ignored)
}

// spaces is what each of a page's directories takes as a PHP array, by the
// directory itself.
func spaces(found map[string]Space) *model.Array {
	out := model.NewArraySize(len(found))
	for dir, usage := range found {
		out.Set(dir, space(usage))
	}
	return out
}

// inventory is the hosts page's facts as a PHP array.
func inventory(found Inventory) *model.Array {
	return phpMap(
		"off", found.Off,
		"error", found.Error,
		"hosts", systems(found.Hosts),
		"logins", logins(found.Logins),
		"cluster", cluster(found.Cluster),
		"history", history(found.History),
	)
}

// systems is the hosts Beszel monitors as a PHP array.
func systems(found []System) *model.Array {
	out := model.NewArraySize(len(found))
	for _, each := range found {
		out.Append(system(each))
	}
	return out
}

// system is one host as Beszel reports it.
func system(found System) *model.Array {
	return phpMap(
		"id", found.ID,
		"name", found.Name,
		"host", found.Host,
		"status", found.Status,
		"up", found.Up,
		"cpus", found.CPUs,
		"agent", found.Agent,
		"uptime", found.Uptime,
		"cpu", found.CPU,
		"memory", found.Memory,
		"disk", found.Disk,
		"measured", found.Measured,
		"memory_total", found.MemoryTotal,
		"memory_used", found.MemoryUsed,
		"disk_total", found.DiskTotal,
		"disk_used", found.DiskUsed,
	)
}

// logins is the ssh logins as a PHP array, keyed by host so a page reads one
// by the address of its line.
func logins(found []Login) *model.Array {
	out := model.NewArraySize(len(found))
	for _, login := range found {
		out.Set(login.Host, login.Login)
	}
	return out
}

// cluster is the pools as a PHP array.
func cluster(found Cluster) *model.Array {
	return phpMap(
		"up", found.Up,
		"down", found.Down,
		"left", found.Left,
		"pools", pools(found.Pools),
		"hosts", legend(found.Hosts),
	)
}

// pools is the resources the hosts share as a PHP array.
func pools(found []Pool) *model.Array {
	out := model.NewArraySize(len(found))
	for _, pool := range found {
		out.Append(phpMap(
			"name", pool.Name,
			"unit", pool.Unit,
			"figure", pool.Figure,
			"used", pool.Used,
			"total", pool.Total,
			"percent", pool.Percent,
			"segments", segments(pool.Segments),
		))
	}
	return out
}

// segments is a pool's bar as a PHP array.
func segments(found []Segment) *model.Array {
	out := model.NewArraySize(len(found))
	for _, segment := range found {
		out.Append(phpMap(
			"host", segment.Host,
			"color", segment.Color,
			"width", segment.Width,
			"fill", segment.Fill,
			"title", segment.Title,
		))
	}
	return out
}

// legend is the pools' legend as a PHP array.
func legend(found []ClusterHost) *model.Array {
	out := model.NewArraySize(len(found))
	for _, entry := range found {
		out.Append(phpMap("name", entry.Name, "color", entry.Color, "left", entry.Left))
	}
	return out
}

// history is Beszel's records of each host as a PHP array, keyed by host ID.
func history(found []Series) *model.Array {
	out := model.NewArraySize(len(found))
	for _, series := range found {
		out.Set(series.ID, samples(series.Samples))
	}
	return out
}

// samples is a host's records as a PHP array.
func samples(found []Sample) *model.Array {
	out := model.NewArraySize(len(found))
	for _, sample := range found {
		out.Append(phpMap(
			"at", sample.At,
			"cpu", sample.CPU,
			"memory", sample.Memory,
			"memory_used", sample.MemoryUsed,
			"disk", sample.Disk,
			"disk_used", sample.DiskUsed,
			"sent", sample.Sent,
			"recv", sample.Recv,
			"extra", mounts(sample.Extra),
		))
	}
	return out
}

// mounts is the filesystems of a record as a PHP array.
func mounts(found []Mount) *model.Array {
	out := model.NewArraySize(len(found))
	for _, mount := range found {
		out.Append(phpMap("name", mount.Name, "used", mount.Used, "total", mount.Total))
	}
	return out
}

// facts is one host of the inventory in full as a PHP array.
func facts(found HostFacts) *model.Array {
	return phpMap(
		"found", found.Found,
		"error", found.Error,
		"system", system(found.System),
		"login", found.Login,
		"containers", containers(found.Containers),
		"measured", found.Measured,
		"sized", found.Sized,
		"unsized", found.Unsized,
		"filesystems", filesystems(found.Filesystems),
		"from_df", found.FromDF,
		"read", found.Read,
		"samples", samples(found.Samples),
		"container_samples", containerSamples(found.ContainerSamples),
	)
}

// containers is a host's containers as a PHP array.
func containers(found []Container) *model.Array {
	out := model.NewArraySize(len(found))
	for _, container := range found {
		out.Append(phpMap(
			"name", container.Name,
			"image", container.Image,
			"state", container.State,
			"status", container.Status,
			"cpu", container.CPU,
			"memory", container.Memory,
			"memory_share", container.MemoryShare,
			"net", container.Net,
			"size_rw", container.SizeRw,
			"size_virtual", container.SizeVirtual,
			"measured", container.Measured,
			"sized", container.Sized,
		))
	}
	return out
}

// filesystems is a host's durable filesystems as a PHP array.
func filesystems(found []Filesystem) *model.Array {
	out := model.NewArraySize(len(found))
	for _, fs := range found {
		out.Append(phpMap(
			"mount", fs.Mount,
			"device", fs.Device,
			"type", fs.Type,
			"size", fs.Size,
			"used", fs.Used,
			"free", fs.Free,
			"percent", fs.Percent,
		))
	}
	return out
}

// toolingScopes is what atkins recorded as a PHP array.
func toolingScopes(found []ToolingScope) *model.Array {
	out := model.NewArraySize(len(found))
	for _, scope := range found {
		out.Append(phpMap(
			"name", scope.Name,
			"repository", scope.Repository,
			"database", scope.Database,
			"days", scope.Days,
			"from", scope.From,
			"to", scope.To,
			"runs", scope.Runs,
			"total_ms", scope.TotalMs,
			"commands", toolingCommands(scope.Commands),
			"usage", toolingUsage(scope.Usage),
			"command_usage", toolingUsage(scope.CommandUsage),
			"samples", toolingSamples(scope.Samples),
		))
	}
	return out
}

// toolingCommands is a scope's command keys as a PHP array.
func toolingCommands(found []ToolingCommand) *model.Array {
	out := model.NewArraySize(len(found))
	for _, command := range found {
		out.Append(phpMap(
			"command", command.Command,
			"tool", command.Tool,
			"runs", command.Runs,
			"total_ms", command.TotalMs,
			"average_ms", command.AverageMs,
			"max_ms", command.MaxMs,
			"failures", command.Failures,
			"last_run_at", command.LastRunAt,
		))
	}
	return out
}

// toolingUsage is a daily series as a PHP array.
func toolingUsage(found []ToolingUsage) *model.Array {
	out := model.NewArraySize(len(found))
	for _, row := range found {
		out.Append(phpMap(
			"date", row.Date,
			"command", row.Command,
			"tool", row.Tool,
			"runs", row.Runs,
			"total_ms", row.TotalMs,
		))
	}
	return out
}

// toolingSamples is the commands of a window one by one as a PHP array.
func toolingSamples(found []ToolingSample) *model.Array {
	out := model.NewArraySize(len(found))
	for _, sample := range found {
		out.Append(phpMap(
			"at", sample.At,
			"command", sample.Command,
			"tool", sample.Tool,
			"job", sample.Job,
			"duration_ms", sample.DurationMs,
			"exit_code", sample.ExitCode,
			"failed", sample.Failed,
		))
	}
	return out
}

// containerSamples is the records of a host's containers as a PHP array.
func containerSamples(found []ContainerSample) *model.Array {
	out := model.NewArraySize(len(found))
	for _, sample := range found {
		stats := model.NewArraySize(len(sample.Containers))
		for _, stat := range sample.Containers {
			stats.Append(phpMap("name", stat.Name, "cpu", stat.CPU, "memory", stat.Memory))
		}
		out.Append(phpMap("at", sample.At, "containers", stats))
	}
	return out
}

// keystore is the key folder as a PHP array.
func keystore(found Keystore) *model.Array {
	return phpMap("dir", found.Dir, "instance", found.Instance, "keys", keys(found.Keys))
}

// keyFolders is the .ssh folder of every project that holds a key, as a PHP array.
func keyFolders(found []KeyFolder) *model.Array {
	out := model.NewArraySize(len(found))
	for _, folder := range found {
		out.Append(phpMap("project", folder.Project, "dir", folder.Dir, "keys", texts(folder.Keys)))
	}
	return out
}

// keys is every key of the key folder as a PHP array.
func keys(found []Key) *model.Array {
	out := model.NewArraySize(len(found))
	for _, key := range found {
		out.Append(phpMap(
			"name", key.Name,
			"purpose", key.Purpose,
			"scope", key.Scope,
			"algo", key.Algo,
			"bits", key.Bits,
			"size", key.Size,
			"fingerprint", key.Fingerprint,
			"comment", key.Comment,
			"public_key", key.PublicKey,
			"created_at", key.CreatedAt,
			"roles", texts(key.Roles),
			"users", texts(key.Users),
			"git", key.Git,
			"encrypts", key.Encrypts,
			"workspace", key.Workspace,
			"backup", key.Backup,
			"backup_fingerprint", key.BackupFingerprint,
			"rotated_at", key.RotatedAt,
		))
	}
	return out
}

// instance is what the server runs with as a PHP array.
func instance(found Instance) *model.Array {
	return phpMap(
		"url", found.URL,
		"listen", found.Listen,
		"root", found.Root,
		"data_dir", found.DataDir,
		"ssh_dir", found.SSHDir,
		"htpasswd", found.Htpasswd,
		"host_data", found.HostData,
		"discover", found.Discover,
		"auto_login", found.AutoLogin,
		"time_zone", found.TimeZone,
		"session_ttl", found.SessionTTL,
		"keep", found.Keep,
		"grace", found.Grace,
		"artifact_history", found.ArtifactHistory,
		"compose_project", found.ComposeProject,
		"beszel", found.Beszel,
		"running", found.Running,
		"backlog", phpMap(
			"parallel", found.Backlog.Parallel,
			"going", found.Backlog.Going,
			"waiting", found.Backlog.Waiting,
			"auto", found.Backlog.Auto,
			"configured", found.Backlog.Configured,
			"length", found.Backlog.Length,
		),
		"parent_env", found.ParentEnv,
		"env", envNames(found.Env),
		"env_base", texts(found.EnvBase),
		"env_sensitive", texts(found.EnvSensitive),
		"measured", found.Measured,
		"storage", storage(found.Storage),
	)
}

// envNames is the server's own variables as a PHP array.
func envNames(found []EnvName) *model.Array {
	out := model.NewArraySize(len(found))
	for _, name := range found {
		out.Append(phpMap("name", name.Name, "sensitive", name.Sensitive, "none", name.None, "auto", name.Auto))
	}
	return out
}

// storage is the Storage table of the Instance page as a PHP array.
func storage(found Storage) *model.Array {
	rows := model.NewArraySize(len(found.Rows))
	for _, row := range found.Rows {
		rows.Append(storageRow(row))
	}
	return phpMap("rows", rows, "total", storageRow(found.Total), "pending", found.Pending)
}

// storageRow is one item of the Storage table as a PHP array.
func storageRow(row StorageRow) *model.Array {
	return phpMap("name", row.Name, "where", row.Where, "size", row.Size, "files", row.Files, "pending", row.Pending)
}

// usage is what the areas of the Data retention table hold as a PHP array:
// the areas keyed by the kind, which is how a row of the table reads its own.
func usage(found Retention) *model.Array {
	areas := model.NewArraySize(len(found.Areas))
	for _, area := range found.Areas {
		areas.Set(area.Kind, phpMap("kind", area.Kind, "count", area.Count, "bytes", area.Bytes, "oldest", area.Oldest))
	}
	return phpMap("off", found.Off, "error", found.Error, "areas", areas)
}

// telemetry is what the instance records of its own work as a PHP array.
func telemetry(found Telemetry) *model.Array {
	return phpMap(
		"error", found.Error,
		"wired", found.Wired,
		"path", found.Path,
		"enabled", found.Enabled,
		"service_name", found.ServiceName,
		"sample_rate", found.SampleRate,
		"ring_buffer_size", found.RingBufferSize,
		"top_requests", found.TopRequests,
		"max_spans_per_trace", found.MaxSpansPerTrace,
		"ignore_paths", texts(found.IgnorePaths),
		"live_stream", found.LiveStream,
		"refresh_interval", found.RefreshInterval,
		"allowed_networks", texts(found.AllowedNetworks),
		"option_path", found.OptionPath,
		"capture_logs", found.CaptureLogs,
		"track_memory_use", found.TrackMemoryUse,
		"trust_request_id", found.TrustRequestID,
	)
}

// integrations is the services of the Integrations settings as a PHP array.
//
// The three maps are written in the order the fields are declared, not in the
// order a Go map answers: what crosses here is read by key and printed by a
// test, and a page that draws the same settings twice draws them the same way.
func integrations(found []Integration) *model.Array {
	out := model.NewArraySize(len(found))
	for _, service := range found {
		out.Append(phpMap(
			"id", service.ID,
			"title", service.Title,
			"description", service.Description,
			"credential", service.Credential,
			"steps", texts(service.Steps),
			"fields", integrationFields(service.Fields),
			"configured", service.Configured,
			"values", settings(service.Fields, service.Values),
			"set", flags(service.Fields, service.Set),
			"checked", texts(service.Checked),
			"enabled", flags(service.Fields, service.Enabled),
			"error", service.Error,
		))
	}
	return out
}

// variables is the value of every plain env variable as a PHP array, by row
// ID, with what the walk of them said.
//
// The order a Go map answers in is not an order; the page reads a value by
// the ID of the row it is drawing, so what crosses is keyed and never walked.
func variables(found Variables) *model.Array {
	out := model.NewArraySize(len(found.Values))
	for id, value := range found.Values {
		out.Set(id, value)
	}
	return phpMap("values", out, "error", found.Error)
}

// variable is one opened variable as a PHP array.
func variable(found Variable) *model.Array {
	return phpMap("content", found.Content, "error", found.Error)
}

// envVars is what the integrations add to a run as a PHP array, in the order
// the page lists them.
func envVars(found []EnvVar) *model.Array {
	out := model.NewArraySize(len(found))
	for _, item := range found {
		out.Append(phpMap("name", item.Name, "value", item.Value, "redacted", item.Redacted))
	}
	return out
}

// integrationFields is the settings of one service as a PHP array.
func integrationFields(found []IntegrationField) *model.Array {
	out := model.NewArraySize(len(found))
	for _, field := range found {
		out.Append(phpMap(
			"name", field.Name,
			"label", field.Label,
			"help", field.Help,
			"placeholder", field.Placeholder,
			"secret", field.Secret,
			"optional", field.Optional,
			"who", field.Who,
			"options", texts(field.Options),
			"generated", field.Generated,
			"env", field.Env,
			"default", field.Default,
			"toggle", field.Toggle,
			"default_on", field.DefaultOn,
			"empty", field.Empty,
			"source", field.Source,
		))
	}
	return out
}

// settings is what a service's plain settings hold, by field name, in the
// order the fields are declared; a field the map leaves out is left out.
func settings(fields []IntegrationField, values map[string]string) *model.Array {
	out := model.NewArraySize(len(values))
	for _, field := range fields {
		if value, ok := values[field.Name]; ok {
			out.Set(field.Name, value)
		}
	}
	return out
}

// flags is one of a service's maps of a field name to a flag, in the order the
// fields are declared.
func flags(fields []IntegrationField, found map[string]bool) *model.Array {
	out := model.NewArraySize(len(found))
	for _, field := range fields {
		if flag, ok := found[field.Name]; ok {
			out.Set(field.Name, flag)
		}
	}
	return out
}

// plugins is the modules the instance serves as a PHP array.
func plugins(found []Plugin) *model.Array {
	out := model.NewArraySize(len(found))
	for _, module := range found {
		out.Append(phpMap("name", module.Name, "title", module.Title, "description", module.Description, "icon", module.Icon))
	}
	return out
}

// projects is the project directories of the workspaces as a PHP array.
func projects(found []Project) *model.Array {
	out := model.NewArraySize(len(found))
	for _, project := range found {
		out.Append(phpMap(
			"id", project.ID,
			"dir", project.Dir,
			"sources", texts(project.Sources),
			"jobs", project.Jobs,
			"error", project.Error,
			"checkout", project.Checkout,
			"remote", project.Remote,
			"branch", project.Branch,
			"commit", project.Commit,
			"subject", project.Subject,
			"when", project.When,
		))
	}
	return out
}
