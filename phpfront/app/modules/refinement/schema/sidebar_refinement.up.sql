-- The refinement board answers at /refinement, as a page of the front end.
-- The row the plugin registered points at /m/titpetric/refinement, which
-- answers nothing: the module the binary shipped is gone and this page took
-- its place. The update keeps the position and the hidden flag the old row
-- held.
DELETE FROM sidebar WHERE href = '/refinement';

UPDATE sidebar SET kind = 'view', href = '/refinement', title = 'Refinement', icon = 'flow', rule = 'module', updated_at = CURRENT_TIMESTAMP WHERE href = '/m/titpetric/refinement';
