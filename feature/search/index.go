package search

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/titpetric/task-ui/storage"
)

// Store is the part of the search storage the index writes through.
type Store interface {
	Documents(ctx context.Context, folderID string) ([]storage.SearchDocument, error)
	DocumentFolders(ctx context.Context) ([]string, error)
	JobDocuments(ctx context.Context) ([]storage.SearchDocument, error)
	ReplaceDocument(ctx context.Context, doc *storage.SearchDocument, sections []storage.SearchSection, terms []storage.SearchTerm) error
	DeleteDocument(ctx context.Context, id string) error
	DeleteFolder(ctx context.Context, folderID string) error
}

// Indexer writes documentation and job targets into the search storage.
type Indexer struct {
	Store   Store
	Root    string
	Sources func(ctx context.Context) ([]Source, error)
	Targets func(ctx context.Context) ([]Target, error)
	After   func(ctx context.Context) error
	Logger  *slog.Logger

	mu      sync.Mutex
	running bool
	done    chan struct{}
}

// NewIndexer returns an indexer over a store; root is what its paths are relative to.
func NewIndexer(store Store, root string) *Indexer {
	return &Indexer{Store: store, Root: root}
}

// fileLimit bounds a document, so a generated file does not become the index.
const fileLimit = 1 << 20

// skippedDirs are the trees an index does not read: build output and dependencies.
var skippedDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"ext":          true,
	"templates_c":  true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"__pycache__":  true,
	"venv":         true,
	".venv":        true,
}

// Run indexes every source and job target and sweeps what left the tree.
func (ix *Indexer) Run(ctx context.Context) (Stats, error) {
	var stats Stats
	if ix == nil || ix.Store == nil {
		return stats, nil
	}
	sources, err := ix.sources(ctx)
	if err != nil {
		return stats, err
	}
	seen := map[string]bool{}
	for _, source := range sources {
		seen[source.FolderID] = true
		if err := ix.source(ctx, source, &stats); err != nil {
			return stats, err
		}
	}
	folders, err := ix.Store.DocumentFolders(ctx)
	if err != nil {
		return stats, err
	}
	for _, folder := range folders {
		if seen[folder] {
			continue
		}
		if err := ix.Store.DeleteFolder(ctx, folder); err != nil {
			return stats, err
		}
		stats.Deleted++
	}
	targets, err := ix.targets(ctx)
	if err != nil {
		return stats, err
	}
	if err := ix.jobs(ctx, targets, &stats); err != nil {
		return stats, err
	}
	if ix.After != nil {
		if err := ix.After(ctx); err != nil {
			return stats, fmt.Errorf("after the index: %w", err)
		}
	}
	return stats, nil
}

// Trigger starts a run in the background, unless one is running.
func (ix *Indexer) Trigger(ctx context.Context) error {
	if ix == nil || ix.Store == nil {
		return nil
	}
	ix.mu.Lock()
	if ix.running {
		ix.mu.Unlock()
		return nil
	}
	ix.running = true
	ix.done = make(chan struct{})
	done := ix.done
	ix.mu.Unlock()

	go func() {
		defer func() {
			ix.mu.Lock()
			ix.running = false
			ix.mu.Unlock()
			close(done)
		}()
		stats, err := ix.Run(context.WithoutCancel(ctx))
		if err != nil {
			ix.logger().Error("the search index failed", "err", err)
			return
		}
		ix.logger().Info("search indexed", "documents", stats.Documents, "indexed", stats.Indexed, "skipped", stats.Skipped, "jobs", stats.Jobs, "deleted", stats.Deleted+stats.JobsDeleted)
	}()
	return nil
}

// Wait blocks until the background run Trigger started has ended.
func (ix *Indexer) Wait() {
	ix.mu.Lock()
	done := ix.done
	ix.mu.Unlock()
	if done != nil {
		<-done
	}
}

// source indexes one project's markdown and text files, and deletes the ones that left it.
func (ix *Indexer) source(ctx context.Context, source Source, stats *Stats) error {
	docs, err := ix.Store.Documents(ctx, source.FolderID)
	if err != nil {
		return err
	}
	known := make(map[string]storage.SearchDocument, len(docs))
	for _, doc := range docs {
		known[doc.Path] = doc
	}
	seen := map[string]bool{}
	if source.Dir != "" {
		_ = filepath.WalkDir(source.Dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if entry.IsDir() {
				if path != source.Dir && skipped(entry.Name()) {
					return fs.SkipDir
				}
				return nil
			}
			if !entry.Type().IsRegular() || !indexed(entry.Name()) {
				return nil
			}
			info, err := entry.Info()
			if err != nil || info.Size() > fileLimit {
				return nil
			}
			rel, err := filepath.Rel(ix.Root, path)
			if err != nil || strings.HasPrefix(rel, "..") {
				return nil
			}
			rel = filepath.ToSlash(rel)
			seen[rel] = true
			stats.Documents++
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			hash := md5hex(data)
			if old, ok := known[rel]; ok && old.Hash == hash {
				stats.Skipped++
				return nil
			}
			title, sections := split(entry.Name(), data)
			if len(sections) == 0 {
				if old, ok := known[rel]; ok {
					if err := ix.Store.DeleteDocument(ctx, old.ID); err != nil {
						return err
					}
					stats.Deleted++
				}
				return nil
			}
			doc := &storage.SearchDocument{
				Kind: storage.SearchDocKind, Path: rel, FolderID: source.FolderID,
				Workspace: source.Workspace, Project: source.Project,
				Title: title, Hash: hash, Size: info.Size(),
			}
			if modified := info.ModTime(); !modified.IsZero() {
				doc.ModifiedAt = &modified
			}
			if err := ix.Store.ReplaceDocument(ctx, doc, sectionRows(sections), termRows(title, rel, sections)); err != nil {
				return err
			}
			stats.Indexed++
			return nil
		})
	}
	for path, doc := range known {
		if seen[path] {
			continue
		}
		if err := ix.Store.DeleteDocument(ctx, doc.ID); err != nil {
			return err
		}
		stats.Deleted++
	}
	return nil
}

// jobs indexes the job targets, and deletes the ones that left the sources.
func (ix *Indexer) jobs(ctx context.Context, targets []Target, stats *Stats) error {
	known := map[string]storage.SearchDocument{}
	docs, err := ix.Store.JobDocuments(ctx)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		known[doc.Path] = doc
	}
	seen := map[string]bool{}
	for _, target := range targets {
		if target.ID == "" || target.Name == "" || target.SourceFile == "" {
			continue
		}
		path := target.SourceFile + "#" + target.ID
		seen[path] = true
		title := target.Name
		if target.Description != "" {
			title += " - " + target.Description
		}
		body := targetBody(target)
		hash := md5hex([]byte(target.FolderID + "\x00" + title + "\x00" + body))
		if old, ok := known[path]; ok && old.Hash == hash {
			stats.Skipped++
			continue
		}
		doc := &storage.SearchDocument{
			Kind: storage.SearchJobKind, Path: path, File: target.SourceFile, FolderID: target.FolderID,
			Workspace: target.Workspace, Project: target.Project, Target: target.ID, Name: target.Name,
			Dir: target.Dir, Title: title, Hash: hash, Size: int64(len(body)),
		}
		sections := []Section{{Ordinal: 0, Level: 1, Heading: target.Name, Breadcrumb: target.Name, Line: 1, Body: body}}
		terms := termRows(title, path, sections)
		terms = jobTerms(terms, target)
		if err := ix.Store.ReplaceDocument(ctx, doc, sectionRows(sections), terms); err != nil {
			return err
		}
		stats.Jobs++
	}
	for path, doc := range known {
		if seen[path] {
			continue
		}
		if err := ix.Store.DeleteDocument(ctx, doc.ID); err != nil {
			return err
		}
		stats.JobsDeleted++
	}
	return nil
}

// split reads a document as markdown or as one plain text section.
func split(name string, data []byte) (string, []Section) {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".markdown":
		return Split(data, name)
	default:
		return SplitText(data, name)
	}
}

// indexed reports a file the index reads.
func indexed(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".markdown", ".txt":
		return true
	}
	return false
}

// skipped reports a directory an index does not descend into.
func skipped(name string) bool {
	return strings.HasPrefix(name, ".") || skippedDirs[name]
}

// sectionRows turns parsed sections into storage rows.
func sectionRows(sections []Section) []storage.SearchSection {
	rows := make([]storage.SearchSection, 0, len(sections))
	for _, section := range sections {
		rows = append(rows, storage.SearchSection{
			Ordinal: int64(section.Ordinal), Level: int64(section.Level), Heading: section.Heading,
			Breadcrumb: section.Breadcrumb, Line: int64(section.Line), Body: section.Body,
		})
	}
	return rows
}

// termRows turns a document's sections into inverted-index rows.
func termRows(title, path string, sections []Section) []storage.SearchTerm {
	rows := make([]storage.SearchTerm, 0, len(sections)*8)
	for _, section := range sections {
		terms := Terms(title, section.Heading, section.Body)
		if section.Ordinal == 0 {
			for _, token := range Tokenize(path) {
				terms[token] += 4
			}
		}
		keys := make([]string, 0, len(terms))
		for term := range terms {
			keys = append(keys, term)
		}
		sort.Strings(keys)
		for _, term := range keys {
			rows = append(rows, storage.SearchTerm{Ordinal: int64(section.Ordinal), Term: term, Weight: int64(terms[term])})
		}
	}
	return rows
}

// jobTerms adds the target's own words to its section's terms.
func jobTerms(rows []storage.SearchTerm, target Target) []storage.SearchTerm {
	extra := map[string]int{}
	for _, field := range []struct {
		text   string
		weight int
	}{
		{target.Name, 4}, {target.Description, 3}, {target.Summary, 2}, {strings.Join(target.Steps, " "), 1}, {strings.Join(target.Aliases, " "), 2},
	} {
		for _, token := range Tokenize(field.text) {
			extra[token] += field.weight
		}
	}
	for term, weight := range extra {
		rows = append(rows, storage.SearchTerm{Ordinal: 0, Term: term, Weight: int64(weight)})
	}
	return rows
}

// targetBody is the text search_read returns for a job target.
func targetBody(target Target) string {
	lines := []string{}
	if target.Description != "" {
		lines = append(lines, target.Description)
	}
	if target.Summary != "" && target.Summary != target.Description {
		lines = append(lines, target.Summary)
	}
	lines = append(lines, "Folder: "+target.FolderTitle+" ("+target.FolderID+")", "Job: "+target.ID, "Kind: "+target.Kind, "Source: "+target.SourceFile)
	if len(target.Command) > 0 {
		lines = append(lines, "Command: "+strings.Join(target.Command, " "))
	}
	if len(target.Steps) > 0 {
		lines = append(lines, "Steps:", "- "+strings.Join(target.Steps, "\n- "))
	}
	if len(target.Aliases) > 0 {
		lines = append(lines, "Aliases: "+strings.Join(target.Aliases, ", "))
	}
	return strings.Join(lines, "\n")
}

// md5hex is the content hash an unchanged file is skipped by.
func md5hex(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// sources reads the configured sources, or none.
func (ix *Indexer) sources(ctx context.Context) ([]Source, error) {
	if ix.Sources == nil {
		return nil, nil
	}
	return ix.Sources(ctx)
}

// targets reads the configured job targets, or none.
func (ix *Indexer) targets(ctx context.Context) ([]Target, error) {
	if ix.Targets == nil {
		return nil, nil
	}
	return ix.Targets(ctx)
}

// logger returns the index's logger, the process one by default.
func (ix *Indexer) logger() *slog.Logger {
	if ix.Logger != nil {
		return ix.Logger
	}
	return slog.Default()
}
