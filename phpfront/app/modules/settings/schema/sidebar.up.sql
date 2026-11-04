-- sidebar: One item of the sidebar's top block, in the order it lists them
CREATE TABLE IF NOT EXISTS sidebar (
    id CHAR(26) PRIMARY KEY NOT NULL,
    -- kind is 'view' for an item of the instance, 'module' for one a
    -- modinfo.php registered. A view's row cannot be deleted, a module's
    -- goes when its directory does
    kind TEXT NOT NULL DEFAULT 'view',
    -- href is the link, and the unique key: two rows cannot name one page
    href TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    -- icon names a symbol of the sprite in _layout.tpl, without the i- prefix
    icon TEXT NOT NULL DEFAULT '',
    -- position orders the block, lowest first
    position INTEGER NOT NULL DEFAULT 0,
    -- hidden leaves the item out of the sidebar. The page behind it still
    -- answers, and the rule below still decides who may open it
    hidden BOOLEAN NOT NULL DEFAULT 0,
    -- rule is the permission the row needs, a path as a template's .Can
    -- spells one, 'module/view'. '' is a row everyone signed in sees, and
    -- 'settings' is the one value that is not a permission: the settings
    -- open for whoever holds the view rule of any of their sections, and
    -- no single rule says that. A user without the rule never sees the row
    rule TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_sidebar_href ON sidebar(href);
CREATE INDEX IF NOT EXISTS idx_sidebar_position ON sidebar(position);
