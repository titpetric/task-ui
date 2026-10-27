-- A prompt or summary step with no folder asks at the workspaces root
-- with no workspace; before this change an empty folder meant the root,
-- internal, so the stored rows take that meaning.
UPDATE workflow_step SET folder = 'internal' WHERE folder = '' AND kind IN ('prompt', 'summary');
UPDATE workflow_run_step SET folder = 'internal' WHERE folder = '' AND kind IN ('prompt', 'summary');
