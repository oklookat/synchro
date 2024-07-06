PRAGMA foreign_keys = ON;

------ REMOTE
CREATE TABLE IF NOT EXISTS remote (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS account (
    id TEXT PRIMARY KEY,
    remote_name INTEGER NOT NULL REFERENCES remote (name) ON DELETE CASCADE,
    alias TEXT NOT NULL DEFAULT 'Without alias',
    auth TEXT NOT NULL,
    added_at INTEGER NOT NULL
);

------ ARTISTS
CREATE TABLE IF NOT EXISTS artist (id TEXT PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_artist (
    id TEXT PRIMARY KEY,
    entity_id INTEGER NOT NULL REFERENCES artist (id) ON DELETE CASCADE,
    remote_name TEXT NOT NULL REFERENCES remote (name) ON DELETE CASCADE,
    id_on_remote TEXT DEFAULT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, remote_name, id_on_remote)
);

------ ALBUMS
CREATE TABLE IF NOT EXISTS album (id TEXT PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_album (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL REFERENCES album (id) ON DELETE CASCADE,
    remote_name TEXT NOT NULL REFERENCES remote (name) ON DELETE CASCADE,
    id_on_remote TEXT DEFAULT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, remote_name, id_on_remote)
);

------ TRACKS
CREATE TABLE IF NOT EXISTS track (id TEXT PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_track (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL REFERENCES track (id) ON DELETE CASCADE,
    remote_name TEXT NOT NULL REFERENCES remote (name) ON DELETE CASCADE,
    id_on_remote TEXT DEFAULT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, remote_name, id_on_remote)
);

------ PLAYLISTS
CREATE TABLE IF NOT EXISTS playlist (id TEXT PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_playlist (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL REFERENCES playlist (id) ON DELETE CASCADE,
    remote_name TEXT NOT NULL REFERENCES account (id) ON DELETE CASCADE,
    id_on_remote TEXT NOT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, remote_name, id_on_remote)
);