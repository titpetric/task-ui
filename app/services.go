package app

import (
	"context"
	"os"
	"time"

	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/feature/artifacts"
	"github.com/titpetric/task-ui/feature/inventory/hostinfo"
	"github.com/titpetric/task-ui/feature/inventory/hostpath"
	"github.com/titpetric/task-ui/feature/orb/recording"
	"github.com/titpetric/task-ui/feature/orb/stt"
	"github.com/titpetric/task-ui/feature/secret"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/integration/beszel"
	"github.com/titpetric/task-ui/integration/git"
	"github.com/titpetric/task-ui/integration/spotify"
	"github.com/titpetric/task-ui/integration/ssh/diskusage"
	"github.com/titpetric/task-ui/integration/ssh/keystore"
	"github.com/titpetric/task-ui/integration/ssh/sshprobe"
	"github.com/titpetric/task-ui/internal/envfile"
	"github.com/titpetric/task-ui/internal/retention"
	"github.com/titpetric/task-ui/internal/tooling"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/storage"
)

// AppServices holds the services an instance shares, built over its storage.
type AppServices struct {
	Keystore     *keystore.Local
	Git          *git.Git
	Secrets      *secret.Store
	Env          *envfile.Store
	Integrations *integration.Store
	SSH          *sshprobe.Prober
	Hosts        *beszel.Inventory
	Recordings   *recording.Store
	Artifacts    *artifactstore.Store
	Disk         *diskusage.Meter
	Measured     *diskusage.Inventory
	Facts        *hostinfo.Reader
	Tooling      *tooling.Prober
	Events       *event.Bus
	Telemetry    *Telemetry
	Cleaner      *retention.Cleaner
	Spotify      *spotify.Service
	STT          *stt.Service
	Instance     *phpfront.Instance

	// mounts and hostData are the container's bind mounts, read once for the runner and the Instance page.
	mounts   hostpath.Mounts
	hostData string
}

// NewAppServices returns the services over s, failing on stored telemetry options that build no tracer.
func NewAppServices(cfg *Config, s *AppStorage) (*AppServices, error) {
	keys := newKeystore(cfg, s)
	secrets := newSecrets(keys, s)
	integrations := &integration.Store{Secrets: secrets}
	prober := &sshprobe.Prober{
		Users: []string{"debian", "root"}, Keys: keys,
		KnownHosts: keys.KnownHosts(), Timeout: 4 * time.Second, TTL: 10 * time.Minute,
		Terminal: true,
	}
	artifacts := newArtifactStore(cfg)
	telemetry, err := NewTelemetry(context.Background(), s.Settings, "task-ui")
	if err != nil {
		return nil, err
	}
	home := os.Getenv("HOME")
	mounts, hostData := hostPaths(cfg)
	compose := cfg.ComposeProject
	if compose == "" {
		compose = ComposeProject("docker")
	}
	return &AppServices{
		Keystore:     keys,
		Git:          &git.Git{Keys: keys},
		Secrets:      secrets,
		Env:          envfile.New(s.EnvFiles, secrets),
		Integrations: integrations,
		SSH:          prober,
		Hosts: &beszel.Inventory{TTL: time.Minute, Settings: func(ctx context.Context) (string, string, error) {
			values, err := integrations.Values(ctx, integration.Beszel)
			return values["url"], values["token"], err
		}},
		Recordings: &recording.Store{Root: cfg.RecordingsDir()},
		Artifacts:  artifacts,
		// The meter walks the artifact tree, the reader signs in to a host and
		// the tool prober runs atkins: both front ends hold the same one, so the
		// work is done once however a reader came to the page, and the two show
		// the same numbers.
		Disk:      &diskusage.Meter{TTL: 10 * time.Minute},
		Measured:  &diskusage.Inventory{TTL: 30 * time.Second, Wait: 2 * time.Second, Items: storageItems(cfg, s.Workspaces, home)},
		Facts:     &hostinfo.Reader{Run: prober.Output, Timeout: 20 * time.Second, TTL: 5 * time.Minute},
		Tooling:   &tooling.Prober{Timeout: 5 * time.Second, TTL: time.Minute},
		Events:    event.NewBus(),
		Telemetry: telemetry,
		Cleaner:   newCleaner(s, cfg, artifacts, home),
		Spotify: &spotify.Service{
			Settings: func(ctx context.Context) (map[string]string, error) {
				return integrations.Values(ctx, integration.Spotify)
			},
			Tokens:  spotify.SecretTokens{Secrets: secrets},
			Client:  &spotify.Client{},
			OAuth:   &spotify.OAuth{},
			States:  &spotify.States{},
			BaseURL: cfg.URL,
		},
		STT: stt.New(func(ctx context.Context) (map[string]string, error) {
			return integrations.Values(ctx, integration.STT)
		}),
		// What the server runs with, as the Instance page draws it. The meter
		// behind it is one, so a page drawn twice walks the disk once.
		Instance: &phpfront.Instance{
			URL: cfg.URL, Listen: cfg.Listen, Root: cfg.Root, DataDir: cfg.DataDir, SSHDir: cfg.SSHDir, Htpasswd: cfg.Htpasswd, HostData: hostDataShown(cfg.DataDir, hostData),
			Discover: cfg.Discover, AutoLogin: cfg.AutoLogin, TimeZone: timeZone(cfg.Clock()()),
			SessionTTL: cfg.SessionTTL.String(), SessionMax: cfg.SessionMax.String(), Keep: cfg.Keep.String(), Grace: cfg.Grace.String(),
			ArtifactHistory: cfg.ArtifactHistory, ComposeProject: compose,
		},
		mounts:   mounts,
		hostData: hostData,
	}, nil
}

// newKeystore returns the key folder, cfg.SSHDir, over the storage.
func newKeystore(cfg *Config, s *AppStorage) *keystore.Local {
	return &keystore.Local{Dir: cfg.SSHDir, Rows: s.KeyInfo, Settings: s.Settings}
}

// newSecrets returns the secret store on keys: every secret on the
// instance's encryption key, the workspaces' and the services'. No SSH key
// encrypts or decrypts with it.
func newSecrets(keys *keystore.Local, s *AppStorage) *secret.Store {
	store := secret.New(s.Secrets, secret.NewVault(keys.Vault(storage.ScopeInstance)))
	store.Workspaces = &secret.Workspaces{Rows: s.Workspaces}
	return store
}

// newArtifactStore returns the artifact tree the runs commit into.
func newArtifactStore(cfg *Config) *artifactstore.Store {
	return &artifactstore.Store{Dir: cfg.ArtifactTree(), Keep: cfg.ArtifactHistory}
}
