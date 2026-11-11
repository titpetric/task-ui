package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestRun_Select(t *testing.T) {
	t.Parallel()
	run := &model.Run{}
	assert.Equal(t, "SELECT * FROM `run`", run.Select())
	query := run.Select(model.WithColumns([]string{"id"}), model.WithWhere("state = ?"), model.WithOrderBy("id DESC"), model.WithLimit(0, 10))
	assert.Equal(t, "SELECT id FROM `run` WHERE state = ? ORDER BY id DESC LIMIT 0, 10", query)
}

func TestRun_Insert(t *testing.T) {
	t.Parallel()
	query := (&model.Run{}).Insert(model.WithColumns([]string{"id", "folder"}))
	assert.Equal(t, "INSERT INTO `run` (id, folder) VALUES (:id, :folder)", query)
}

func TestRun_Update(t *testing.T) {
	t.Parallel()
	query := (&model.Run{}).Update(model.WithColumns([]string{"state"}), model.WithWhere("id = :id"))
	assert.Contains(t, query, "UPDATE `run` SET state=:state")
	assert.Contains(t, query, "WHERE id = :id")
}

func TestRun_Delete(t *testing.T) {
	t.Parallel()
	query := (&model.Run{}).Delete(model.WithWhere("id = ?"))
	assert.Equal(t, "DELETE FROM `run` WHERE id = ?", query)
}

func TestQueryConfig_Apply(t *testing.T) {
	t.Parallel()
	cfg := (&model.QueryConfig{Table: "a"}).Apply(model.WithTable("b"), model.WithStatement("REPLACE INTO"))
	assert.Equal(t, "b", cfg.Table)
	assert.Equal(t, "REPLACE INTO", cfg.Statement)
	var option model.QueryOption = cfg.WithTable("c").WithColumns([]string{"x"}).WithWhere("w").WithOrderBy("o").WithLimit(1, 2).WithStatement("s")
	assert.Equal(t, "c", option.(*model.QueryConfig).Table)
}

func TestUser_Getters(t *testing.T) {
	t.Parallel()
	user := &model.User{}
	user.SetName("ops")
	user.SetRole(model.RoleRunner)
	assert.Equal(t, "ops", user.GetName())
	assert.Equal(t, model.RoleRunner, user.GetRole())
	assert.Equal(t, "`user`", model.UserTable)
	assert.NotContains(t, model.UserFields, "password")
	assert.Equal(t, []string{"id"}, model.UserPrimaryFields)
}
