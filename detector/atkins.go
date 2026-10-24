package detector

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/titpetric/task-ui/skills"
)

// AtkinsKind names the atkins convention.
const AtkinsKind = "atkins"

// atkinsNames are the file names atkins discovers, in its order. A
// folder holding one has an atkins pipeline; which of them atkins reads
// is atkins' own answer, and the listing reports it.
var atkinsNames = []string{".atkins.yml", ".atkins.yaml", "atkins.yml", "atkins.yaml"}

// prober is the one atkins the detector asks, a package value and no
// field: Detectors is a list of values with no wiring, and a probe caches
// on the files it read, so two detectors share one cache. UseProber sets
// it to the prober the skill chips use, so one atkins answers both; a test
// points it at an atkins of its own.
var prober = &skills.Prober{}

// probing guards the swap a test makes when it points the detector at a
// different atkins.
var probing sync.Mutex

// UseProber points the atkins detector at a prober and returns a function that puts the previous one back.
func UseProber(p *skills.Prober) func() {
	probing.Lock()
	defer probing.Unlock()

	previous := prober
	prober = p
	return func() {
		probing.Lock()
		defer probing.Unlock()
		prober = previous
	}
}

// atkinsDetector reads a folder's atkins pipeline by asking atkins what
// runs there.
//
// atkins publishes the whole job in `atkins -l --json`: its description,
// its aliases, what it depends on, its steps, the files it reads and
// whether it hides itself. Reading the YAML here instead would
// be a second reader of a schema atkins owns, and a field added there
// would go missing here with nothing to report it.
type atkinsDetector struct{}

func (atkinsDetector) Kind() string { return AtkinsKind }

// Detect returns the folder's pipeline, or nil when it has none.
//
// A folder with a pipeline file atkins cannot read still answers, with
// Error set, so the folder lists what is wrong.
func (d atkinsDetector) Detect(dir string) *Source {
	name := pipelineFile(dir)
	if name == "" {
		return nil
	}
	return d.detectFile(dir, "")
}

// detectFile reads the folder's own pipeline, or the one file named, for
// a configuration that points at a file no convention claims.
func (d atkinsDetector) detectFile(dir, file string) *Source {
	name := file
	if name == "" {
		name = pipelineFile(dir)
	}

	probing.Lock()
	listing := prober.List(dir, file)
	probing.Unlock()

	return fromListing(dir, name, listing)
}

// fromListing reads atkins' answer for the pipeline name in dir into a
// source. A pipeline atkins' linter refuses carries the findings in
// Errors and lists the jobs atkins still gave, if any.
func fromListing(dir, name string, listing skills.Listing) *Source {
	source := &Source{Kind: AtkinsKind, File: name, Command: defaultCommand(AtkinsKind, name)}
	if listing.Broken() {
		source.Error = name + ": " + listing.Error
		if listing.Error == "" {
			source.Error = name + ": atkins found errors in this pipeline"
		}
		source.Errors = findings(dir, name, listing)
	}
	section := pipelineSection(listing.Sections)
	if section == nil {
		if source.Error == "" {
			source.Error = name + ": atkins could not read this pipeline"
		}
		return source
	}

	source.Name = section.Desc
	source.Description = strings.TrimSpace(section.Help)
	source.Includes = appendUnique(nil, section.Include...)

	for _, cmd := range section.Cmds {
		if target := atkinsTarget(cmd); target != nil {
			source.Targets = append(source.Targets, target)
		}
	}

	return source
}

// findings takes atkins' lint findings into the source, each file named
// relative to the folder: atkins names it absolute, a path of the host
// the page has no use for. A finding without a file is the pipeline's,
// and an error without findings is one.
func findings(dir, name string, listing skills.Listing) []skills.Finding {
	lint := listing.Lint
	if len(lint) == 0 {
		lint = []skills.Finding{{Detail: listing.Error}}
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	result := make([]skills.Finding, 0, len(lint))
	for _, finding := range lint {
		switch rel, err := filepath.Rel(dir, finding.Path); {
		case finding.Path == "":
			finding.Path = name
		case filepath.IsAbs(finding.Path) && err == nil:
			finding.Path = filepath.ToSlash(rel)
		}
		result = append(result, finding)
	}
	return result
}

// pipelineFile returns the pipeline file name a folder holds, empty
// when it holds none.
func pipelineFile(dir string) string {
	for _, name := range atkinsNames {
		if isFile(dir, name) {
			return name
		}
	}
	return ""
}

// pipelineSection returns the section of the folder's own pipeline. The
// skills sections beside it belong to the folders they came from, and a
// chip shows them; they are not this folder's jobs.
func pipelineSection(sections []skills.Section) *skills.Section {
	for _, section := range sections {
		if section.Kind == skills.KindPipeline {
			return &section
		}
	}
	return nil
}

// atkinsTarget turns one listed job into a target, and returns nil for a
// job a folder does not list: one the file hides with show: false, and
// one with nothing to say about itself, which is a job written for
// another job to call.
//
// A job without a desc takes its first step as its description, the way
// `atkins -l` shows one.
func atkinsTarget(cmd skills.Command) *Target {
	desc := strings.TrimSpace(cmd.Desc)
	if desc == "" && len(cmd.Steps) > 0 {
		desc = cmd.Steps[0]
	}
	if cmd.Hidden || desc == "" {
		return nil
	}

	return &Target{
		Name:        cmd.Name,
		Description: desc,
		Aliases:     cmd.Aliases,
		Interactive: cmd.Interactive,
		TakesArgs:   cmd.Args,
		Exec:        cmd.Exec,
		Steps:       cmd.Steps,
		DependsOn:   cmd.DependsOn,
	}
}
