-- module: A plugin written in PHP, as its modinfo.php registered it
CREATE TABLE IF NOT EXISTS module (
    id CHAR(26) PRIMARY KEY NOT NULL,
    -- name is <vendor>/<plugin>, which decides the link, /m/<name>
    name TEXT NOT NULL,
    -- vendor and plugin are the two halves of the name, kept apart so a
    -- page groups by vendor without splitting the string again
    vendor TEXT NOT NULL DEFAULT '',
    plugin TEXT NOT NULL DEFAULT '',
    -- root is the directory holding the modinfo.php, absolute. A module
    -- found under another root on the next start takes the new one
    root TEXT NOT NULL DEFAULT '',
    -- title, description and icon are what the sidebar row and the page
    -- heading show; icon names a symbol of the sprite without the i- prefix
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    -- version is whatever the module calls its own
    version TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_module_name ON module(name);
