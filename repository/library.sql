PRAGMA foreign_keys = ON;

------ REMOTE
CREATE TABLE IF NOT EXISTS service (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS account (
    id INTEGER PRIMARY KEY,
    service_name INTEGER NOT NULL REFERENCES service (name) ON DELETE CASCADE,
    alias TEXT NOT NULL DEFAULT 'Without alias',
    auth TEXT NOT NULL,
    added_at INTEGER NOT NULL
);

------ ARTISTS
CREATE TABLE IF NOT EXISTS artist (id INTEGER PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_artist (
    id INTEGER PRIMARY KEY,
    entity_id INTEGER NOT NULL REFERENCES artist (id) ON DELETE CASCADE,
    service_name INTEGER NOT NULL REFERENCES service (name) ON DELETE CASCADE,
    id_on_service TEXT DEFAULT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, service_name, id_on_service)
);

------ ALBUMS
CREATE TABLE IF NOT EXISTS album (id INTEGER PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_album (
    id INTEGER PRIMARY KEY,
    entity_id INTEGER NOT NULL REFERENCES album (id) ON DELETE CASCADE,
    service_name INTEGER NOT NULL REFERENCES service (name) ON DELETE CASCADE,
    id_on_service TEXT DEFAULT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, service_name, id_on_service)
);

------ TRACKS
CREATE TABLE IF NOT EXISTS track (id INTEGER PRIMARY KEY);

CREATE TABLE IF NOT EXISTS linked_track (
    id INTEGER PRIMARY KEY,
    entity_id INTEGER NOT NULL REFERENCES track (id) ON DELETE CASCADE,
    service_name INTEGER NOT NULL REFERENCES service (name) ON DELETE CASCADE,
    id_on_service TEXT DEFAULT NULL,
    modified_at INTEGER NOT NULL DEFAULT 0,
    UNIQUE (entity_id, service_name, id_on_service)
);