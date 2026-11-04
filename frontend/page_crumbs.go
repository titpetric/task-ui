package frontend

import (
	"strings"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// crumbName is a folder's name in a crumb: Internal for the default
// workspace, internal, as the Keys and Workspaces pages call it.
func crumbName(folder string) string {
	if folder == model.Default {
		return "Internal"
	}
	return folder
}

// crumbsOf walks a nested folder ID into a breadcrumb, one step per
// segment, with a last step that has no link when given. A folder in no
// named workspace starts at the default workspace, Internal.
func crumbsOf(id, last string, workspaces map[string]int) []view.Crumb {
	var crumbs []view.Crumb
	if strings.HasPrefix(view.FolderHref(id, workspaces), "/f/"+model.Default+"/") {
		crumbs = append(crumbs, view.Crumb{Text: crumbName(model.Default), Href: view.FolderHref(model.Default, workspaces)})
	}
	path := ""
	for _, segment := range strings.Split(id, "/") {
		if path != "" {
			path += "/"
		}
		path += segment
		crumbs = append(crumbs, view.Crumb{Text: crumbName(segment), Href: view.FolderHref(path, workspaces)})
	}
	if last != "" {
		crumbs = append(crumbs, view.Crumb{Text: last})
	} else if len(crumbs) > 0 {
		crumbs[len(crumbs)-1].Href = ""
	}
	return crumbs
}
