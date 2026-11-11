// Package layout maps folder IDs to directories under the root.
//
// A workspace is a directory under the root, the root itself for the
// default workspace _, with two directories in it: home, the workspace's
// home, a checkout of its home repository or an empty directory, and
// src, which holds its projects side by side. Folder IDs leave the two
// out: the project dockerfiles of the workspace malta lives at
// <root>/malta/src/dockerfiles and its ID is malta/dockerfiles; a project
// of _ lives at <root>/src/<project> and its ID is the project's name.
// The home is a folder of its own, malta/home, home for _, and no
// project.
//
// Every place that turns an ID into a directory, or a directory into an
// ID, calls this package, so the shape of the tree lives in one place:
// discovery, the pages, the API, the runner and the layout migration.
package layout
