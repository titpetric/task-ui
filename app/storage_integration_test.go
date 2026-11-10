//go:build integration

package app_test

import (
	"context"
	"reflect"
	"slices"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// newDB opens a private in-memory database with the schema applied.
func newDB(t *testing.T) storage.DB {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// nilFields names the exported fields of a container the constructor left
// nil, skipping the ones named. A field added without a line in the
// constructor names itself here.
func nilFields(container any, skip ...string) []string {
	value := reflect.ValueOf(container).Elem()
	var missing []string
	for i := range value.NumField() {
		field := value.Type().Field(i)
		if !field.IsExported() || slices.Contains(skip, field.Name) {
			continue
		}
		// A value field is never nil, and IsNil panics on one: an exported
		// one added here is a finding for the test to report, not a crash.
		switch field.Type.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		default:
			continue
		}
		if value.Field(i).IsNil() {
			missing = append(missing, field.Name)
		}
	}
	return missing
}

// Every storage of an instance is constructed. The wiring reads these
// fields and builds none of its own, so a field with no line in the
// constructor is a nil pointer the first request dereferences.
func TestNewAppStorage(t *testing.T) {
	t.Parallel()
	db := newDB(t)
	store := app.NewAppStorage(db)

	assert.Empty(t, nilFields(store))
	assert.Equal(t, db, store.DB)
}

// Each field holds a driver of its own: two names on one storage would
// read the wrong table.
func TestAppStorage(t *testing.T) {
	t.Parallel()
	value := reflect.ValueOf(app.NewAppStorage(newDB(t))).Elem()

	seen := map[uintptr]string{}
	for i := range value.NumField() {
		field := value.Type().Field(i)
		if field.Name == "DB" {
			continue
		}
		at := value.Field(i).Pointer()
		assert.Equal(t, "", seen[at], field.Name+" shares a driver with "+seen[at])
		seen[at] = field.Name
	}
	assert.Len(t, seen, value.NumField()-1)
}
