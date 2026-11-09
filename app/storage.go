package app

import (
	"github.com/titpetric/task-ui/storage"
)

// AppStorage holds one storage driver per table over the open database.
type AppStorage struct {
	DB storage.DB

	Users     *storage.UserStorage
	Emails    *storage.UserEmailStorage
	Passwords *storage.UserAuthPasswordStorage
	Sessions  *storage.UserSessionStorage
	Tokens    *storage.UserAuthTokenStorage

	Runs         *storage.RunStorage
	RunSummaries *storage.RunSummaryStorage
	Secrets      *storage.SecretStorage
	SecretReads  *storage.SecretReadStorage
	Projects     *storage.ProjectStorage
	Workspaces   *storage.WorkspaceStorage
	Artifacts    *storage.ArtifactStorage
	Schedules    *storage.ScheduleStorage
	Workflows    *storage.WorkflowStorage
	WorkflowRuns *storage.WorkflowRunStorage
	FolderRefs   *storage.FolderRefStorage
	Tools        *storage.ToolStorage
	Modules      *storage.ModuleStorage
	Sidebar      *storage.SidebarStorage
	Notes        *storage.NoteStorage
	Pins         *storage.PinStorage
	Hosts        *storage.HostStorage
	Settings     *storage.SettingStorage
	EnvFiles     *storage.EnvFileStorage

	KeyInfo     *storage.KeyInfoStorage
	KeyPolicies *storage.SSHKeyPolicyStorage
	KeyGrants   *storage.WorkspaceKeyGrantStorage
}

// NewAppStorage returns every storage over db, each constructed once.
func NewAppStorage(db storage.DB) *AppStorage {
	return &AppStorage{
		DB: db,

		Users:     storage.NewUserStorage(db),
		Emails:    storage.NewUserEmailStorage(db),
		Passwords: storage.NewUserAuthPasswordStorage(db),
		Sessions:  storage.NewUserSessionStorage(db),
		Tokens:    storage.NewUserAuthTokenStorage(db),

		Runs:         storage.NewRunStorage(db),
		RunSummaries: storage.NewRunSummaryStorage(db),
		Secrets:      storage.NewSecretStorage(db),
		SecretReads:  storage.NewSecretReadStorage(db),
		Projects:     storage.NewProjectStorage(db),
		Workspaces:   storage.NewWorkspaceStorage(db),
		Artifacts:    storage.NewArtifactStorage(db),
		Schedules:    storage.NewScheduleStorage(db),
		Workflows:    storage.NewWorkflowStorage(db),
		WorkflowRuns: storage.NewWorkflowRunStorage(db),
		FolderRefs:   storage.NewFolderRefStorage(db),
		Tools:        storage.NewToolStorage(db),
		Modules:      storage.NewModuleStorage(db),
		Sidebar:      storage.NewSidebarStorage(db),
		Notes:        storage.NewNoteStorage(db),
		Pins:         storage.NewPinStorage(db),
		Hosts:        storage.NewHostStorage(db),
		Settings:     storage.NewSettingStorage(db),
		EnvFiles:     storage.NewEnvFileStorage(db),

		KeyInfo:     storage.NewKeyInfoStorage(db),
		KeyPolicies: storage.NewSSHKeyPolicyStorage(db),
		KeyGrants:   storage.NewWorkspaceKeyGrantStorage(db),
	}
}
