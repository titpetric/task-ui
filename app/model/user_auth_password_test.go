package model_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestUserAuthPassword_Usable(t *testing.T) {
	t.Parallel()
	used := time.Now()
	assert.True(t, (&model.UserAuthPassword{Hash: "$2h", Kind: model.AuthPassword}).Usable())
	assert.True(t, (&model.UserAuthPassword{Hash: "$2h", Kind: model.AuthBackup}).Usable())
	assert.False(t, (&model.UserAuthPassword{Hash: "$2h", Kind: model.AuthBackup, UsedAt: &used}).Usable())
	assert.False(t, (&model.UserAuthPassword{Kind: model.AuthPassword}).Usable())
	assert.False(t, (&model.UserAuthPassword{Hash: "$2h", Kind: "other"}).Usable())
	var none *model.UserAuthPassword
	assert.False(t, none.Usable())
}
