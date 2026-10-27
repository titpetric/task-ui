package view

import (
	"sort"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// SecretGroup is a workspace on the secrets pages: the secrets its
// key encrypts, those it shares with its folders first, and the folders
// a secret of its can name as its scope.
type SecretGroup struct {
	// Workspace is the workspace's folder, internal for the default one.
	Workspace string
	Title     string
	// Key is the workspace's encryption key, nil while it has none.
	Key     *keystore.Key
	Secrets []model.Secret
	Folders []string
}

// Shared returns the scope of the secrets the workspace shares with its folders.
func (g SecretGroup) Shared() string {
	return "workspace:" + g.Workspace
}

// SharedSecrets returns the secrets the workspace shares with every folder in it.
func (g SecretGroup) SharedSecrets() []model.Secret {
	var result []model.Secret
	for _, secret := range g.Secrets {
		if secret.Scope == g.Shared() {
			result = append(result, secret)
		}
	}
	return result
}

// Projects returns the folders with secrets of their own, by folder, each with its secrets.
func (g SecretGroup) Projects() []SecretProject {
	index := map[string]int{}
	var result []SecretProject
	for _, secret := range g.Secrets {
		if secret.Scope == g.Shared() {
			continue
		}
		i, ok := index[secret.Scope]
		if !ok {
			i = len(result)
			index[secret.Scope] = i
			result = append(result, SecretProject{Folder: secret.Scope})
		}
		result[i].Secrets = append(result[i].Secrets, secret)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Folder < result[j].Folder })
	return result
}
