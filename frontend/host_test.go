package phpfront_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/internal/assert"
)

// What the host answers crosses into the runtime as JSON-shaped values, so the
// tags are the names the PHP indexes by. A renamed field is a page that reads
// an empty string.
func TestHost_Shapes(t *testing.T) {
	t.Parallel()

	folder, err := json.Marshal(phpfront.Folder{ID: "infra", Title: "Infra", Dir: "/src/infra", Allow: []string{"ops"}, Jobs: []phpfront.Job{{ID: "build", Name: "build"}}})
	assert.NoError(t, err)
	assert.Equal(t, `{"id":"infra","title":"Infra","dir":"/src/infra","allow":["ops"],"sources":null,"error":"","errors":null,"jobs":[{"found":false,"id":"build","name":"build","dir":"","shell":false,"host":false,"allow":null,"folder_allow":null,"kind":"","source":"","source_lines":0,"description":"","summary":"","command":null,"steps":null,"aliases":null,"depends":null,"interactive":false,"takes_args":false,"args_command":"","key":"","confirm":""}]}`, string(folder))

	job, err := json.Marshal(phpfront.Job{Found: true, ID: "@shell", Name: "shell", Shell: true, Allow: []string{"admin"}, FolderAllow: []string{"ops"}})
	assert.NoError(t, err)
	assert.Equal(t, `{"found":true,"id":"@shell","name":"shell","dir":"","shell":true,"host":false,"allow":["admin"],"folder_allow":["ops"],"kind":"","source":"","source_lines":0,"description":"","summary":"","command":null,"steps":null,"aliases":null,"depends":null,"interactive":false,"takes_args":false,"args_command":"","key":"","confirm":""}`, string(job))

	running, err := json.Marshal(phpfront.Running{Key: "infra/build", ID: "01hz", State: "running"})
	assert.NoError(t, err)
	assert.Equal(t, `{"key":"infra/build","id":"01hz","state":"running"}`, string(running))

	rules, err := json.Marshal(phpfront.Rules{
		Rules:       []phpfront.Rule{{Key: "view", Title: "Allow View"}},
		Sections:    []phpfront.Rule{{Key: "note", Title: "Notes"}},
		Permissions: []phpfront.Rule{{Key: "run", Title: "Run Jobs"}},
		Paths:       []string{"note/all", "note/view"},
		Checks:      []string{"note/view"},
		Extras:      []phpfront.Rule{{Key: "secret/read", Title: "Read Secrets"}},
	})
	assert.NoError(t, err)
	assert.Contains(t, string(rules), `"paths":["note/all","note/view"]`)
	assert.Contains(t, string(rules), `"checks":["note/view"]`)
	// An extra is keyed by the path a check names, not by the rule alone.
	assert.Contains(t, string(rules), `"extras":[{"key":"secret/read","title":"Read Secrets"}]`)
}

// The lists are ordered, because Paths is read in file order and the page that
// draws the permission matrix follows it.
func TestRules_Order(t *testing.T) {
	t.Parallel()

	rules := phpfront.Rules{Paths: []string{"workspace/all", "workspace/view", "run"}}
	assert.Equal(t, "workspace/all", rules.Paths[0])
	assert.Equal(t, "run", rules.Paths[len(rules.Paths)-1])
}
