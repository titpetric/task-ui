-- search_document: one indexed file, or one job target a folder declares
CREATE TABLE IF NOT EXISTS search_document (
    id CHAR(26) PRIMARY KEY NOT NULL,
    -- doc for a markdown or text file, job for a target of a pipeline, makefile or script.
    kind TEXT NOT NULL DEFAULT 'doc',
    -- path is relative to the root; a job's is <source>#<target> so it stays unique.
    path TEXT NOT NULL UNIQUE,
    file TEXT NOT NULL DEFAULT '',
    -- folder_id is the project folder the text belongs to, as folder.Join builds it.
    folder_id TEXT NOT NULL DEFAULT '',
    workspace TEXT NOT NULL DEFAULT '',
    project TEXT NOT NULL DEFAULT '',
    -- target is the job ID the schedule names, name the target as it is written.
    target TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    dir TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    hash TEXT NOT NULL DEFAULT '',
    size INTEGER NOT NULL DEFAULT 0,
    modified_at DATETIME,
    indexed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_search_document_folder ON search_document(folder_id);
CREATE INDEX IF NOT EXISTS idx_search_document_target ON search_document(kind, target, name);

-- search_section: a heading and the text under it, in the document's order
CREATE TABLE IF NOT EXISTS search_section (
    document_id CHAR(26) NOT NULL,
    ordinal INTEGER NOT NULL,
    level INTEGER NOT NULL DEFAULT 0,
    heading TEXT NOT NULL DEFAULT '',
    breadcrumb TEXT NOT NULL DEFAULT '',
    line INTEGER NOT NULL DEFAULT 1,
    body TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (document_id, ordinal)
);

-- search_term: the inverted index, one row per term a section holds
CREATE TABLE IF NOT EXISTS search_term (
    document_id CHAR(26) NOT NULL,
    ordinal INTEGER NOT NULL,
    term TEXT NOT NULL,
    weight INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (document_id, ordinal, term)
);

CREATE INDEX IF NOT EXISTS idx_search_term_term ON search_term(term);
