package layout

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/titpetric/task-ui/model"
)

// Plan lists the moves that bring a root laid out as before to the home
// and src layout, for the workspaces named, the default one, _, among
// them while the instance has it; a root laid out already returns none,
// and a directory that exists on both sides of a move is an error.
func Plan(root string, workspaces []string, homes map[string]string) ([]Move, error) {
	var moves []Move
	named := slices.DeleteFunc(slices.Clone(workspaces), func(name string) bool { return name == model.Default })
	for _, workspace := range workspaces {
		dir := WorkspaceDir(root, workspace)
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		home, src := HomeDir(root, workspace), SrcDir(root, workspace)
		homeProject := homes[workspace]
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() || name == Home || name == Src || model.IsReserved(name) || name[0] == '.' {
				continue
			}
			if workspace == model.Default && slices.Contains(named, name) {
				continue
			}
			from := filepath.Join(dir, name)
			if name == homeProject {
				if isDir(home) {
					return nil, fmt.Errorf("%s: both %s and home exist", workspace, name)
				}
				moves = append(moves, Move{From: from, To: home})
				continue
			}
			if isDir(filepath.Join(src, name)) {
				return nil, fmt.Errorf("%s: both %s and src/%s exist", workspace, name, name)
			}
			moves = append(moves, Move{From: from, To: filepath.Join(src, name)})
		}
		if !isDir(home) && !slices.ContainsFunc(moves, func(m Move) bool { return m.To == home }) {
			moves = append(moves, Move{To: home})
		}
		if !isDir(src) {
			moves = append(moves, Move{To: src})
		}
	}
	sort.SliceStable(moves, func(i, j int) bool {
		// Directories first, then the moves into them, in path order.
		if (moves[i].From == "") != (moves[j].From == "") {
			return moves[i].From == ""
		}
		return moves[i].To < moves[j].To
	})
	return moves, nil
}

// Apply makes the directories and moves what Plan listed, in order.
func Apply(moves []Move) error {
	for _, move := range moves {
		if move.From == "" {
			if err := os.MkdirAll(move.To, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(move.To), 0o755); err != nil {
			return err
		}
		if err := os.Rename(move.From, move.To); err != nil {
			return err
		}
	}
	return nil
}

// isDir reports whether path is a directory.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
