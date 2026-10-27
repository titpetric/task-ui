-- The default workspace, the root, is called Internal, not Default.
UPDATE workspace SET title = 'Internal' WHERE folder = 'internal' AND title = 'Default';
