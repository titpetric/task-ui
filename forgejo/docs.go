// Package forgejo syncs a Forgejo instance into task-ui: each
// organisation the token's user belongs to is a workspace, the
// repository named like the organisation is its home, every other
// repository a project cloned into its src, and a push hook on each
// keeps the checkouts fresh. The client speaks the Gitea-compatible API
// v1 with a personal access token; Sync is the work of task-ui
// forgejo-sync, the root's @forgejo-sync, and never deletes; Sign and
// Verify are the HMAC-SHA256 of a hook's body, which the webhook
// checks.
package forgejo
