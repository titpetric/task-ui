package phpfront

import (
	"github.com/titpetric/phpscript/model"
)

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
		"session_max", found.SessionMax,
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
		"debug_span_min_duration", found.DebugSpanMinDuration,
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
			"fixed", field.Fixed,
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
