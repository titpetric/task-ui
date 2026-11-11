package retention

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// tooling are the tool history databases that are there, each the size
// of its file with its write-ahead log and shared memory.
func (c *Cleaner) tooling() []Item {
	items := []Item{}
	if c.Tooling == nil {
		return items
	}
	for _, path := range c.Tooling() {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		item := Item{ID: path, At: info.ModTime(), Size: info.Size(), Busy: true}
		for _, suffix := range []string{"-wal", "-shm"} {
			if info, err := os.Stat(path + suffix); err == nil {
				item.Size += info.Size()
			}
		}
		items = append(items, item)
	}
	return items
}

// prune has atkins remove the runs older than the rule's age from each
// database, dry passed through, and prints its line for each; without
// databases, without atkins or with an age of never it prints what the
// databases hold.
func (c *Cleaner) prune(ctx context.Context, area Area, rule Rule, items []Item, dry bool, out io.Writer) error {
	says := rule.Keeps()
	if len(items) == 0 || rule.Age == Never || c.Atkins == nil {
		note := ""
		if c.Atkins == nil && len(items) > 0 && rule.Age != Never {
			note = "; atkins not on PATH"
		}
		_, err := fmt.Fprintf(out, "%s (%s): %s%s\n", area.Title, says, area.Holds(usageOf(items)), note)
		return err
	}
	for _, item := range items {
		args := []string{"--prune", "--db", item.ID, "--older-than", rule.Age}
		if dry {
			args = append(args, "--dry-run")
		}
		output, err := c.Atkins(ctx, args...)
		if err != nil {
			return fmt.Errorf("%s: %s: %w: %s", area.Title, item.ID, err, strings.TrimSpace(string(output)))
		}
		if _, err := fmt.Fprintf(out, "%s (%s): %s\n", area.Title, says, strings.TrimSpace(string(output))); err != nil {
			return err
		}
	}
	return nil
}
