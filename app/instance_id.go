package app

import (
	"context"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// InstanceIDSetting is the name the instance's own identifier lives under in
// the setting table, beside the retention rules and the telemetry options.
const InstanceIDSetting = "instance.id"

// instanceID returns the instance's identifier, making one on the first start
// and keeping it in the setting table. Every start after reads the one it
// made, so the page names the same instance across a restart and a deploy.
//
// A value that is not one of those identifiers is replaced rather than kept:
// the setting holds what this code wrote, and a hand-edited row does not name
// the instance.
func instanceID(ctx context.Context, settings *storage.SettingStorage) (string, error) {
	if settings == nil {
		return model.NewID(), nil
	}
	values, err := settings.Values(ctx)
	if err != nil {
		return "", err
	}
	if id := values[InstanceIDSetting]; model.ValidID(id) {
		return id, nil
	}
	id := model.NewID()
	if err := settings.Set(ctx, map[string]string{InstanceIDSetting: id}); err != nil {
		return "", err
	}
	return id, nil
}
